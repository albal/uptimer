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

// KeywordChecker performs HTTP requests and checks for keyword presence.
type KeywordChecker struct{}

func (c *KeywordChecker) Type() string { return models.MonitorKeyword }

func (c *KeywordChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	client := &http.Client{
		Timeout: time.Duration(monitor.TimeoutSeconds) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", monitor.URL, nil)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("creating request: %w", err),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // max 5MB
	responseTime := time.Since(start)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			StatusCode:   resp.StatusCode,
			Error:        fmt.Errorf("reading body: %w", err),
		}
	}

	bodyStr := string(body)
	keywordFound := strings.Contains(bodyStr, monitor.Keyword)

	switch monitor.KeywordType {
	case "exists":
		if !keywordFound {
			return CheckResult{
				Status:       models.StatusDown,
				ResponseTime: responseTime,
				StatusCode:   resp.StatusCode,
				Error:        fmt.Errorf("keyword '%s' not found on page", monitor.Keyword),
			}
		}
	case "not_exists":
		if keywordFound {
			return CheckResult{
				Status:       models.StatusDown,
				ResponseTime: responseTime,
				StatusCode:   resp.StatusCode,
				Error:        fmt.Errorf("keyword '%s' found on page (should not exist)", monitor.Keyword),
			}
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: responseTime,
		StatusCode:   resp.StatusCode,
	}
}
