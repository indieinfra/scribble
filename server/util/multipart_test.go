package util

import (
	"bytes"
	"io"
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

func TestParseMultipartWithFirstFile_Success(t *testing.T) {
	req, _ := makeMultipartRequest(t, map[string]string{"title": "hello"}, map[string]io.Reader{"file": bytes.NewBufferString("data")})
	rr := httptest.NewRecorder()

	values, file, header, field, ok := ParseMultipartWithFirstFile(rr, req, 1_000_000, 0, []string{"file"}, true)
	if !ok {
		t.Fatalf("expected parse to succeed")
	}
	if field != "file" || header == nil {
		t.Fatalf("expected matching file field")
	}
	if values["title"].(string) != "hello" {
		t.Fatalf("expected title to be parsed")
	}
	data, _ := io.ReadAll(file)
	if string(data) != "data" {
		t.Fatalf("unexpected file contents %q", data)
	}
	file.Close()
}

func TestParseMultipartWithFirstFile_MissingRequired(t *testing.T) {
	req, _ := makeMultipartRequest(t, map[string]string{"title": "hello"}, nil)
	rr := httptest.NewRecorder()

	if _, _, _, _, ok := ParseMultipartWithFirstFile(rr, req, 1_000_000, 0, []string{"file"}, true); ok {
		t.Fatalf("expected missing file to fail")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestParseMultipartWithFile_Single(t *testing.T) {
	req, _ := makeMultipartRequest(t, map[string]string{"title": "hello"}, map[string]io.Reader{"file": bytes.NewBufferString("data")})
	rr := httptest.NewRecorder()

	values, file, header, ok := ParseMultipartWithFile(rr, req, 1_000_000, 0, "file", true)
	if !ok {
		t.Fatalf("expected success")
	}
	if header == nil || values["title"] == nil {
		t.Fatalf("expected file and values parsed")
	}
	file.Close()
}

func makeMultipartRequest(t *testing.T, fields map[string]string, files map[string]io.Reader) (*http.Request, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	for name, r := range files {
		fw, err := w.CreateFormFile(name, name+".txt")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := io.Copy(fw, r); err != nil {
			t.Fatalf("copy file: %v", err)
		}
	}
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, w.Boundary()
}
