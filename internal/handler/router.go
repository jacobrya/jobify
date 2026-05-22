package handler

import (
	"net/http"

	"github.com/abzalserikbay/jobify/internal/middleware"
	rediscache "github.com/abzalserikbay/jobify/internal/repository/redis"
	jwtpkg "github.com/abzalserikbay/jobify/pkg/jwt"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Deps struct {
	AuthHandler        *AuthHandler
	UserHandler        *UserHandler
	JobHandler         *JobHandler
	ApplicationHandler *ApplicationHandler
	SavedJobHandler    *SavedJobHandler
	JWT                *jwtpkg.Manager
	RateLimitStore     *rediscache.RateLimitStore
	RateLimitPerMin    int
}

func NewRouter(deps *Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.BodyLimit(1 << 20)) // 1 MB
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))
	r.Use(middleware.RateLimit(deps.RateLimitStore, deps.RateLimitPerMin))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/swagger/*", httpSwagger.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", deps.AuthHandler.Register)
		r.Post("/auth/login", deps.AuthHandler.Login)
		r.Post("/auth/refresh", deps.AuthHandler.Refresh)
		r.Post("/auth/logout", deps.AuthHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(deps.JWT))

			r.Get("/me", deps.UserHandler.GetProfile)
			r.Put("/me", deps.UserHandler.UpdateProfile)

			r.Get("/jobs", deps.JobHandler.List)
			r.Get("/jobs/{id}", deps.JobHandler.GetByID)

			r.Get("/applications", deps.ApplicationHandler.List)
			r.Post("/applications", deps.ApplicationHandler.Create)
			r.Put("/applications/{id}", deps.ApplicationHandler.UpdateStatus)
			r.Delete("/applications/{id}", deps.ApplicationHandler.Delete)

			r.Post("/saved-jobs", deps.SavedJobHandler.Save)
			r.Delete("/saved-jobs/{job_id}", deps.SavedJobHandler.Unsave)
			r.Get("/saved-jobs", deps.SavedJobHandler.List)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/jobs", deps.JobHandler.Create)
				r.Put("/jobs/{id}", deps.JobHandler.Update)
				r.Delete("/jobs/{id}", deps.JobHandler.Delete)
			})
		})
	})

	return r
}
