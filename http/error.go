package http

import (
	"fmt"
	go_http "net/http"

	"github.com/pjsaksa/go-utils/log"
)

type HttpError struct {
	Message string
	Status  int
}

func (err HttpError) Error() string {
	return fmt.Sprintf(`%s (%d)`, err.Message, err.Status)
}

func (err HttpError) WriteResponse(out go_http.ResponseWriter) log.Message {
	go_http.Error(out, err.Message, err.Status)

	switch {
	case err.Status >= 500:
		return log.ErrorMsg(err.Message)
	case err.Status >= 400:
		return log.WarningMsg(err.Message)
	default:
		return log.DebugMsg(err.Message)
	}
}
