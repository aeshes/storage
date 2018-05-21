package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var storage = DataStorage{}
var local = LocalStorage{}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func testingEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	meta, err := ParseMeta(r)
	check(err)

	fmt.Println(meta.FileName)

	file, err := createTempFile("temp.jpg")
	check(err)
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		log.Fatal(err)
	}
}

func createTempFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0600)
	check(err)

	return file, nil
}

func getFileEndpoint(w http.ResponseWriter, r *http.Request) {
	file, err := storage.OpenFile("hello")
	if err != nil {
		respondWithError(w, 404, "grid file not found")
		return
	}
	defer file.Close()
	if _, err := io.Copy(w, file); err != nil {
		respondWithError(w,
			http.StatusInternalServerError,
			"error while reading grid file")
	}
}

// HEAD request
/*
/commonfs/<fileid>
Request headers:
none

Response status:
200 - presents, 404 - absents

Response headers:
<name> : file name
<hash> : file hash
<creator> : creator property
<sysId>: sysId property

Response body:
none */

func handleFileID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileid := vars["fileid"]
	fmt.Println(fileid)
	meta := storage.QueryMeta(fileid)
	if meta != nil {
		w.Header().Add("name", meta.Name)
		w.Header().Add("hash", meta.Hash)
		w.Header().Add("creator", meta.Creator)
		w.Header().Add("sysId", meta.SysID)
		w.WriteHeader(http.StatusOK)

		return
	}

	w.WriteHeader(404)
}

// PUT request
//Atomically creates "small" file which can be POSTed in one request
func handleCreateAtomically(w http.ResponseWriter, r *http.Request) {
	meta, err := ParseMeta(r)
	if err != nil {
		log.Println("In handleCreateAtomically: ", err)
		respondWithError(w, 400, "cant parse meta info")
		return
	}

	if meta.Property.isValid() {

		meta.Property.Dump()

		file, err := local.CreateTempFile(meta.Property.Name)
		if err != nil {
			log.Println("When CreateAtomically: create temporary, ", err)
			respondWithError(w, 400, "cant create local file")
			return
		}

		defer file.Close()

		if _, err := io.Copy(file, r.Body); err != nil {
			log.Println("When CreateAtomically, write file: ", err)
			respondWithError(w, 400, "cant write local file")
			return
		}

		// Note: why file handle is invalid in StoreFromDisk?
		fileOnDisk := &LocalFile{Path: "./tmp/" + meta.Property.Name,
			Hash:   meta.Property.Hash,
			Handle: file,
			Prev:   nil}
		storage.StoreFromDisk(fileOnDisk, meta.Property)
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w,
		code,
		map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	storage.Connect()
	storage.SaveFileToDisk("meow.jpg")

	router := mux.NewRouter()
	router.HandleFunc("/testing", testingEndpoint).Methods("PUT")
	router.HandleFunc("/testing", getFileEndpoint).Methods("GET")
	router.HandleFunc("/commonfs/{fileid}", handleFileID).Methods("HEAD")
	router.HandleFunc("/commonfs/createAtomically", handleCreateAtomically).Methods("PUT")

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal(err)
	}
}
