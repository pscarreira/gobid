package api

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pscarreira/gobid/internal/jsonutils"
	"github.com/pscarreira/gobid/internal/services"
	"github.com/pscarreira/gobid/internal/usecase/product"
)

func (api *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	userID, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]string{"error": "must be logged in"})
		return
	}

	product_id, err := api.ProductsService.CreateProduct(
		r.Context(),
		userID,
		data.ProductName,
		data.Description,
		data.BasePrice,
		data.AuctionEnd,
	)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{"error": "failed to create product"})
		return
	}

	ctx, _ := context.WithDeadline(context.Background(), data.AuctionEnd)

	auctionRoom := services.NewAuctionRoom(ctx, product_id, api.BidsService)

	go auctionRoom.Run()

	api.AuctionLobby.Lock()
	api.AuctionLobby.Rooms[product_id] = auctionRoom
	api.AuctionLobby.Unlock()

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]string{
		"message": "product created successfully. auction room created and will end until " + data.AuctionEnd.Format(time.Kitchen),
		"id":      product_id.String(),
	})
}
