package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aesuhaendi/go-snippetbox/pkg/models"

	"github.com/justinas/nosurf"
)

type Middleware func(http.Handler) http.Handler

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		// fmt.Println("^secureHeaders")
		next.ServeHTTP(w, r)
		// fmt.Println("$secureHeaders")
	})
}

func logRequest(app *application) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
			// fmt.Println("^logRequest")
			next.ServeHTTP(w, r)
			// fmt.Println("$logRequest")
		})
	}
}

func recoverPanic(app *application) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "close")
					app.serverError(w, fmt.Errorf("%s", err))
				}
			}()
			// fmt.Println("^recoverPanic")
			next.ServeHTTP(w, r)
			// fmt.Println("$recoverPanic")
		})
	}
}

func requireAuthentication(app *application) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !app.isAuthenticated(r) {
				app.session.Put(r, "redirectPathAfterLogin", r.URL.Path)
				http.Redirect(w, r, "/user/login", http.StatusSeeOther)
				return
			}
			w.Header().Add("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	}
}

func authenticate(app *application) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			exists := app.session.Exists(r, "authenticatedUserID")
			if !exists {
				next.ServeHTTP(w, r)
				return
			}

			user, err := app.users.Get(app.session.GetInt(r, "authenticatedUserID"))
			if errors.Is(err, models.ErrNoRecord) || !user.Active {
				app.session.Remove(r, "authenticatedUserID")
				next.ServeHTTP(w, r)
				return
			} else if err != nil {
				app.serverError(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}
