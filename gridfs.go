package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// DataStorage is responsible for all database interaactions
type DataStorage struct {
	session *mgo.Session
	db      *mgo.Database
}

type FileProperties struct {
	Reference bson.ObjectId `bson:"reference" json:"reference"`
	Name      string        `bson:"name" json:"name"`
	Creator   string        `bson:"creator" json:"creator"`
	Hash      string        `bson:"hash" json:"hash"`
	SysID     string        `bson:"sysId" json:"sysId"`
}

// MongoServerAddr is the default server to connect with
const MongoServerAddr = "127.0.0.1"

// MongoDefaultDB is the default database to store files
const MongoDefaultDB = "binary"

// FileMetaCollection is a mongo collection for file meta data:
// reference to grid file, name, creator, hash, sysId
const FileMetaCollection = "meta"

// Connect connects to the default server and opens default DB
func (s *DataStorage) Connect() {
	session, err := mgo.Dial(MongoServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	s.session = session
	s.db = session.DB(MongoDefaultDB)
}

// CreateGridFile creates file in GridFS
func (s *DataStorage) CreateGridFile(name string) (*mgo.GridFile, error) {
	file, err := s.db.GridFS("fs").Create(name)
	if err != nil {
		return nil, errors.New("Can not create Grid file")
	}
	return file, nil
}

// InsertMetaInfo inserta meta information about new file into table 'meta'
func (s *DataStorage) InsertMetaInfo(file *mgo.GridFile, m *FileMeta) {

	// Open meta collection
	meta := s.session.DB(MongoDefaultDB).C(FileMetaCollection)
	var properties = &FileProperties{Reference: file.Id().(bson.ObjectId),
		Name:    m.Name,
		Creator: m.Creator,
		Hash:    m.Hash,
		SysID:   m.SysID}

	// Save new meta info
	if err := meta.Insert(&properties); err != nil {
		log.Println("Cant insert meta information", err)
	}
}

// QueryMeta queries meta information from collection 'meta'
// about file identified by id, which must be a string representation
//  of an ObjectId
func (s *DataStorage) QueryMeta(id string) *FileMeta {
	var reference bson.ObjectId
	var meta FileMeta

	if bson.IsObjectIdHex(id) {
		reference = bson.ObjectIdHex(id)
		c := s.db.C(FileMetaCollection)
		err := c.Find(bson.M{"reference": reference}).One(&meta)
		if err != nil {
			log.Println("While finding meta info by ObjectId: ", err)
			return nil
		}
		return &meta
	}

	return nil
}

// StoreFromDisk stores disk file in GridFS
// If local file's sha-256 is not equal to sha-256 value in FileMeta,
// error is returned
func (s *DataStorage) StoreFromDisk(file *LocalFile, meta *FileMeta) error {
	if file.Sha256() == meta.Hash {
		gridFile, err := s.CreateGridFile(meta.Name)
		if err != nil {
			log.Println("In StoreFromDisk: ", err)
			return err
		}
		defer gridFile.Close()

		content, err := ioutil.ReadFile(file.Path)
		if err != nil {
			log.Println("While reading local file in StoreFromDisk: ", err)
			return err
		}

		bytesWritten, err := gridFile.Write(content)
		if err != nil {
			log.Println("While writing local file to GridFS: ", err)
			return err
		}
		s.InsertMetaInfo(gridFile, meta)
		log.Printf("Copied %d bytes to GridFS.", bytesWritten)
	}

	return errors.New("file.sha256 != meta.sha256")
}

// OpenFile opens grid file for reading
func (s *DataStorage) OpenFile(name string) (io.ReadCloser, error) {
	file, err := s.db.GridFS("fs").Open(name)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return file, nil
}

// SaveFileToDisk reads grid file from GridFS
// and saves read data to disk
func (s *DataStorage) SaveFileToDisk(name string) {
	// opens file in mongo GridFS
	file, err := s.db.GridFS("fs").Open(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	dest, err := os.OpenFile("./tmpfile.jpg",
		os.O_CREATE|os.O_WRONLY,
		0644)
	if err != nil {
		log.Println("While creating file: ", err)
	}
	defer dest.Close()

	// Copies from grid file to disk file
	if _, err := io.Copy(dest, file); err != nil {
		fmt.Println(err)
	}
}
