package main

import "os"
import "log"

// LocalStorage manages files saved on disk
// All files are saved into dir directory
type LocalStorage struct {
}

const dir = "./tmp/"

// CreateTempFile creates temporary file in dir directory
func (s *LocalStorage) CreateTempFile(name string) (*os.File, error) {
	file, err := os.OpenFile(dir+name,
		os.O_CREATE|os.O_WRONLY,
		0600)
	if err != nil {
		log.Println("In CreateTempFile: ", err)
		return nil, err
	}
	return file, nil
}
