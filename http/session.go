package http

import (
	"fmt"
	go_http "net/http"
	"time"

	"github.com/pjsaksa/go-utils/log"
)

func (srv *Server) doSignIn(req *go_http.Request, cookies *[]*go_http.Cookie) Resolution {
	if req.Method != "POST" {
		return &MethodNotAllowedResolution{Allowed: "POST"}
	}
	if res := srv.ctrl.VerifyCsrf(
		req.URL.EscapedPath(),
		"",
		req.PostFormValue("csrf"),
	); res != nil {
		return res
	}

	u := req.PostFormValue("user")
	p := req.PostFormValue("password")
	if u != "" {
		if user := srv.ctrl.Login(u, p); user != nil {
			token, err := srv.ctrl.NewSession(user)
			if err != nil {
				log.ERROR("NewSession: %s", err)
				panic(&ErrorResolution{
					Status:  go_http.StatusInternalServerError,
					Message: fmt.Sprintf("Internal server error: %s", err),
				})
			}

			log.INFO("Sign-in '%s'", u)

			cookieInfo := srv.ctrl.SessionDetails()
			*cookies = append(*cookies, &go_http.Cookie{
				Name:     cookieInfo.Name,
				Value:    token,
				Path:     "/",
				MaxAge:   int(cookieInfo.MaxAge.Seconds()),
				Secure:   cookieInfo.Secure,
				HttpOnly: cookieInfo.HttpOnly,
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

func (srv *Server) doSignOut(req *go_http.Request, cookies *[]*go_http.Cookie, activeUser User, sessionKey string) Resolution {
	if req.Method != "POST" {
		return &MethodNotAllowedResolution{Allowed: "POST"}
	}
	if res := srv.ctrl.VerifyCsrf(
		req.URL.EscapedPath(),
		activeUser.Username(),
		req.PostFormValue("csrf"),
	); res != nil {
		return res
	}

	srv.ctrl.DeleteSession(sessionKey)
	log.INFO("Sign-out '%s'", activeUser.Username())

	cookieInfo := srv.ctrl.SessionDetails()
	*cookies = append(*cookies, &go_http.Cookie{
		Name:     cookieInfo.Name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: cookieInfo.HttpOnly,
	})

	return &RedirectResolution{
		Status: go_http.StatusSeeOther,
		Url:    "/",
	}
}

func (srv *Server) getOpenSession(req *go_http.Request, cookies *[]*go_http.Cookie) (User, string) {
	cookieInfo := srv.ctrl.SessionDetails()
	if cookie, err := req.Cookie(cookieInfo.Name); err != go_http.ErrNoCookie && cookie != nil && len(cookie.Value) > 0 {
		var session Session
		var key string

		ok := len(cookie.Value) == cookieInfo.CookieSize

		if ok {
			key = cookie.Value[:cookieInfo.KeySize]
			session, ok = srv.ctrl.GetSession(key)
			defer session.Unlock()
			if !ok {
				log.WARNING("Requested session not found")
			}
		}

		if ok && session == nil {
			// Found nil Session. Make noise:
			log.ERROR(`http.Server.getOpenSession: "sessions" had nil entry: %s`, key)
			srv.ctrl.DeleteSession(key)

			ok = false
		}

		if ok && !session.VerifyToken(cookie.Value) {
			log.WARNING("Session key, verification fail: %q", key)

			ok = false
		}

		if ok && time.Since(session.RefreshTime()) > cookieInfo.MaxAge {
			log.INFO("Session expired '%s'", session.User().Username())

			ok = false
		}

		if ok {
			// Refresh session (unless it's fresh enough)
			if time.Since(session.RefreshTime()) > time.Hour {
				session.SetRefreshTime(time.Now())

				*cookies = append(*cookies, &go_http.Cookie{
					Name:     cookieInfo.Name,
					Value:    cookie.Value,
					Path:     "/",
					MaxAge:   int(cookieInfo.MaxAge.Seconds()),
					Secure:   cookieInfo.Secure,
					HttpOnly: cookieInfo.HttpOnly,
				})
			}

			// Return valid user information
			return session.User(), key
		} else {
			// Request contained a session cookie but one of the above checks
			// caused the session to be rejected

			*cookies = append(*cookies, &go_http.Cookie{
				Name:     cookieInfo.Name,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: cookieInfo.HttpOnly,
			})

			panic(&RedirectResolution{
				Status: go_http.StatusSeeOther,
				Url:    "/",
			})
		}
	}

	return nil, ""
}
