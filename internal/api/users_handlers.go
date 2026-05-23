package api

import (
	"errors"
	"net/http"

	"github.com/pscarreira/gobid/internal/jsonutils"
	"github.com/pscarreira/gobid/internal/services"
	"github.com/pscarreira/gobid/internal/usecase/user"
)

func (api *Api) handleSignUpUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.CreateUserReq](r)
	if err != nil {
		if problems == nil {
			problems = map[string]string{"error": err.Error()}
		}
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	id, err := api.UsersService.CreateUser(
		r.Context(),
		data.Username,
		data.Password,
		data.Email,
		data.Bio,
	)

	if err != nil {
		if errors.Is(err, services.ErrDuplicatedEmailOrUsername) {
			_ = jsonutils.EncodeJson(w, r, http.StatusConflict, map[string]string{"error": "email or username already in use"})
			return
		}
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{"id": id})
}

func (api *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.LoginUserReq](r)
	if err != nil {
		if problems == nil {
			problems = map[string]string{"error": err.Error()}
		}
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	id, err := api.UsersService.AuthenticateUser(
		r.Context(),
		data.Email,
		data.Password,
	)

	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	err = api.Sessions.RenewToken(r.Context())
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	api.Sessions.Put(r.Context(), "AuthenticatedUserId", id)

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{"message": "logged in successfully"})
}

func (api *Api) handleLogoutUser(w http.ResponseWriter, r *http.Request) {
	err := api.Sessions.RenewToken(r.Context())
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	api.Sessions.Remove(r.Context(), "AuthenticatedUserId")

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{"message": "logged out successfully"})
}
