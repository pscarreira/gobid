package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pscarreira/gobid/internal/store/pgstore"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewProductsService(pool *pgxpool.Pool) ProductsService {
	return ProductsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (ps *ProductsService) CreateProduct(
	ctx context.Context,
	seller_id uuid.UUID,
	product_name string,
	description string,
	base_price float64,
	auction_end_time time.Time,
) (uuid.UUID, error) {

	id, err := ps.queries.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID:    seller_id,
		ProductName: product_name,
		Description: description,
		BasePrice:   base_price,
		AuctionEnd:  auction_end_time,
	})

	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func (ps *ProductsService) GetProductByID(ctx context.Context, id uuid.UUID) (pgstore.Product, error) {
	product, err := ps.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Product{}, ErrProductNotFound
		}
		return pgstore.Product{}, err
	}
	return product, nil
}
