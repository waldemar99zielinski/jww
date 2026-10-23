package health_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	"jww/internal/infrastructure/http_server/routers/health"
	healthservice "jww/internal/service/health"
)

type dependencyChecker struct {
	err         error
	checkCalled bool
}

func (c *dependencyChecker) Check(domain.AppContext) error {
	c.checkCalled = true
	return c.err
}

func TestGetHealth(t *testing.T) {
	database := &dependencyChecker{err: errors.New("connection refused")}
	healthService := healthservice.NewService(
		healthservice.Config{},
		healthservice.Dependency{Name: "database", Required: true, Checker: database},
	)
	router := chi.NewRouter()
	health.Mount(router, router, &kernel.Deps{HealthService: healthService})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}

	var body struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Status != healthservice.HealthStatusOK {
		t.Errorf("status = %q, want %q", body.Data.Status, healthservice.HealthStatusOK)
	}
	if database.checkCalled {
		t.Error("liveness must not check dependencies")
	}
}

func TestGetReadiness(t *testing.T) {
	tests := []struct {
		name        string
		databaseErr error
		optionalErr error
		wantStatus  int
		wantOK      bool
	}{
		{name: "all dependencies available", wantStatus: http.StatusOK, wantOK: true},
		{name: "optional dependency unavailable stays ready", optionalErr: errors.New("provider offline"), wantStatus: http.StatusOK, wantOK: true},
		{name: "database unavailable", databaseErr: errors.New("connection refused"), wantStatus: http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthService := healthservice.NewService(
				healthservice.Config{},
				healthservice.Dependency{Name: "database", Required: true, Checker: &dependencyChecker{err: tt.databaseErr}},
				healthservice.Dependency{Name: "openai", Checker: &dependencyChecker{err: tt.optionalErr}},
			)
			router := chi.NewRouter()
			health.Mount(router, router, &kernel.Deps{HealthService: healthService})

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/readiness", nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			body := rec.Body.String()
			if tt.wantOK {
				var response struct {
					Data struct {
						Status string `json:"status"`
					} `json:"data"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if response.Data.Status != "ok" {
					t.Errorf("status = %q, want ok", response.Data.Status)
				}
			}
			for _, detail := range []string{"database", "openai", "latency_ms", "required"} {
				if strings.Contains(body, detail) {
					t.Errorf("response leaked dependency detail %q: %s", detail, body)
				}
			}
		})
	}
}
