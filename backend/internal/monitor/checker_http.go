package monitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// HTTPChecker performs HTTP/HTTPS monitoring checks.
type HTTPChecker struct{}

func (c *HTTPChecker) Type() string { return models.MonitorHTTP }

func (c *HTTPChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	client := &http.Client{
		Timeout: time.Duration(monitor.TimeoutSeconds) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	if !monitor.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	method := monitor.HTTPMethod
	if method == "" {
		method = "GET"
	}

	var bodyReader io.Reader
	if monitor.HTTPBody != "" {
		bodyReader = strings.NewReader(monitor.HTTPBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, monitor.URL, bodyReader)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("creating request: %w", err),
		}
	}

	// Set headers
	for key, value := range monitor.HTTPHeaders {
		req.Header.Set(key, value)
	}

	// Set basic auth if configured
	if monitor.HTTPAuthType == "basic" && monitor.HTTPUsername != "" {
		req.SetBasicAuth(monitor.HTTPUsername, monitor.HTTPPasswordEnc)
	}

	resp, err := client.Do(req)
	responseTime := time.Since(start)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	// Read body (limited) to ensure connection is fully established
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024*1024)) // max 1MB

	// Check against expected status codes
	expectedCodes := monitor.ExpectedStatusCodes
	if len(expectedCodes) == 0 {
		expectedCodes = []int{200}
	}

	statusOK := false
	for _, code := range expectedCodes {
		if resp.StatusCode == code {
			statusOK = true
			break
		}
	}

	if !statusOK {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			StatusCode:   resp.StatusCode,
			Error:        fmt.Errorf("unexpected status code: %d (expected %v)", resp.StatusCode, expectedCodes),
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: responseTime,
		StatusCode:   resp.StatusCode,
	}
}
