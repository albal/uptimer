package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/albal/uptimer/internal/models"
)

// DomainChecker checks domain registration expiration.
type DomainChecker struct{}

func (c *DomainChecker) Type() string { return models.MonitorDomain }

func (c *DomainChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	domain := monitor.URL
	// Strip protocol and path
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(domain, prefix) {
			domain = domain[len(prefix):]
			break
		}
	}
	if idx := strings.Index(domain, "/"); idx >= 0 {
		domain = domain[:idx]
	}

	raw, err := whois.Whois(domain)
	responseTime := time.Since(start)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("WHOIS lookup failed: %w", err),
		}
	}

	parsed, err := whoisparser.Parse(raw)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("parsing WHOIS data: %w", err),
		}
	}

	expiryDate, parseErr := time.Parse("2006-01-02T15:04:05Z", parsed.Domain.ExpirationDate)
	if parseErr != nil {
		expiryDate, parseErr = time.Parse("2006-01-02", parsed.Domain.ExpirationDate)
	}
	if parseErr != nil {
		expiryDate, parseErr = time.Parse(time.RFC3339, parsed.Domain.ExpirationDate)
	}
	if parseErr != nil {
		return CheckResult{
			Status:       models.StatusUp,
			ResponseTime: responseTime,
		}
	}

	daysUntilExpiry := int(time.Until(expiryDate).Hours() / 24)
	reminderDays := monitor.DomainExpiryReminder
	if reminderDays == 0 {
		reminderDays = 30
	}

	if daysUntilExpiry < 0 {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("domain expired on %s", expiryDate.Format("2006-01-02")),
		}
	}

	if daysUntilExpiry <= reminderDays {
		return CheckResult{
			Status:       models.StatusDegraded,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("domain expires in %d days (on %s)", daysUntilExpiry, expiryDate.Format("2006-01-02")),
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: responseTime,
	}
}
