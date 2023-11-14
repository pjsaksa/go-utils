package http

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	go_http "net/http"
	"time"

	"github.com/pjsaksa/go-utils/log"
)

func (srv *Server) doSignIn(req *go_http.Request, cookies *[]*go_http.Cookie) Resolution {
	if req.Method != "POST" {
		return &MethodNotAllowedResolution{Allowed: "POST"}
	}

	u := req.PostFormValue("user")
	p := req.PostFormValue("password")
	if u != "" {
		if user := srv.ctrl.Login(u, p); user != nil {
			var token string

			// Mutex-zone
			func() {
				srv.sessionsMutex.Lock()
				defer srv.sessionsMutex.Unlock()

				// Create tokens until a fresh one is found
				for {
					token = newSessionToken()
					if _, exists := srv.sessions[token]; !exists {
						break
					}
				}

				srv.sessions[token] = &Session{
					User:        user,
					RefreshTime: time.Now(),
				}
				srv.ctrl.RefreshSession(token, srv.sessions)
			}()

			log.INFO("Sign-in '%s'", u)

			*cookies = append(*cookies, &go_http.Cookie{
				Name:   srv.ctrl.SessionCookieName(),
				Value:  token,
				Path:   "/",
				MaxAge: int(srv.ctrl.SessionMaxAge().Seconds()),
			})

			return &RedirectResolution{
				Status: go_http.StatusSeeOther,
				Url:    "/u/",
			}
		}
	}
	return &ErrorResolution{
		Status:  go_http.StatusForbidden,
		Message: fmt.Sprintf("Invalid sign-in '%s'", u),
	}
}

func (srv *Server) doSignOut(req *go_http.Request, cookies *[]*go_http.Cookie, activeUser User, activeCookie string) Resolution {
	if req.Method != "POST" {
		return &MethodNotAllowedResolution{Allowed: "POST"}
	}

	// Mutex-zone
	func() {
		srv.sessionsMutex.Lock()
		defer srv.sessionsMutex.Unlock()

		delete(srv.sessions, activeCookie)
		srv.ctrl.RefreshSession(activeCookie, srv.sessions)
	}()

	log.INFO("Sign-out '%s'", activeUser.Username())

	*cookies = append(*cookies, &go_http.Cookie{
		Name:   srv.ctrl.SessionCookieName(),
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	return &RedirectResolution{
		Status: go_http.StatusSeeOther,
		Url:    "/",
	}
}

func (srv *Server) getOpenSession(req *go_http.Request, cookies *[]*go_http.Cookie) (User, string) {
	if cookie, err := req.Cookie(srv.ctrl.SessionCookieName()); err != go_http.ErrNoCookie && cookie != nil && len(cookie.Value) > 0 {
		srv.sessionsMutex.Lock()
		defer srv.sessionsMutex.Unlock()

		session, ok := srv.sessions[cookie.Value]
		if !ok {
			log.WARNING("Requested session not found")
		}

		if ok && session == nil {
			// SessionMap contains nil entry. Make noise because this needs to
			// be tracked down.
			log.ERROR(`http.Server.getOpenSession: "sessions" had nil entry: %s`, cookie.Value)

			// Delete invalid session entry
			delete(srv.sessions, cookie.Value)
			srv.ctrl.RefreshSession(cookie.Value, srv.sessions)

			ok = false
		}

		if ok && time.Since(session.RefreshTime) > srv.ctrl.SessionMaxAge() {
			// Session has expired
			log.INFO("Session expired '%s'", session.User.Username())

			ok = false
		}

		if ok {
			// Refresh session (unless it's fresh enough)
			if time.Since(session.RefreshTime) > time.Hour {
				session.RefreshTime = time.Now()
				srv.ctrl.RefreshSession(cookie.Value, srv.sessions)

				*cookies = append(*cookies, &go_http.Cookie{
					Name:   srv.ctrl.SessionCookieName(),
					Value:  cookie.Value,
					Path:   "/",
					MaxAge: int(srv.ctrl.SessionMaxAge().Seconds()),
				})
			}

			// Return valid user information
			return session.User, cookie.Value
		} else {
			// Request contained a session cookie but one of the above checks
			// caused the session to be rejected

			*cookies = append(*cookies, &go_http.Cookie{
				Name:   srv.ctrl.SessionCookieName(),
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})

			panic(&RedirectResolution{
				Status: go_http.StatusSeeOther,
				Url:    "/",
			})
		}
	}

	return nil, ""
}

// ------------------------------------------------------------

func newSessionToken() string {
	const (
		sessionTokenSize = 24 // 24 bytes => 32 in base64
	)

	var data [sessionTokenSize]byte
	n, err := rand.Read(data[:])
	switch {
	case err != nil:
		panic(&ErrorResolution{
			Status:  go_http.StatusInternalServerError,
			Message: fmt.Sprintf("http.newSessionToken: %s", err.Error()),
		})
	case n != sessionTokenSize:
		panic(&ErrorResolution{
			Status:  go_http.StatusInternalServerError,
			Message: fmt.Sprintf("http.newSessionToken: invalid number of output bytes"),
		})
	}

	return base64.StdEncoding.EncodeToString(data[:])
}
