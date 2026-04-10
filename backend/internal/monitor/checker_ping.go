package monitor

import (
	"context"
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/albal/uptimer/internal/models"
)

// PingChecker performs ICMP ping monitoring checks.
type PingChecker struct{}

func (c *PingChecker) Type() string { return models.MonitorPing }

func (c *PingChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	target := monitor.IPAddress
	if target == "" {
		target = monitor.URL
	}

	pinger, err := probing.NewPinger(target)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("creating pinger: %w", err),
		}
	}

	pinger.Count = 3
	pinger.Timeout = time.Duration(monitor.TimeoutSeconds) * time.Second
	pinger.SetPrivileged(false) // Use unprivileged ICMP

	if err := pinger.RunWithContext(ctx); err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("ping failed: %w", err),
		}
	}

	stats := pinger.Statistics()
	if stats.PacketLoss == 100 {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("100%% packet loss"),
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: stats.AvgRtt,
	}
}
