package http

import (
	go_http "net/http"
	"sync"

	"github.com/pjsaksa/go-utils/log"
)

type ServerController interface {
	BindAddress() string
	RequestHandler(url []string) RequestHandlerFunc
	SessionCookieName() string
	Login(user, password string) User
}

type User interface {
	Username() string
}

// ------------------------------------------------------------

type Server struct {
	httpServer    go_http.Server
	sessions      map[string]User
	sessionsMutex sync.RWMutex
	ctrl          ServerController
}

func NewServer(ctrl ServerController) *Server {
	srv := &Server{
		sessions: map[string]User{},
		ctrl:     ctrl,
	}

	srv.httpServer = go_http.Server{
		Addr:           srv.ctrl.BindAddress(),
		Handler:        srv,
		MaxHeaderBytes: 1 << 15,
	}
	return srv
}

func (srv *Server) Start() {
	log.INFO("Listening HTTP at %s", srv.httpServer.Addr)
	if err := srv.httpServer.ListenAndServe(); err != go_http.ErrServerClosed {
		panic(err.Error())
	}
}

func (srv *Server) ServeHTTP(out go_http.ResponseWriter, req *go_http.Request) {
	// Initialize log messages.
	reqMsg := log.EventMsg(
		"[%s] %s %s %s",
		req.RemoteAddr,
		req.Method,
		req.URL.EscapedPath(),
		req.URL.RawQuery)
	respMsg := log.FatalMsg("Handler is missing response log message")
	defer func(msg1, msg2 *log.Message) {
		log.LOG(*msg1, *msg2)
	}(&reqMsg, &respMsg)

	// Check if request contains session information.
	var sessionUser User
	{
		var sessionFound bool
		sessionFound, sessionUser = srv.getOpenSession(out, req)

		if sessionFound && sessionUser == nil {
			// Active session has been expired. Request has already been redirected.
			respMsg = log.WarningMsg("No session found")
			return
		}
	}

	urlParts := splitUrlPath(req.URL.EscapedPath())
	if urlParts == nil || len(urlParts) == 0 {
		go_http.Error(out, "Invalid URL", go_http.StatusBadRequest)
		respMsg = log.WarningMsg("Invalid URL \"%s\"", req.URL.EscapedPath())
		return
	}

	if urlParts[0] == "u" {
		// User-specific page handler.

		if sessionUser == nil {
			go_http.Error(out, "Forbidden", go_http.StatusForbidden)
			respMsg = log.ErrorMsg("Forbidden")
			return
		}

		if UrlPartsMatch(urlParts, "u", "sign-out") {
			respMsg = srv.doSignOut(out, req, urlParts, sessionUser)
			return
		}
	} else {
		if UrlPartsMatch(urlParts, "sign-in") {
			respMsg = srv.doSignIn(out, req, urlParts, sessionUser)
			return
		}
	}

	if handler := srv.ctrl.RequestHandler(urlParts); handler != nil {
		respMsg = handler.safeCall(out, req, urlParts, sessionUser)
	} else {
		go_http.NotFound(out, req)
		respMsg = log.WarningMsg("Not Found")
	}
}
