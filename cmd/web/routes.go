package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()
	router.Use(recoverPanic(app))
	router.Use(logRequest(app))
	router.Use(secureHeaders)
	router.Use(middleware.GetHead)

	// Public & Private Routes
	router.Group(func(r chi.Router) {
		r.Use(app.session.Enable)
		r.Use(noSurf)
		r.Use(authenticate(app))

		// Public Routes
		r.Group(func(r chi.Router) {
			r.Get("/", app.home)
			r.Get("/snippet/{id:[0-9]+}", app.showSnippet)
			r.Get("/user/signup", app.signupUserForm)
			r.Post("/user/signup", app.signupUser)
			r.Get("/user/login", app.loginUserForm)
			r.Post("/user/login", app.loginUser)

			r.Get("/ping", ping)
			r.Get("/about", app.about)
		})

		// Private Routes
		r.Group(func(r chi.Router) {
			r.Use(requireAuthentication(app))

			r.Get("/user/change-password", app.changePasswordForm)
			r.Post("/user/change-password", app.changePassword)
			r.Get("/user/profile", app.userProfile)
			r.Post("/user/logout", app.logoutUser)
			r.Get("/snippet/create", app.createSnippetForm)
			r.Post("/snippet/create", app.createSnippet)
		})
	})

	// Static Files Routes
	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Mount("/static", http.StripPrefix("/static", fileServer))

	router.MethodNotAllowed(app.methodNotAllowed)

	return router
}
