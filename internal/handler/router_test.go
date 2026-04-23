package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/abzalserikbay/jobify/docs"
)

func TestRouter_SwaggerRoute(t *testing.T) {
	router := NewRouter(&Deps{})

	tests := []struct {
		path       string
		wantStatus []int
		wantBody   string
	}{
		{"/swagger/index.html", []int{http.StatusOK}, "swagger"},
		{"/swagger/doc.json", []int{http.StatusOK}, "Jobify API"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			ok := false
			for _, s := range tt.wantStatus {
				if rec.Code == s {
					ok = true
					break
				}
			}
			if !ok {
				t.Fatalf("status: got %d, want one of %v (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(strings.ToLower(rec.Body.String()), strings.ToLower(tt.wantBody)) {
				t.Errorf("body missing %q: got %q", tt.wantBody, rec.Body.String())
			}
		})
	}
}
