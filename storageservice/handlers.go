//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"encoding/json"
	"net/http"
)

func postChunk(w http.ResponseWriter, r *http.Request) {

}

func getChunk(w http.ResponseWriter, r *http.Request) {

}

func deleteChunk(w http.ResponseWriter, r *http.Request) {

}

// Ping function to verify if the service is on
// or not.
func getPing(w http.ResponseWriter, r *http.Request) {
	response := StandardResponse{}
	response.Status = AckResponse

	/* return value */
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
