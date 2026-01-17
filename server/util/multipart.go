package util

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/indieinfra/scribble/server/resp"
)

// ParseMultipartWithFile parses a multipart/form-data request with a single expected file field.
// It delegates to ParseMultipartWithFirstFile to avoid duplication. If requireFile is true, missing
// files return an invalid_request.
func ParseMultipartWithFile(w http.ResponseWriter, r *http.Request, maxSize int64, fileField string, requireFile bool) (map[string]any, multipart.File, *multipart.FileHeader, bool) {
	values, file, header, _, ok := ParseMultipartWithFirstFile(w, r, maxSize, []string{fileField}, requireFile)
	return values, file, header, ok
}

// ParseMultipartWithFirstFile parses a multipart/form-data request, returning the first matching
// file for the provided field order. The name of the matched field is returned so callers can map
// to Micropub properties. If requireFile is true, missing files write an invalid_request.
func ParseMultipartWithFirstFile(w http.ResponseWriter, r *http.Request, maxSize int64, fileFields []string, requireFile bool) (map[string]any, multipart.File, *multipart.FileHeader, string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		WriteInvalidMultipart(w, err)
		return nil, nil, nil, "", false
	}

	values := make(map[string]any)
	if r.MultipartForm != nil {
		for key, arr := range r.MultipartForm.Value {
			switch len(arr) {
			case 0:
				continue
			case 1:
				values[key] = arr[0]
			default:
				asAny := make([]any, len(arr))
				for i, v := range arr {
					asAny[i] = v
				}
				values[key] = asAny
			}
		}
	}

	var selectedFile multipart.File
	var selectedHeader *multipart.FileHeader
	selectedField := ""
	fileCount := 0

	for _, field := range fileFields {
		files := r.MultipartForm.File[field]
		if len(files) == 0 {
			continue
		}

		fileCount += len(files)
		if fileCount > 1 {
			WriteInvalidMultipart(w, fmt.Errorf("only one file is allowed per request"))
			return nil, nil, nil, "", false
		}

		if len(files[0].Filename) == 0 {
			WriteInvalidMultipart(w, fmt.Errorf("file part is missing filename"))
			return nil, nil, nil, "", false
		}

		f, err := files[0].Open()
		if err != nil {
			WriteInvalidMultipart(w, err)
			return nil, nil, nil, "", false
		}

		selectedFile = f
		selectedHeader = files[0]
		selectedField = field
	}

	if fileCount == 0 && requireFile {
		WriteInvalidMultipart(w, fmt.Errorf("missing required file field"))
		return nil, nil, nil, "", false
	}

	return values, selectedFile, selectedHeader, selectedField, true
}

// WriteInvalidMultipart writes a standardized invalid_request for multipart parsing issues.
func WriteInvalidMultipart(w http.ResponseWriter, err error) {
	resp.WriteInvalidRequest(w, fmt.Sprintf("failed to read multipart form: %v", err))
}
