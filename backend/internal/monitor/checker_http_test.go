package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/albal/uptimer/internal/models"
)

func TestHTTPChecker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	checker := &HTTPChecker{}
	monitor := &models.Monitor{
		URL:                 ts.URL,
		TimeoutSeconds:      5,
		ExpectedStatusCodes: []int{200},
		HTTPMethod:          "GET",
	}

	result := checker.Check(context.Background(), monitor)

	if result.Status != models.StatusUp {
		t.Errorf("expected status Up, got %s", result.Status)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", result.StatusCode)
	}
}

func TestHTTPChecker_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	checker := &HTTPChecker{}
	monitor := &models.Monitor{
		URL:                 ts.URL,
		TimeoutSeconds:      5,
		ExpectedStatusCodes: []int{200},
		HTTPMethod:          "GET",
	}

	result := checker.Check(context.Background(), monitor)

	if result.Status != models.StatusDown {
		t.Errorf("expected status Down, got %s", result.Status)
	}
}
