package post

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/server/resp"
	"github.com/indieinfra/scribble/server/util"
)

type ParsedBody struct {
	Data        map[string]any
	File        *ParsedFile
	AccessToken string
}

type ParsedFile struct {
	File   multipart.File
	Header *multipart.FileHeader
	Field  string
}

func ReadBody(cfg *config.Config, w http.ResponseWriter, r *http.Request) (*ParsedBody, bool) {
	_, contentType, ok := util.RequireValidMicropubContentType(w, r)
	if !ok {
		return nil, false
	}

	switch contentType {
	case "application/json":
		return &ParsedBody{Data: readJsonBody(cfg, w, r)}, true
	case "application/x-www-form-urlencoded":
		body := readFormUrlEncodedBody(cfg, w, r)
		token := util.PopAccessToken(body)
		return &ParsedBody{Data: body, AccessToken: token}, true
	case "multipart/form-data":
		return readMultipartBody(cfg, w, r)
	}

	return nil, false
}

func readJsonBody(cfg *config.Config, w http.ResponseWriter, r *http.Request) map[string]any {
	out := make(map[string]any)

	r.Body = http.MaxBytesReader(w, r.Body, int64(cfg.Server.Limits.MaxPayloadSize))
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		resp.WriteInvalidRequest(w, "Invalid JSON body")
		return nil
	}

	return out
}

func readFormUrlEncodedBody(cfg *config.Config, w http.ResponseWriter, r *http.Request) map[string]any {
	out := make(map[string]any)

	r.Body = http.MaxBytesReader(w, r.Body, int64(cfg.Server.Limits.MaxPayloadSize))
	if err := r.ParseForm(); err != nil {
		resp.WriteInvalidRequest(w, fmt.Sprintf("Invalid form body: %v", err))
		return nil
	}

	for key, values := range r.Form {
		switch len(values) {
		case 0:
			continue
		case 1:
			out[key] = values[0]
		default:
			arr := make([]any, len(values))
			for i, v := range values {
				arr[i] = v
			}
			out[key] = arr
		}
	}

	return out
}

func readMultipartBody(cfg *config.Config, w http.ResponseWriter, r *http.Request) (*ParsedBody, bool) {
	maxSize := int64(cfg.Server.Limits.MaxFileSize)
	fields := []string{"photo", "video", "audio", "file"}
	values, file, header, field, ok := util.ParseMultipartWithFirstFile(w, r, maxSize, fields, false)
	if !ok {
		return nil, false
	}

	token := util.PopAccessToken(values)

	var parsedFile *ParsedFile
	if file != nil && header != nil {
		parsedFile = &ParsedFile{File: file, Header: header, Field: field}
	}

	return &ParsedBody{Data: values, File: parsedFile, AccessToken: token}, true
}
