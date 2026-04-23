package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/middleware"
	"github.com/abzalserikbay/jobify/internal/service"
	"github.com/abzalserikbay/jobify/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SavedJobHandler struct {
	svc *service.SavedJobService
}

func NewSavedJobHandler(svc *service.SavedJobService) *SavedJobHandler {
	return &SavedJobHandler{svc: svc}
}

// @Summary      Save a job
// @Tags         saved-jobs
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body object{job_id=string} true "Job ID"
// @Success      201
// @Failure      400 {object} response.Response
// @Failure      409 {object} response.Response
// @Router       /saved-jobs [post]
func (h *SavedJobHandler) Save(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextKeyUserID).(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		JobID uuid.UUID `json:"job_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.JobID == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "invalid job_id")
		return
	}

	if err := h.svc.Save(r.Context(), userID, req.JobID); err != nil {
		if err == domain.ErrConflict {
			response.Error(w, http.StatusConflict, "already saved")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary      Unsave a job
// @Tags         saved-jobs
// @Security     BearerAuth
// @Param        job_id path string true "Job ID"
// @Success      204
// @Failure      404 {object} response.Response
// @Router       /saved-jobs/{job_id} [delete]
func (h *SavedJobHandler) Unsave(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextKeyUserID).(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobID, err := uuid.Parse(chi.URLParam(r, "job_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job_id")
		return
	}

	if err := h.svc.Unsave(r.Context(), userID, jobID); err != nil {
		if err == domain.ErrNotFound {
			response.Error(w, http.StatusNotFound, "not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      List saved jobs
// @Tags         saved-jobs
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Page"
// @Param        limit query int false "Limit"
// @Success      200 {object} response.Response
// @Router       /saved-jobs [get]
func (h *SavedJobHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextKeyUserID).(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	jobs, total, err := h.svc.List(r.Context(), userID, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.WithMeta(w, http.StatusOK, jobs, &response.Meta{
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
