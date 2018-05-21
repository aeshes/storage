package main

import "net/http"
import "mime"
import "errors"
import "fmt"
import "log"

// Meta describes HTTP metainfo
type Meta struct {
	MediaType string
	Boundary  string
	Range     *Range
	FileName  string
	Property  *FileMeta
}

// Range describes HTTP range
type Range struct {
	Start int64
	End   int64
	Size  int64
}

// FileMeta describes user-define file info
type FileMeta struct {
	Name    string `bson:"name" json:"name"`
	Hash    string `bson:"hash" json:"hash"`
	Creator string `bson:"creator" json:"creator"`
	SysID   string `bson:"sysId" json:"sysId"`
}

// ParseMeta parses request information and makes Meta
func ParseMeta(req *http.Request) (*Meta, error) {
	meta := &Meta{}

	if err := meta.parseContentType(req.Header.Get("Content-Type")); err != nil {
		return nil, err
	}

	if err := meta.parseContentRange(req.Header.Get("Content-Range")); err != nil {
		return nil, err
	}

	if err := meta.parseContentDisposition(req.Header.Get("Content-Disposition")); err != nil {
		return nil, err
	}

	// Parse user defined HTTP headers
	meta.parseUserProperties(req)

	return meta, nil
}

func (meta *Meta) parseContentType(ct string) error {
	if ct == "" {
		meta.MediaType = "application/octet-stream"
		return nil
	}

	mediatype, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}

	if mediatype == "multipart/form-data" {
		boundary, ok := params["boundary"]
		if !ok {
			return errors.New("Meta: boundary not defined")
		}

		meta.MediaType = mediatype
		meta.Boundary = boundary
	} else {
		meta.MediaType = "application/octet-stream"
	}

	return nil
}

func (meta *Meta) parseContentRange(cr string) error {
	if cr == "" {
		return nil
	}

	var start, end, size int64

	_, err := fmt.Sscanf(cr, "bytes %d-%d/%d", &start, &end, &size)
	if err != nil {
		return err
	}

	meta.Range = &Range{Start: start, End: end, Size: size}

	return nil
}

func (meta *Meta) parseContentDisposition(cd string) error {
	if cd == "" {
		return nil
	}

	_, params, err := mime.ParseMediaType(cd)
	if err != nil {
		return err
	}

	filename, ok := params["filename"]
	if !ok {
		return errors.New("Meta: file in Content-Disposition not defined")
	}

	meta.FileName = filename

	return nil
}

// Parse user defined headers

func (meta *Meta) parseUserProperties(req *http.Request) {
	var name, hash, creator, id string

	name = parseName(req.Header.Get("name"))
	creator = parseCreator(req.Header.Get("creator"))
	hash = parseHash(req.Header.Get("hash"))
	id = parseSysID(req.Header.Get("sysId"))

	meta.Property = &FileMeta{Name: name,
		Creator: creator,
		Hash:    hash,
		SysID:   id}
}

func parseName(name string) string {
	if name == "" {
		log.Println("Request with empty Name header")
	}

	return name
}

func parseHash(h string) string {
	if h == "" {
		log.Println("Request with empty Hash header")
	}

	return h
}

func parseCreator(creator string) string {
	if creator == "" {
		log.Println("Request with empty Creator header")
	}

	return creator
}

func parseSysID(sysID string) string {
	if sysID == "" {
		log.Println("Request with empty sysId header")
	}

	return sysID
}

func (p *FileMeta) isValid() bool {
	return p.Name != "" &&
		p.Creator != "" &&
		p.Hash != "" &&
		p.SysID != ""
}

func (p *FileMeta) Dump() {
	fmt.Printf("Name = %s\n", p.Name)
	fmt.Printf("Creator = %s\n", p.Creator)
	fmt.Printf("Hash = %s\n", p.Hash)
	fmt.Printf("sysId = %s\n", p.SysID)
}
