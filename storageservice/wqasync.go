//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Internal dependencies
import (
	s3c "github.com/nexocrew/3nigm4/lib/s3"
)

// updateUploadRequestStatus update the tx status when a
// status update is returned from the working queue to the
// UploadedChan chan (s3 client).
func updateUploadRequestStatus(ur s3c.UploadRequest) {
	session := db.Copy()
	defer session.Close()

	at, err := session.GetAsyncTx(ur.ID)
	if err != nil {
		log.ErrorLog("Retrieving tx async doc %s produced error %s, ignoring.\n", ur.ID, err.Error())
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
// TODO: implement the update status logic.
func updateDownloadRequestStatus(dr s3c.DownloadRequest) {

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
	var errc_closed, uploadedc_closed, downloadedc_closed bool
	for {
		if errc_closed == true {
			log.CriticalLog("S3 error chan is closed, unable to proceed managing chan queue.\n")
			return
		}
		if uploadedc_closed == true {
			log.CriticalLog("S3 upload chan is closed, unable to proceed managing chan queue.\n")
			return
		}
		if downloadedc_closed == true {
			log.CriticalLog("S3 download chan is closed, unable to proceed managing chan queue.\n")
			return
		}
		// select on channels
		select {
		case err, errc_ok := <-s3backend.ErrorChan:
			if !errc_ok {
				errc_closed = true
			} else {
				go manageAsyncError(err)
			}
		case uploaded, uploadedc_ok := <-s3backend.UploadedChan:
			if !uploadedc_ok {
				uploadedc_closed = true
			} else {
				go updateUploadRequestStatus(uploaded)
			}
		case downloaded, downloadedc_ok := <-s3backend.DownloadedChan:
			if !downloadedc_ok {
				downloadedc_closed = true
			} else {
				go updateDownloadRequestStatus(downloaded)
			}
		}
	}
}
