//
// 3nigm4 storageservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Internal libs
import (
	"github.com/nexocrew/3nigm4/lib/auth"
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// createStorageResource upload a data chunk to the S3 backend service
// after authorising the user. It operates in async mode to perform the
// actual upload using a working queue to integrate S3 backend.
func createStorageResource(w http.ResponseWriter, r *http.Request, args *ct.JobPostRequest, userInfo *auth.UserInfoResponseArg) {
	if args.Arguments.Data == nil ||
		len(args.Arguments.Data) == 0 {
		riseError(http.StatusBadRequest,
			"data in request body is nil", w,
			r.RemoteAddr)
		return
	}

	// get time stamp
	now := time.Now()
	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()
	// insert file log in the database
	checksum := sha256.Sum256(args.Arguments.Data)
	fl := &FileLog{
		Id:         args.Arguments.ResourceID,
		Size:       len(args.Arguments.Data),
		Bucket:     arguments.s3Bucket,
		Creation:   now,
		TimeToLive: args.Arguments.TimeToLive,
		CheckSum: ct.CheckSum{
			Hash: checksum[:],
			Type: "SHA256",
		},
		Ownership: Owner{
			Username:  userInfo.Username,
			OriginIp:  r.RemoteAddr,
			UserAgent: r.UserAgent(),
		},
		Acl: Acl{
			Permission:   args.Arguments.Permission,
			SharingUsers: args.Arguments.SharingUsers,
		},
	}
	err := dbSession.SetFileLog(fl)
	if err != nil {
		riseError(http.StatusInternalServerError,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// generate tx id
	jobId := generateTranscationId(fl.Id, userInfo.Username, &now)

	// add async tx record
	err = dbSession.SetAsyncTx(&AsyncTx{
		Id:        jobId,
		Complete:  false,
		TimeStamp: fl.Creation,
		Ownership: fl.Ownership,
	})
	if err != nil {
		riseError(http.StatusInternalServerError,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// upload data to the S3 backend
	var expireTime *time.Time
	if fl.TimeToLive != 0 {
		ttl := fl.Creation.Add(fl.TimeToLive)
		expireTime = &ttl
	}
	s3backend.Upload(fl.Bucket, fl.Id, jobId, args.Arguments.Data, expireTime)

	// return upload response message
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(
		&ct.JobPostResponse{
			JobID: jobId,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Upload request %s accepted, waiting for upload verification", jobId)
	}
}

// checkAclPermission verify all possible acl scenarios and check if the
// requiring user has required permissions to access the file. If user can
// download it it'll return true otherwise false.
func checkAclPermission(userInfo *auth.UserInfoResponseArg, fileLog *FileLog) bool {
	// check access credentials
	switch fileLog.Acl.Permission {
	case Private:
		if fileLog.Ownership.Username == userInfo.Username {
			return true
		}
	case Public:
		return true
	case Shared:
		// the owner can always have access to the file even
		// if it's not on the sharing users list.
		if fileLog.Ownership.Username == userInfo.Username {
			return true
		}
		for _, permitted := range fileLog.Acl.SharingUsers {
			if permitted == userInfo.Username {
				return true
			}
		}
	}
	return false
}

// retrieveStorageResource implements the first step of a file download request
// it is exposed via a REST GET method and returns a txId usable with the verify
// API call toretrieve the actual downloaded data (from S3 storage). The user
// must be correctly authenticated to be able to access the requested resource.
func retrieveStorageResource(w http.ResponseWriter, r *http.Request, args *ct.JobPostRequest, userInfo *auth.UserInfoResponseArg) {
	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()
	// get resources info
	fileLog, err := dbSession.GetFileLog(args.Arguments.ResourceID)
	if err != nil {
		riseError(http.StatusNotFound,
			fmt.Sprintf("requested file not found"), w,
			r.RemoteAddr)
		return
	}

	// check permission
	granted := checkAclPermission(userInfo, fileLog)
	if !granted {
		riseError(http.StatusUnauthorized,
			fmt.Sprintf("you are not authorised to access this resource"), w,
			r.RemoteAddr)
		return
	}

	now := time.Now()
	// generate tx id
	jobId := generateTranscationId(args.Arguments.ResourceID, userInfo.Username, &now)
	// add async tx record
	err = dbSession.SetAsyncTx(&AsyncTx{
		Id:        jobId,
		Complete:  false,
		TimeStamp: now,
		Ownership: Owner{
			Username:  userInfo.Username,
			OriginIp:  r.RemoteAddr,
			UserAgent: r.UserAgent(),
		},
	})
	if err != nil {
		riseError(http.StatusInternalServerError,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// require S3 download
	s3backend.Download(arguments.s3Bucket, args.Arguments.ResourceID, jobId)

	// return download response message
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(
		&ct.JobPostResponse{
			JobID: jobId,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Download request %s accepted, waiting for download verification", jobId)
	}
}

// deleteStorageResource remove a file from the S3 storage: only the original file
// owner (who uploaded it) can remove a file from there.
func deleteStorageResource(w http.ResponseWriter, r *http.Request, args *ct.JobPostRequest, userInfo *auth.UserInfoResponseArg) {
	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()
	// get resources info
	fileLog, err := dbSession.GetFileLog(args.Arguments.ResourceID)
	if err != nil {
		riseError(http.StatusNotFound,
			fmt.Sprintf("requested file not found"), w,
			r.RemoteAddr)
		return
	}

	// check permission
	var granted bool
	// strict acl verification: only the file owner is able
	// to delete it.
	if fileLog.Ownership.Username == userInfo.Username {
		granted = true
	}
	if !granted {
		riseError(http.StatusUnauthorized,
			fmt.Sprintf("you are not authorised to delete this resource"), w,
			r.RemoteAddr)
		return
	}

	now := time.Now()
	// generate tx id
	jobId := generateTranscationId(args.Arguments.ResourceID, userInfo.Username, &now)
	// add async tx record
	err = dbSession.SetAsyncTx(&AsyncTx{
		Id:        jobId,
		Complete:  false,
		TimeStamp: now,
		Ownership: Owner{
			Username:  userInfo.Username,
			OriginIp:  r.RemoteAddr,
			UserAgent: r.UserAgent(),
		},
	})
	if err != nil {
		riseError(http.StatusInternalServerError,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// require S3 download
	s3backend.Delete(arguments.s3Bucket, args.Arguments.ResourceID, jobId)

	// return download response message
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(
		&ct.JobPostResponse{
			JobID: jobId,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Delete request %s accepted, waiting for delete verification", jobId)
	}
}
