package http

import (
	"fmt"
	"io"
	go_http "net/http"
)

func RequireUser(user User) {
	if user == nil {
		panic(&ErrorResolution{Status: go_http.StatusForbidden})
	}
}

func RequireEmptyContent(req *go_http.Request) {
	if req.ContentLength > 0 {
		panic(&ErrorResolution{
			Status:  go_http.StatusRequestEntityTooLarge,
			Message: "Content must be empty",
		})
	}
}

func ReadContent(req *go_http.Request, max int64) []byte {
	switch {
	case req.ContentLength < 0:
		panic(&ErrorResolution{Status: go_http.StatusLengthRequired})
	case req.ContentLength > max:
		panic(&ErrorResolution{Status: go_http.StatusRequestEntityTooLarge})
	}

	inputBytes := make([]byte, req.ContentLength)
	inputBytes_n, err := io.ReadFull(req.Body, inputBytes)
	switch {
	case err != nil:
		panic(&ErrorResolution{
			Status:  go_http.StatusInternalServerError,
			Message: fmt.Sprintf("Reading payload failed (%s)", err.Error()),
		})
	case int64(inputBytes_n) != req.ContentLength:
		panic(&ErrorResolution{
			Status:  go_http.StatusInternalServerError,
			Message: fmt.Sprintf("Reading payload failed (%d bytes < %d bytes)", inputBytes_n, req.ContentLength),
		})
	}

	return inputBytes
}
