package main

import (
	"log"
	"net/http"

	"github.com/aeshes/storage/api"
	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/testing", api.TestingEndpoint).Methods("PUT")
	router.HandleFunc("/testing", api.GetFileEndpoint).Methods("GET")
	router.HandleFunc("/commonfs/{fileid}", api.HandleFileID).Methods("HEAD")
	router.HandleFunc("/commonfs/createAtomically", api.HandleCreateAtomically).Methods("PUT")

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal(err)
	}
}
