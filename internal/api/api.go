package api

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/pscarreira/gobid/internal/services"
)

type Api struct {
	Router          *chi.Mux
	UsersService    services.UsersService
	ProductsService services.ProductsService
	Sessions        *scs.SessionManager
}
