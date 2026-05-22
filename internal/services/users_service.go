package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pscarreira/gobid/internal/store/pgstore"
	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicatedEmailOrPassword = errors.New("username or email already exists")

type UsersService struct {
	queries *pgstore.Queries
	pool    *pgxpool.Pool
}

func NewUsersService(pool *pgxpool.Pool) UsersService {
	return UsersService{
		queries: pgstore.New(pool),
		pool:    pool,
	}
}

func (us *UsersService) CreateUser(ctx context.Context, username, password, email, bio string) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return uuid.UUID{}, err
	}

	args := pgstore.CreateUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Bio:          bio,
	}

	id, err := us.queries.CreateUser(ctx, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.UUID{}, ErrDuplicatedEmailOrPassword
		}
		return uuid.UUID{}, err
	}
	return id, nil
}
