//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Standard golang
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	s3c "github.com/nexocrew/3nigm4/lib/s3"
)

// generateTranscationId generate a randomised tx id to
// avoid conflicts caused, for example, by multiple accesses
// to the same file in a short time range.
func generateTranscationId(id, user string, t *time.Time) string {
	var raw []byte
	randoms, _ := ct.RandomBytesForLen(16)
	raw = append(raw, randoms...)
	composed := fmt.Sprintf("%s.%s.%d.%d",
		id,
		user,
		t.Unix(),
		t.UnixNano())
	raw = append(raw, []byte(composed)...)
	checksum := sha256.Sum256(raw)
	return hex.EncodeToString(checksum[:])
}

// updateUploadRequestStatus update the tx status when a
// status update is returned from the working queue to the
// UploadedChan chan (s3 client).
func updateUploadRequestStatus(ur s3c.OpResult) {
	session := db.Copy()
	defer session.Close()

	at, err := session.GetAsyncTx(ur.RequestID)
	if err != nil {
		log.ErrorLog("Retrieving tx async doc %s produced error %s, ignoring.\n", ur.RequestID, err.Error())
		return
	}
	fl, err := session.GetFileLog(ur.ID)
	if err != nil {
		log.ErrorLog("Retrieving log doc %s produced error %s, ignoring.\n", ur.ID, err.Error())
		return
	}

	// update status
	at.Complete = true
	fl.Complete = true
	at.Error = ur.Error

	// update in the db
	err = session.UpdateAsyncTx(at)
	if err != nil {
		log.ErrorLog("Unable to update %s tx async doc cause %s, ignoring.\n", at.Id, err.Error())
		return
	}
	err = session.UpdateFileLog(fl)
	if err != nil {
		log.ErrorLog("Unable to update %s log doc cause %s, ignoring.\n", at.Id, err.Error())
		return
	}
}

// updateDownloadRequestStatus manage workingqueue messages from
// completed S3 download operations.
func updateDownloadRequestStatus(dr s3c.OpResult) {
	session := db.Copy()
	defer session.Close()

	at, err := session.GetAsyncTx(dr.RequestID)
	if err != nil {
		log.ErrorLog("Retrieving tx async doc %s produced error %s, ignoring.\n", dr.RequestID, err.Error())
		return
	}

	// update status
	at.Complete = true
	at.Error = dr.Error
	at.Data = dr.Data
	hash := sha256.Sum256(at.Data)
	at.CheckSum = ct.CheckSum{
		Hash: hash[:],
		Type: "SHA256",
	}

	// update in the db
	err = session.UpdateAsyncTx(at)
	if err != nil {
		log.ErrorLog("Unable to update %s tx async doc cause %s, ignoring.\n", at.Id, err.Error())
		return
	}
}

// updateDeleteRequestStatus update status related to an async
// delete operation.
func updateDeleteRequestStatus(dr s3c.OpResult) {
	session := db.Copy()
	defer session.Close()

	at, err := session.GetAsyncTx(dr.RequestID)
	if err != nil {
		log.ErrorLog("Retrieving tx async doc %s produced error %s, ignoring.\n", dr.RequestID, err.Error())
		return
	}

	// delete file record in db
	err = session.RemoveFileLog(dr.ID)
	if err != nil {
		log.ErrorLog("Unable to remove required %s doc from the database cause %s, continuing", dr.ID, err.Error())
		// this error will not block the operation (cause the file on S3 has been already deleted).
	}

	// update status
	at.Complete = true
	at.Error = dr.Error

	// update in the db
	err = session.UpdateAsyncTx(at)
	if err != nil {
		log.ErrorLog("Unable to update %s tx async doc cause %s, ignoring.\n", at.Id, err.Error())
		return
	}
}

// manageAsyncError handles error returned by S3 workers
// nothing special can be done, at this moment, apart from
// logging it.
// TODO: try to define better recovery or alert strategies.
func manageAsyncError(err error) {
	log.ErrorLog("Error while uploading with S3 working queue: %s.\n", err.Error())
}

// manageS3chans manages chan messages from working queue
// async S3 upload/download.
func manageS3chans(s3backend *s3c.Session) {
	var errcClosed, uploadedcClosed, downloadedcClosed, deletedcClosed bool
	for {
		if errcClosed == true {
			log.CriticalLog("S3 error chan is closed, unable to proceed managing chan queue.\n")
			return
		}
		if uploadedcClosed == true {
			log.CriticalLog("S3 upload chan is closed, unable to proceed managing chan queue.\n")
			return
		}
		if downloadedcClosed == true {
			log.CriticalLog("S3 download chan is closed, unable to proceed managing chan queue.\n")
			return
		}
		if deletedcClosed == true {
			log.CriticalLog("S3 delete chan is closed, unable to proceed managing chan queue.\n")
			return
		}
		// select on channels
		select {
		case err, errcOk := <-s3backend.ErrorChan:
			if !errcOk {
				errcClosed = true
			} else {
				go manageAsyncError(err)
			}
		case uploaded, uploadedcOk := <-s3backend.UploadedChan:
			if !uploadedcOk {
				uploadedcClosed = true
			} else {
				go updateUploadRequestStatus(uploaded)
			}
		case downloaded, downloadedcOk := <-s3backend.DownloadedChan:
			if !downloadedcOk {
				downloadedcClosed = true
			} else {
				go updateDownloadRequestStatus(downloaded)
			}
		case deleted, deletedcOk := <-s3backend.DeletedChan:
			if !deletedcOk {
				deletedcClosed = true
			} else {
				go updateDeleteRequestStatus(deleted)
			}
		}
	}
}
