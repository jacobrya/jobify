package handler

import (
	"errors"
	"net/http"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/service"
	"github.com/abzalserikbay/jobify/pkg/response"
	"github.com/abzalserikbay/jobify/pkg/validator"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register godoc
// @Summary      Register a new developer account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body registerRequest true "Credentials"
// @Success      201 {object} response.Response{data=domain.User}
// @Failure      400 {object} response.Response
// @Failure      409 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := validator.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if !validator.ValidateEmail(req.Email) {
		response.Error(w, http.StatusBadRequest, "invalid email")
		return
	}
	if !validator.ValidatePassword(req.Password) {
		response.Error(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	user, err := h.svc.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, domain.ErrConflict) {
		response.Error(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "registration failed")
		return
	}

	response.JSON(w, http.StatusCreated, user)
}

// Login godoc
// @Summary      Exchange credentials for a JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Credentials"
// @Success      200 {object} response.Response{data=map[string]string}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := validator.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.svc.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, domain.ErrUnauthorized) {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "login failed")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"token": token})
}
