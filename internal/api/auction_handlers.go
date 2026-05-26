package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pscarreira/gobid/internal/jsonutils"
	"github.com/pscarreira/gobid/internal/services"
)

func (api *Api) handleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	rawProductID := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(rawProductID)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]string{"error": "invalid product id. must be a valid uuid"})
		return
	}

	_, err = api.ProductsService.GetProductByID(r.Context(), productID)

	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]string{"error": "product not found"})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve product"})
		return
	}

	userId, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{"error": "error in identifying user"})
		return
	}

	api.AuctionLobby.Lock()
	room, ok := api.AuctionLobby.Rooms[productID]
	api.AuctionLobby.Unlock()

	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]string{"error": "auction has ended or does not exist"})
		return
	}

	conn, err := api.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{"error": "could not upgrade connection to websocket"})
		return
	}

	client := services.NewClient(conn, userId, room)
	room.Register <- client
	go client.ReadEventLoop()
	go client.WriteEventLoop()
}
