package product

import (
	"context"
	"time"

	"github.com/pscarreira/gobid/internal/validator"
)

type CreateProductRequest struct {
	ProductName string    `json:"product_name"`
	Description string    `json:"description"`
	BasePrice   float64   `json:"base_price"`
	AuctionEnd  time.Time `json:"auction_end"`
}

func (r CreateProductRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(r.ProductName), "product_name", "product name must not be blank")
	eval.CheckField(validator.NotBlank(r.Description), "description", "description must not be blank")
	eval.CheckField(
		validator.MinChars(r.Description, 10) && validator.MaxChars(r.Description, 255),
		"description",
		"description must be between 10 and 255 characters long",
	)
	eval.CheckField(validator.Positive(r.BasePrice), "base_price", "base price must be greater than zero")
	eval.CheckField(validator.Future(r.AuctionEnd), "auction_end", "auction end must be in the future")
	eval.CheckField(validator.FutureWithMinDuration(r.AuctionEnd, time.Hour*2), "auction_end", "auction needs to end at least 2 hours from now")
	return eval
}
