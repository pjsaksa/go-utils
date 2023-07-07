package http

import (
	"crypto/rand"
	"encoding/base64"
	go_http "net/http"

	"github.com/pjsaksa/go-utils/log"
)

func (srv *Server) doSignIn(out go_http.ResponseWriter, req *go_http.Request, urlParts []string, activeUser User) log.Message {
	switch req.Method {
	case "POST":
		u := req.PostFormValue("user")
		p := req.PostFormValue("password")
		if u != "" {
			if user := srv.ctrl.Login(u, p); user != nil {
				// Protect shared parts
				srv.sessionsMutex.Lock()
				defer srv.sessionsMutex.Unlock()

				token := newSessionToken()
				srv.sessions[token] = user
				go_http.SetCookie(out, &go_http.Cookie{
					Name:  srv.ctrl.SessionCookieName(),
					Value: token,
					Path:  "/",
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

func (srv *Server) doSignOut(out go_http.ResponseWriter, req *go_http.Request, urlParts []string, activeUser User) log.Message {
	switch req.Method {
	case "POST":
		// Protect shared parts
		srv.sessionsMutex.Lock()
		defer srv.sessionsMutex.Unlock()

		for token, user := range srv.sessions {
			if user == activeUser {
				delete(srv.sessions, token)
			}
		}

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

func (srv *Server) getOpenSession(out go_http.ResponseWriter, req *go_http.Request) (bool, User) {
	if cookie, err := req.Cookie(srv.ctrl.SessionCookieName()); err != go_http.ErrNoCookie && cookie != nil && len(cookie.Value) > 0 {
		// Protect shared parts
		srv.sessionsMutex.RLock()
		defer srv.sessionsMutex.RUnlock()

		user, ok := srv.sessions[cookie.Value]
		if ok && user != nil {
			return true, user
		} else {
			go_http.SetCookie(out, &go_http.Cookie{
				Name:   srv.ctrl.SessionCookieName(),
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})

			go_http.Redirect(out, req, "/", go_http.StatusSeeOther)
			return true, nil
		}
	}

	return false, nil
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
