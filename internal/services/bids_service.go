package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pscarreira/gobid/internal/store/pgstore"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrBidIsTooLow     = errors.New("bid amount must be greater than current highest bid")
)

type BidsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewBidsService(pool *pgxpool.Pool) BidsService {
	return BidsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (bs *BidsService) PlaceBid(
	ctx context.Context,
	product_id, bidder_id uuid.UUID,
	bid_amount float64,
) (pgstore.Bid, error) {
	// Amount > previous amount
	// Amount > base price
	product, err := bs.queries.GetProductById(ctx, product_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, ErrProductNotFound
		}
		return pgstore.Bid{}, err
	}

	highestBid, err := bs.queries.GetHighestBidByProductID(ctx, product_id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return pgstore.Bid{}, err
	}

	if bid_amount <= product.BasePrice || highestBid.BidAmount >= bid_amount {
		return pgstore.Bid{}, ErrBidIsTooLow
	}

	highestBid, err = bs.queries.CreateBid(ctx, pgstore.CreateBidParams{
		ProductID: product_id,
		BidderID:  bidder_id,
		BidAmount: bid_amount,
	})

	if err != nil {
		return pgstore.Bid{}, err
	}

	return highestBid, nil
}
