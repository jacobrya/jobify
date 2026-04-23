package handler

import (
	"net/http"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/middleware"
	"github.com/abzalserikbay/jobify/internal/service"
	"github.com/abzalserikbay/jobify/pkg/response"
	"github.com/abzalserikbay/jobify/pkg/validator"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetProfile godoc
// @Summary      Get the current user's profile
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=map[string]interface{}}
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get profile")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"user":    user,
		"profile": profile,
	})
}

type updateProfileRequest struct {
	Name            string   `json:"name"`
	Bio             string   `json:"bio"`
	Skills          []string `json:"skills"`
	ExperienceYears int      `json:"experience_years"`
	SalaryMin       int      `json:"salary_min"`
	SalaryMax       int      `json:"salary_max"`
	RemoteOnly      bool     `json:"remote_only"`
	GithubURL       string   `json:"github_url"`
}

// UpdateProfile godoc
// @Summary      Update the current user's developer profile
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body updateProfileRequest true "Profile fields"
// @Success      200 {object} response.Response{data=domain.DeveloperProfile}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /me [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := validator.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	profile := &domain.DeveloperProfile{
		UserID:          userID,
		Name:            req.Name,
		Bio:             req.Bio,
		Skills:          req.Skills,
		ExperienceYears: req.ExperienceYears,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		RemoteOnly:      req.RemoteOnly,
		GithubURL:       req.GithubURL,
	}

	if err := h.svc.UpdateProfile(r.Context(), profile); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}

func userIDFromCtx(r *http.Request) (uuid.UUID, error) {
	raw, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
	return uuid.Parse(raw)
}
