package util

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseMultipartFiles_BaseAndArray(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("title", "hello")

	fw1, _ := w.CreateFormFile("photo", "a.jpg")
	_, _ = fw1.Write([]byte("abc"))

	fw2, _ := w.CreateFormFile("photo[]", "b.jpg")
	_, _ = fw2.Write([]byte("def"))

	w.Close()

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	values, files, ok := ParseMultipartFiles(rr, req, 1<<20, 1<<20, []string{"photo"}, true)
	if !ok {
		t.Fatalf("expected ok parsing multipart")
	}

	if got := values["title"]; got != "hello" {
		t.Fatalf("expected title value, got %#v", got)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	for _, f := range files {
		defer f.File.Close()
		if f.Field != "photo" {
			t.Fatalf("expected field name photo, got %q", f.Field)
		}
	}
}

func TestParseMultipartFiles_FileTooLarge(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("photo", "a.jpg")
	_, _ = fw.Write([]byte("0123456789")) // 10 bytes
	w.Close()

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	_, _, ok := ParseMultipartFiles(rr, req, 1<<20, 5, []string{"photo"}, true)
	if ok {
		t.Fatalf("expected failure for oversized file")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 response, got %d", rr.Code)
	}
}

func TestParseMultipartFiles_MissingRequired(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("title", "hello")
	w.Close()

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	_, _, ok := ParseMultipartFiles(rr, req, 1<<20, 1<<20, []string{"photo"}, true)
	if ok {
		t.Fatalf("expected failure when file required")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 response, got %d", rr.Code)
	}
}
