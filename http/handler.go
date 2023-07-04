package http

import (
	go_http "net/http"

	"github.com/pjsaksa/go-utils/log"
)

type RequestHandlerFunc func(go_http.ResponseWriter, *go_http.Request, []string, User) log.Message

func (handler RequestHandlerFunc) safeCall(out go_http.ResponseWriter, req *go_http.Request, urlParts []string, user User) (respMsg log.Message) {
	defer func() {
		if err := recover(); err != nil {
			respMsg = err.(HttpError).WriteResponse(out)
		}
	}()

	return handler(out, req, urlParts, user)
}
