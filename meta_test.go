package main

import "testing"
import "net/http"

func TestParseMeta(t *testing.T) {
	req, _ := http.NewRequest("POST", "/files", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----Zam1WUeLK7vBj4wN")
	req.Header.Set("Content-Range", "bytes 512000-1023999/1141216")
	req.Header.Set("Content-Disposition", `attachment; filename="picture.jpg"`)

	meta, err := ParseMeta(req)
	if err != nil {
		t.Error("Expected nil, got ", err)
	}

	if meta.MediaType != "multipart/form-data" {
		t.Error("Expected multipart/form-data, got ", meta.MediaType)
	}

	if meta.Boundary != "----Zam1WUeLK7vBj4wN" {
		t.Error("Expected ----Zam1WUeLK7vBj4wN, got ", meta.Boundary)
	}

	if meta.Range.Start != 512000 {
		t.Error("Expected 512000, got ", meta.Range.Start)
	}

	if meta.Range.End != 1023999 {
		t.Error("Expected 1023999, got ", meta.Range.End)
	}

	if meta.Range.Size != 1141216 {
		t.Error("Expected 1141216, got ", meta.Range.Start)
	}

	if meta.FileName != "picture.jpg" {
		t.Error("Expected picture.jpg, got ", meta.FileName)
	}
}
