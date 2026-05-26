package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/pscarreira/gobid/internal/jsonutils"
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

	id, err := api.ProductsService.CreateProduct(
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

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]string{"id": id.String()})
}
