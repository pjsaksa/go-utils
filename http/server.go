package http

import (
	"context"
	go_http "net/http"
	"time"

	"github.com/pjsaksa/go-utils/log"
)

type ServerController interface {
	BindAddress() string
	SessionDetails() SessionDetails
	ConfigureHttpServer(*go_http.Server)

	HandleRequest(*go_http.Request, []string, User) Resolution
	MessageSummary(*go_http.Request, Resolution)

	Login(user, password string) User
	NewSession(user User) (string, error)
	GetSession(string) (Session, bool)
	DeleteSession(string)

	VerifyCsrf(string, string, string) Resolution
}

type User interface {
	Username() string
}

type Session interface {
	Unlock()

	User() User
	VerifyToken(string) bool

	RefreshTime() time.Time
	SetRefreshTime(time.Time)

	Address() string
	UserAgent() string
}

type SessionDetails struct {
	Name     string
	MaxAge   time.Duration
	Secure   bool
	HttpOnly bool
	SameSite go_http.SameSite

	CookieSize int
	KeySize    int
}

// ------------------------------------------------------------

type Server struct {
	ctrl       ServerController
	httpServer go_http.Server
}

func NewServer(ctrl ServerController) *Server {
	srv := &Server{
		ctrl: ctrl,
	}

	srv.httpServer = go_http.Server{
		Addr:           srv.ctrl.BindAddress(),
		Handler:        srv,
		MaxHeaderBytes: 1 << 15,
	}

	ctrl.ConfigureHttpServer(&srv.httpServer)
	return srv
}

func (srv *Server) Start() {
	log.INFO("Listening HTTP at %s", srv.httpServer.Addr)
	if err := srv.httpServer.ListenAndServe(); err != go_http.ErrServerClosed {
		panic(err.Error())
	}
}

func (srv *Server) StartTLS(certFile, keyFile string) {
	log.INFO("Listening HTTPS at %s", srv.httpServer.Addr)
	if err := srv.httpServer.ListenAndServeTLS(certFile, keyFile); err != go_http.ErrServerClosed {
		panic(err.Error())
	}
}

func (srv *Server) Shutdown() {
	srv.httpServer.Shutdown(context.Background())
}

// ------------------------------------------------------------

func (srv *Server) ServeHTTP(out go_http.ResponseWriter, req *go_http.Request) {
	var cookies []*go_http.Cookie

	// Handle request
	resolution := srv.handleRequest(req, &cookies)

	// Produce response
	for _, c := range cookies {
		go_http.SetCookie(out, c)
	}
	resolution.WriteResponse(out, req)

	srv.ctrl.MessageSummary(req, resolution)
}

func (srv *Server) handleRequest(req *go_http.Request, cookies *[]*go_http.Cookie) (resolution Resolution) {
	defer func() {
		if err := recover(); err != nil {
			switch errT := err.(type) {
			case Resolution:
				resolution = errT
			default:
				panic(err)
			}
		}
	}()

	var urlParts []string
	urlParts, resolution = splitUrlPath(req.URL.EscapedPath())
	if resolution != nil {
		return
	}

	var sessionUser User
	sessionUser, resolution = srv.handleSessions(urlParts, req, cookies)
	if resolution != nil {
		return
	}

	resolution = srv.ctrl.HandleRequest(req, urlParts, sessionUser)
	if resolution != nil {
		return
	}

	resolution = &ErrorResolution{Status: go_http.StatusNotFound}
	return
}

func (srv *Server) handleSessions(urlParts []string, req *go_http.Request, cookies *[]*go_http.Cookie) (User, Resolution) {
	// Check if request contains session information.
	sessionUser, sessionKey := srv.getOpenSession(req, cookies)

	if urlParts[0] == "u" {
		// User-specific page handler.

		if sessionUser == nil {
			return nil, &ErrorResolution{Status: go_http.StatusForbidden}
		}

		if UrlPartsMatch(urlParts, "u", "sign-out") {
			return nil, srv.doSignOut(req, cookies, sessionUser, sessionKey)
		}
	} else {
		if UrlPartsMatch(urlParts, "sign-in") {
			return nil, srv.doSignIn(req, cookies)
		}
	}

	return sessionUser, nil
}
