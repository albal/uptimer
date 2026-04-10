package monitor

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// PortChecker performs TCP port monitoring checks.
type PortChecker struct{}

func (c *PortChecker) Type() string { return models.MonitorPort }

func (c *PortChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	host := monitor.IPAddress
	if host == "" {
		host = monitor.URL
	}

	port := 80
	if monitor.Port != nil {
		port = *monitor.Port
	}

	address := fmt.Sprintf("%s:%d", host, port)
	dialer := net.Dialer{Timeout: time.Duration(monitor.TimeoutSeconds) * time.Second}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	responseTime := time.Since(start)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("TCP connection failed: %w", err),
		}
	}
	conn.Close()

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: responseTime,
	}
}
