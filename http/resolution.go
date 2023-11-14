package http

import (
	"fmt"
	go_http "net/http"
	"os"
	"time"

	"github.com/pjsaksa/go-utils/log"
)

type Resolution interface {
	WriteResponse(go_http.ResponseWriter, *go_http.Request)
	LogMessage() log.Message
}

// ------------------------------------------------------------

type ContentResolution struct {
	ContentType string
	Content     []byte
}

func (res *ContentResolution) WriteResponse(out go_http.ResponseWriter, req *go_http.Request) {
	if len(res.ContentType) > 0 {
		out.Header().Set("Content-Type", res.ContentType)
	}
	out.Header().Set("Content-Length", fmt.Sprintf("%d", len(res.Content)))
	if len(res.Content) > 0 {
		out.Write(res.Content)
	} else {
		out.WriteHeader(go_http.StatusNoContent)
	}
}

func (res *ContentResolution) LogMessage() log.Message {
	return log.DebugMsg("%d bytes", len(res.Content))
}

// ------------------------------------------------------------

type FileResolution struct {
	fileName    string
	contentType string
	maxAge      int
	modTime     time.Time
	content     *os.File
}

func ServeFile(req *go_http.Request, fileName string, contentType string, maxAge int) Resolution {
	if req.Method != "GET" {
		return &MethodNotAllowedResolution{Allowed: "GET"}
	}
	if len(fileName) == 0 {
		return &ErrorResolution{Status: go_http.StatusNotFound}
	}

	var info os.FileInfo
	var err error
	info, err = os.Stat(fileName)
	switch {
	case err != nil:
		return &ErrorResolution{Status: go_http.StatusNotFound}
	case (info.Mode() & os.ModeType) != 0:
		return &ErrorResolution{Status: go_http.StatusInternalServerError}
	}

	var content *os.File
	content, err = os.Open(fileName)
	switch {
	case err != nil:
		return &ErrorResolution{
			Status:  go_http.StatusInternalServerError,
			Message: fmt.Sprintf("http.ServeFile: os.Open: ", err.Error()),
		}
	}

	return &FileResolution{
		fileName:    fileName,
		contentType: contentType,
		maxAge:      maxAge,
		modTime:     info.ModTime(),
		content:     content,
	}
}

func (res *FileResolution) WriteResponse(out go_http.ResponseWriter, req *go_http.Request) {
	if len(res.contentType) > 0 {
		out.Header().Set("Content-Type", res.contentType)
	}
	if res.maxAge > 0 {
		out.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", res.maxAge))
	}

	go_http.ServeContent(out, req, res.fileName, res.modTime, res.content)
}

func (res *FileResolution) LogMessage() log.Message {
	return log.DebugMsg(`File "%s"`, res.fileName)
}

// ------------------------------------------------------------

type RedirectResolution struct {
	Status int
	Url    string
}

func (res *RedirectResolution) WriteResponse(out go_http.ResponseWriter, req *go_http.Request) {
	go_http.Redirect(out, req, res.Url, res.Status)
}

func (res *RedirectResolution) LogMessage() log.Message {
	return log.DebugMsg(`Redirect "%s"`, res.Url)
}

// ------------------------------------------------------------

type ErrorResolution struct {
	Status  int
	Message string
}

func (res *ErrorResolution) WriteResponse(out go_http.ResponseWriter, req *go_http.Request) {
	var msg string
	if len(res.Message) == 0 {
		msg = go_http.StatusText(res.Status)
	} else {
		msg = res.Message
	}

	go_http.Error(out, msg, res.Status)
}

func (res *ErrorResolution) LogMessage() log.Message {
	var msg string
	if len(res.Message) > 0 {
		msg = res.Message
	} else {
		msg = go_http.StatusText(res.Status)
	}

	switch {
	case res.Status >= 400 && res.Status < 500:
		return log.WarningMsg("%d %s", res.Status, msg)
	default:
		return log.ErrorMsg("%d %s", res.Status, msg)
	}
}

// ------------------------------------------------------------

type MethodNotAllowedResolution struct {
	Allowed string
}

func (res *MethodNotAllowedResolution) WriteResponse(out go_http.ResponseWriter, req *go_http.Request) {
	if len(res.Allowed) > 0 {
		out.Header().Add("Allow", res.Allowed)
	}
	(&ErrorResolution{Status: go_http.StatusMethodNotAllowed}).WriteResponse(out, req)
}

func (res *MethodNotAllowedResolution) LogMessage() log.Message {
	return log.WarningMsg("%d %s",
		go_http.StatusMethodNotAllowed,
		go_http.StatusText(go_http.StatusMethodNotAllowed))
}