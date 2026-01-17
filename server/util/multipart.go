package util

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/indieinfra/scribble/server/resp"
)

// MultipartFile represents a parsed multipart file part with its field name.
type MultipartFile struct {
	Field  string
	File   multipart.File
	Header *multipart.FileHeader
}

// ParseMultipartWithFile parses a multipart/form-data request with a single expected file field.
// It delegates to ParseMultipartWithFirstFile to avoid duplication. If requireFile is true, missing
// files return an invalid_request.
func ParseMultipartWithFile(w http.ResponseWriter, r *http.Request, maxMemory int64, maxFileSize int64, fileField string, requireFile bool) (map[string]any, multipart.File, *multipart.FileHeader, bool) {
	values, file, header, _, ok := ParseMultipartWithFirstFile(w, r, maxMemory, maxFileSize, []string{fileField}, requireFile)
	return values, file, header, ok
}

// ParseMultipartFiles parses a multipart/form-data request and returns all files for the provided
// field names in the given order. If requireFile is true, missing files write an invalid_request.
// maxMemory caps the total request body (for ParseMultipartForm) while maxFileSize enforces a per-file limit when > 0.
func ParseMultipartFiles(w http.ResponseWriter, r *http.Request, maxMemory int64, maxFileSize int64, fileFields []string, requireFile bool) (map[string]any, []MultipartFile, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		WriteInvalidMultipart(w, err)
		return nil, nil, false
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

	var filesOut []MultipartFile

	for _, field := range fileFields {
		candidates := append([]*multipart.FileHeader{}, r.MultipartForm.File[field]...)
		candidates = append(candidates, r.MultipartForm.File[field+"[]"]...)

		for _, fh := range candidates {
			if maxFileSize > 0 && fh.Size > maxFileSize {
				WriteInvalidMultipart(w, fmt.Errorf("file exceeds max size"))
				return nil, nil, false
			}

			if fh.Filename == "" {
				WriteInvalidMultipart(w, fmt.Errorf("file part is missing filename"))
				return nil, nil, false
			}

			f, err := fh.Open()
			if err != nil {
				WriteInvalidMultipart(w, err)
				return nil, nil, false
			}

			filesOut = append(filesOut, MultipartFile{Field: field, File: f, Header: fh})
		}
	}

	if len(filesOut) == 0 && requireFile {
		WriteInvalidMultipart(w, fmt.Errorf("missing required file field"))
		return nil, nil, false
	}

	return values, filesOut, true
}

// ParseMultipartWithFirstFile parses a multipart/form-data request, returning the first matching
// file for the provided field order. The name of the matched field is returned so callers can map
// to Micropub properties. If requireFile is true, missing files write an invalid_request.
func ParseMultipartWithFirstFile(w http.ResponseWriter, r *http.Request, maxMemory int64, maxFileSize int64, fileFields []string, requireFile bool) (map[string]any, multipart.File, *multipart.FileHeader, string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)
	if err := r.ParseMultipartForm(maxMemory); err != nil {
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
		candidates := append([]*multipart.FileHeader{}, r.MultipartForm.File[field]...)
		candidates = append(candidates, r.MultipartForm.File[field+"[]"]...)
		if len(candidates) == 0 {
			continue
		}

		fileCount += len(candidates)
		if fileCount > 1 {
			WriteInvalidMultipart(w, fmt.Errorf("only one file is allowed per request"))
			return nil, nil, nil, "", false
		}

		fh := candidates[0]
		if maxFileSize > 0 && fh.Size > maxFileSize {
			WriteInvalidMultipart(w, fmt.Errorf("file exceeds max size"))
			return nil, nil, nil, "", false
		}

		if len(fh.Filename) == 0 {
			WriteInvalidMultipart(w, fmt.Errorf("file part is missing filename"))
			return nil, nil, nil, "", false
		}

		f, err := fh.Open()
		if err != nil {
			WriteInvalidMultipart(w, err)
			return nil, nil, nil, "", false
		}

		selectedFile = f
		selectedHeader = fh
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
