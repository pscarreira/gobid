package api

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func (api *Api) BindRoutes() {
	api.Router.Use(middleware.RequestID)
	api.Router.Use(middleware.Recoverer)
	api.Router.Use(middleware.Logger)
	api.Router.Use(api.Sessions.LoadAndSave)

	environment := os.Getenv("GOBID_ENVIRONMENT")
	isProduction := environment == "production"

	if isProduction {
		csrfMiddleware := csrf.Protect(
			[]byte(os.Getenv("GOBID_CSRF_KEY")),
			csrf.Secure(true),
		)
		api.Router.Use(csrfMiddleware)
	}

	api.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			if isProduction {
				r.Get("/csrf-token", api.HandleGetCSRFToken)
			}
			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", api.handleSignUpUser)
				r.Post("/login", api.handleLoginUser)
				r.Group(func(r chi.Router) {
					r.Use(api.AuthMiddleware)
					r.Post("/logout", api.handleLogoutUser)
				})
			})
			r.Route("/products", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(api.AuthMiddleware)
					r.Post("/", api.handleCreateProduct)

					r.Get("/ws/subscribe/{product_id}", api.handleSubscribeUserToAuction)
				})
			})
		})

	})
}
