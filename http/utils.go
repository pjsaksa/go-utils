package http

import (
	"fmt"
	"io"
	go_http "net/http"

	"github.com/pjsaksa/go-utils/log"
)

func ServeStaticFile(fileName, contentType string, maxAge int) RequestHandlerFunc {
	return func(out go_http.ResponseWriter, req *go_http.Request, urlParts []string, user User) log.Message {
		switch req.Method {
		case "", "GET":
			out.Header().Set("Content-Type", contentType)
			if maxAge > 0 {
				out.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
			}
			go_http.ServeFile(out, req, "static/"+fileName)
			return log.DebugMsg("Ok")
		default:
			out.Header().Add("Allow", "GET")
			go_http.Error(out, "Method Not Allowed", go_http.StatusMethodNotAllowed)
			return log.WarningMsg("Method Not Allowed (%s)", req.Method)
		}
	}
}

// ------------------------------------------------------------

func RequireEmptyContent(req *go_http.Request) {
	if req.ContentLength > 0 {
		panic(HttpError{
			Message: "Content must be empty",
			Status:  go_http.StatusRequestEntityTooLarge,
		})
	}
}

func ReadContent(req *go_http.Request, max int64) []byte {
	switch {
	case req.ContentLength < 0:
		panic(HttpError{
			Message: "Content length required",
			Status:  go_http.StatusLengthRequired,
		})
	case req.ContentLength > max:
		panic(HttpError{
			Message: "Content too large",
			Status:  go_http.StatusRequestEntityTooLarge,
		})
	}

	inputBytes := make([]byte, req.ContentLength)
	inputBytes_n, err := io.ReadFull(req.Body, inputBytes)
	switch {
	case err != nil:
		panic(HttpError{
			Message: fmt.Sprintf("Reading payload failed (%s)", err.Error()),
			Status:  go_http.StatusInternalServerError,
		})
	case int64(inputBytes_n) != req.ContentLength:
		panic(HttpError{
			Message: fmt.Sprintf("Reading payload failed (%d bytes < %d bytes)", inputBytes_n, req.ContentLength),
			Status:  go_http.StatusInternalServerError,
		})
	}

	return inputBytes
}
