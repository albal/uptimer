package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// APIChecker performs API monitoring with response assertions.
type APIChecker struct{}

func (c *APIChecker) Type() string { return models.MonitorAPI }

func (c *APIChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	client := &http.Client{
		Timeout: time.Duration(monitor.TimeoutSeconds) * time.Second,
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

	for key, value := range monitor.HTTPHeaders {
		req.Header.Set(key, value)
	}

	if req.Header.Get("Content-Type") == "" && monitor.HTTPBody != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if monitor.HTTPAuthType == "basic" && monitor.HTTPUsername != "" {
		req.SetBasicAuth(monitor.HTTPUsername, monitor.HTTPPasswordEnc)
	}

	resp, err := client.Do(req)
	responseTime := time.Since(start)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("API request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))

	// Run assertions
	for _, assertion := range monitor.APIAssertions {
		if err := runAssertion(assertion, resp, body, responseTime); err != nil {
			return CheckResult{
				Status:       models.StatusDown,
				ResponseTime: responseTime,
				StatusCode:   resp.StatusCode,
				Error:        err,
			}
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: responseTime,
		StatusCode:   resp.StatusCode,
	}
}

// runAssertion evaluates a single API assertion.
func runAssertion(assertion models.APIAssertion, resp *http.Response, body []byte, responseTime time.Duration) error {
	var actual string

	switch assertion.Source {
	case "status_code":
		actual = strconv.Itoa(resp.StatusCode)
	case "response_time":
		actual = strconv.Itoa(int(responseTime.Milliseconds()))
	case "header":
		actual = resp.Header.Get(assertion.Property)
	case "body":
		// Simple JSON path extraction for body assertions
		if assertion.Property != "" {
			actual = extractJSONValue(body, assertion.Property)
		} else {
			actual = string(body)
		}
	default:
		return fmt.Errorf("unknown assertion source: %s", assertion.Source)
	}

	switch assertion.Comparison {
	case "equals":
		if actual != assertion.Value {
			return fmt.Errorf("assertion failed: %s.%s expected '%s' got '%s'", assertion.Source, assertion.Property, assertion.Value, actual)
		}
	case "contains":
		if !strings.Contains(actual, assertion.Value) {
			return fmt.Errorf("assertion failed: %s.%s does not contain '%s'", assertion.Source, assertion.Property, assertion.Value)
		}
	case "greater_than":
		actualNum, err1 := strconv.ParseFloat(actual, 64)
		expectedNum, err2 := strconv.ParseFloat(assertion.Value, 64)
		if err1 != nil || err2 != nil {
			return fmt.Errorf("assertion failed: cannot compare non-numeric values")
		}
		if actualNum <= expectedNum {
			return fmt.Errorf("assertion failed: %s.%s value %f is not greater than %f", assertion.Source, assertion.Property, actualNum, expectedNum)
		}
	case "less_than":
		actualNum, err1 := strconv.ParseFloat(actual, 64)
		expectedNum, err2 := strconv.ParseFloat(assertion.Value, 64)
		if err1 != nil || err2 != nil {
			return fmt.Errorf("assertion failed: cannot compare non-numeric values")
		}
		if actualNum >= expectedNum {
			return fmt.Errorf("assertion failed: %s.%s value %f is not less than %f", assertion.Source, assertion.Property, actualNum, expectedNum)
		}
	default:
		return fmt.Errorf("unknown comparison: %s", assertion.Comparison)
	}

	return nil
}

// extractJSONValue extracts a value from JSON using a simple dot-notation path.
func extractJSONValue(data []byte, path string) string {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return ""
	}

	parts := strings.Split(path, ".")
	current := obj
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return fmt.Sprintf("%v", current)
		}
	}

	return fmt.Sprintf("%v", current)
}
