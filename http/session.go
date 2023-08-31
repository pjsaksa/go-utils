package http

import (
	"crypto/rand"
	"encoding/base64"
	go_http "net/http"
	"time"

	"github.com/pjsaksa/go-utils/log"
)

func (srv *Server) doSignIn(out go_http.ResponseWriter, req *go_http.Request) log.Message {
	switch req.Method {
	case "POST":
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

				go_http.SetCookie(out, &go_http.Cookie{
					Name:   srv.ctrl.SessionCookieName(),
					Value:  token,
					Path:   "/",
					MaxAge: int(srv.ctrl.SessionMaxAge().Seconds()),
				})
				go_http.Redirect(out, req, "/u/", go_http.StatusSeeOther)
				return log.InfoMsg("Sign-in '%s'", u)
			}
		}
		go_http.Error(out, "Forbidden", go_http.StatusForbidden)
		return log.ErrorMsg("Invalid sign-in '%s'", u)

	default:
		out.Header().Add("Allow", "GET")
		go_http.Error(out, "Method Not Allowed", go_http.StatusMethodNotAllowed)
		return log.WarningMsg("Method Not Allowed (%s)", req.Method)
	}
}

func (srv *Server) doSignOut(out go_http.ResponseWriter, req *go_http.Request, activeUser User, activeCookie string) log.Message {
	switch req.Method {
	case "POST":
		// Mutex-zone
		func() {
			srv.sessionsMutex.Lock()
			defer srv.sessionsMutex.Unlock()

			delete(srv.sessions, activeCookie)
			srv.ctrl.RefreshSession(activeCookie, srv.sessions)
		}()

		go_http.SetCookie(out, &go_http.Cookie{
			Name:   srv.ctrl.SessionCookieName(),
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		go_http.Redirect(out, req, "/", go_http.StatusSeeOther)
		return log.InfoMsg("Sign-out: " + activeUser.Username())

	default:
		out.Header().Add("Allow", "GET")
		go_http.Error(out, "Method Not Allowed", go_http.StatusMethodNotAllowed)
		return log.WarningMsg("Method Not Allowed (%s)", req.Method)
	}
}

func (srv *Server) getOpenSession(out go_http.ResponseWriter, req *go_http.Request) (bool, User, string) {
	if cookie, err := req.Cookie(srv.ctrl.SessionCookieName()); err != go_http.ErrNoCookie && cookie != nil && len(cookie.Value) > 0 {
		srv.sessionsMutex.Lock()
		defer srv.sessionsMutex.Unlock()

		session, ok := srv.sessions[cookie.Value]

		if ok && session == nil {
			// SessionMap contains invalid entry. If this happens, make noise
			// because it's better to find out what causes it.

			log.ERROR(`http.Server: "sessions" had nil entry: %s`, cookie.Value)

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
			// Refresh session, unless it's fresh enough
			if time.Since(session.RefreshTime) > time.Hour {
				session.RefreshTime = time.Now()
				srv.ctrl.RefreshSession(cookie.Value, srv.sessions)

				go_http.SetCookie(out, &go_http.Cookie{
					Name:   srv.ctrl.SessionCookieName(),
					Value:  cookie.Value,
					Path:   "/",
					MaxAge: int(srv.ctrl.SessionMaxAge().Seconds()),
				})
			}

			// Return valid user information
			return true, session.User, cookie.Value
		} else {
			// Request had session cookie but one of the above things caused the
			// session to fail

			go_http.SetCookie(out, &go_http.Cookie{
				Name:   srv.ctrl.SessionCookieName(),
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})

			go_http.Redirect(out, req, "/", go_http.StatusSeeOther)
			return true, nil, ""
		}
	}

	return false, nil, ""
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
		panic(err.Error())
	case n != sessionTokenSize:
		panic("newSessionToken(): invalid number of output bytes")
	}

	return base64.StdEncoding.EncodeToString(data[:])
}
