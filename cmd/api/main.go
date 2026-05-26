package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pscarreira/gobid/internal/api"
	"github.com/pscarreira/gobid/internal/services"
)

func main() {
	gob.Register(uuid.UUID{})

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		os.Getenv("GOBID_DATABASE_USER"),
		os.Getenv("GOBID_DATABASE_PASSWORD"),
		os.Getenv("GOBID_DATABASE_HOST"),
		os.Getenv("GOBID_DATABASE_PORT"),
		os.Getenv("GOBID_DATABASE_NAME")))

	if err != nil {
		panic(err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to the database!")

	s := scs.New()
	s.Store = pgxstore.New(pool)
	s.Lifetime = 24 * time.Hour
	s.Cookie.HttpOnly = true
	s.Cookie.SameSite = http.SameSiteLaxMode
	s.Cookie.Secure = false

	api := api.Api{
		Router:          chi.NewMux(),
		UsersService:    services.NewUsersService(pool),
		ProductsService: services.NewProductsService(pool),
		BidsService:     services.NewBidsService(pool),
		Sessions:        s,
		WsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				//em produção, implemente uma verificação adequada para garantir que apenas origens confiáveis possam se conectar
				return true
			},
		},
		AuctionLobby: services.AuctionLobby{
			Rooms: make(map[uuid.UUID]*services.AuctionRoom),
		},
	}

	api.BindRoutes()

	fmt.Println("Starting server on port 3000...")

	if err := http.ListenAndServe("localhost:3000", api.Router); err != nil {
		panic(err)
	}

}
