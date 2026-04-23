package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/repository"
	"github.com/abzalserikbay/jobify/internal/service"
	"github.com/abzalserikbay/jobify/pkg/response"
	"github.com/abzalserikbay/jobify/pkg/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type JobHandler struct {
	svc     *service.JobService
	userSvc *service.UserService
}

func NewJobHandler(svc *service.JobService, userSvc *service.UserService) *JobHandler {
	return &JobHandler{svc: svc, userSvc: userSvc}
}

// List godoc
// @Summary      List jobs with filters and pagination
// @Tags         jobs
// @Produce      json
// @Security     BearerAuth
// @Param        page        query int    false "Page number (default 1)"
// @Param        limit       query int    false "Page size (default 20)"
// @Param        skills      query string false "Comma-separated skills filter"
// @Param        remote      query bool   false "Only remote jobs"
// @Param        salary_min  query int    false "Minimum salary"
// @Param        salary_max  query int    false "Maximum salary"
// @Success      200 {object} response.Response{data=[]domain.Job,meta=response.Meta}
// @Failure      401 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /jobs [get]
func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	f := repository.JobFilter{
		Page:  parseIntDefault(q.Get("page"), 1),
		Limit: parseIntDefault(q.Get("limit"), 20),
	}

	if skills := q.Get("skills"); skills != "" {
		f.Skills = strings.Split(skills, ",")
	}
	if remote := q.Get("remote"); remote == "true" {
		b := true
		f.Remote = &b
	}
	if v := q.Get("salary_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.SalaryMin = &n
		}
	}
	if v := q.Get("salary_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.SalaryMax = &n
		}
	}

	jobs, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list jobs")
		return
	}

	pages := total / f.Limit
	if total%f.Limit != 0 {
		pages++
	}

	response.WithMeta(w, http.StatusOK, jobs, &response.Meta{
		Total: total,
		Page:  f.Page,
		Limit: f.Limit,
		Pages: pages,
	})
}

// GetByID godoc
// @Summary      Get a job by ID with skill match against current user
// @Tags         jobs
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Job UUID"
// @Success      200 {object} response.Response{data=domain.JobWithMatch}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /jobs/{id} [get]
func (h *JobHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	var userSkills []string
	if userID, parseErr := userIDFromCtx(r); parseErr == nil {
		if _, profile, err := h.userSvc.GetProfile(r.Context(), userID); err == nil && profile != nil {
			userSkills = profile.Skills
		}
	}

	job, err := h.svc.GetByID(r.Context(), id, userSkills)
	if errors.Is(err, domain.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "job not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get job")
		return
	}

	response.JSON(w, http.StatusOK, job)
}

type createJobRequest struct {
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
	SalaryMin   int      `json:"salary_min"`
	SalaryMax   int      `json:"salary_max"`
	IsRemote    bool     `json:"is_remote"`
	Location    string   `json:"location"`
	URL         string   `json:"url"`
}

// Create godoc
// @Summary      Create a new job (admin only)
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body createJobRequest true "Job payload"
// @Success      201 {object} response.Response{data=domain.Job}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /jobs [post]
func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest
	if err := validator.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Title == "" || req.Company == "" {
		response.Error(w, http.StatusBadRequest, "title and company are required")
		return
	}

	job := &domain.Job{
		Title:       req.Title,
		Company:     req.Company,
		Description: req.Description,
		Skills:      req.Skills,
		SalaryMin:   req.SalaryMin,
		SalaryMax:   req.SalaryMax,
		IsRemote:    req.IsRemote,
		Location:    req.Location,
		URL:         req.URL,
	}

	if err := h.svc.Create(r.Context(), job); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	response.JSON(w, http.StatusCreated, job)
}

// Update godoc
// @Summary      Update an existing job (admin only)
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string            true "Job UUID"
// @Param        body body createJobRequest  true "Job payload"
// @Success      200 {object} response.Response{data=domain.Job}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /jobs/{id} [put]
func (h *JobHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	var req createJobRequest
	if err := validator.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	job := &domain.Job{
		ID:          id,
		Title:       req.Title,
		Company:     req.Company,
		Description: req.Description,
		Skills:      req.Skills,
		SalaryMin:   req.SalaryMin,
		SalaryMax:   req.SalaryMax,
		IsRemote:    req.IsRemote,
		Location:    req.Location,
		URL:         req.URL,
		IsActive:    true,
	}

	if err := h.svc.Update(r.Context(), job); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update job")
		return
	}

	response.JSON(w, http.StatusOK, job)
}

// Delete godoc
// @Summary      Delete a job (admin only)
// @Tags         jobs
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Job UUID"
// @Success      200 {object} response.Response{data=map[string]string}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /jobs/{id} [delete]
func (h *JobHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete job")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func parseIntDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return def
}
