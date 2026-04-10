package monitor

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// UDPChecker performs UDP monitoring checks.
type UDPChecker struct{}

func (c *UDPChecker) Type() string { return models.MonitorUDP }

func (c *UDPChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	host := monitor.IPAddress
	if host == "" {
		host = monitor.URL
	}

	port := 53
	if monitor.Port != nil {
		port = *monitor.Port
	}

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("udp", address, time.Duration(monitor.TimeoutSeconds)*time.Second)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("UDP connection failed: %w", err),
		}
	}
	defer conn.Close()

	// Set read/write deadline
	deadline := time.Now().Add(time.Duration(monitor.TimeoutSeconds) * time.Second)
	conn.SetDeadline(deadline)

	// Send data if configured
	if monitor.UDPData != "" {
		_, err := conn.Write([]byte(monitor.UDPData))
		if err != nil {
			return CheckResult{
				Status:       models.StatusDown,
				ResponseTime: time.Since(start),
				Error:        fmt.Errorf("UDP write failed: %w", err),
			}
		}
	}

	// If we expect a response, read it
	if monitor.UDPExpected != "" {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		responseTime := time.Since(start)
		if err != nil {
			return CheckResult{
				Status:       models.StatusDown,
				ResponseTime: responseTime,
				Error:        fmt.Errorf("UDP read failed: %w", err),
			}
		}

		response := string(buf[:n])
		if !strings.Contains(response, monitor.UDPExpected) {
			return CheckResult{
				Status:       models.StatusDown,
				ResponseTime: responseTime,
				Error:        fmt.Errorf("UDP response does not contain expected data"),
			}
		}

		return CheckResult{
			Status:       models.StatusUp,
			ResponseTime: responseTime,
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: time.Since(start),
	}
}
