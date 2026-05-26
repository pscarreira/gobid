package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pscarreira/gobid/internal/store/pgstore"
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
