package handler

import (
	"errors"
	"net/http"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/service"
	"github.com/abzalserikbay/jobify/pkg/response"
	"github.com/abzalserikbay/jobify/pkg/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ApplicationHandler struct {
	svc *service.ApplicationService
}

func NewApplicationHandler(svc *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

// List godoc
// @Summary      List the current user's applications
// @Tags         applications
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=[]domain.Application}
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /applications [get]
func (h *ApplicationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	apps, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list applications")
		return
	}

	response.JSON(w, http.StatusOK, apps)
}

type createAppRequest struct {
	JobID string `json:"job_id"`
	Note  string `json:"note"`
}

// Create godoc
// @Summary      Save/apply to a job
// @Tags         applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body createAppRequest true "Application payload"
// @Success      201 {object} response.Response{data=domain.Application}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /applications [post]
func (h *ApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createAppRequest
	if err := validator.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	jobID, err := uuid.Parse(req.JobID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job_id")
		return
	}

	app, err := h.svc.Create(r.Context(), userID, jobID, req.Note)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create application")
		return
	}

	response.JSON(w, http.StatusCreated, app)
}

type updateStatusRequest struct {
	Status domain.ApplicationStatus `json:"status"`
}

// UpdateStatus godoc
// @Summary      Update application status
// @Tags         applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string              true "Application UUID"
// @Param        body body updateStatusRequest true "New status"
// @Success      200 {object} response.Response{data=map[string]string}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /applications/{id} [put]
func (h *ApplicationHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid application id")
		return
	}

	var req updateStatusRequest
	if err := validator.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, userID, req.Status); errors.Is(err, domain.ErrForbidden) {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	} else if errors.Is(err, domain.ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, "invalid status")
		return
	} else if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

// Delete godoc
// @Summary      Delete an application
// @Tags         applications
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Application UUID"
// @Success      200 {object} response.Response{data=map[string]string}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /applications/{id} [delete]
func (h *ApplicationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid application id")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); errors.Is(err, domain.ErrForbidden) {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	} else if errors.Is(err, domain.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "application not found")
		return
	} else if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete application")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
