package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// HeartbeatChecker checks if heartbeat pings are arriving within the grace period.
type HeartbeatChecker struct{}

func (c *HeartbeatChecker) Type() string { return models.MonitorHeartbeat }

func (c *HeartbeatChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	gracePeriod := time.Duration(monitor.HeartbeatGraceSec) * time.Second
	if gracePeriod == 0 {
		gracePeriod = 5 * time.Minute
	}

	if monitor.HeartbeatLastPing == nil {
		// Never received a ping, check if monitor was just created
		if time.Since(monitor.CreatedAt) < gracePeriod {
			return CheckResult{
				Status:       models.StatusUp,
				ResponseTime: time.Since(start),
			}
		}
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("no heartbeat received"),
		}
	}

	timeSinceLastPing := time.Since(*monitor.HeartbeatLastPing)
	if timeSinceLastPing > gracePeriod {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("heartbeat overdue by %s (last ping: %s ago)", timeSinceLastPing-gracePeriod, timeSinceLastPing.Round(time.Second)),
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: time.Since(start),
	}
}
