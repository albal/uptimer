package monitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// SSLChecker checks SSL certificate validity and expiration.
type SSLChecker struct{}

func (c *SSLChecker) Type() string { return models.MonitorSSL }

func (c *SSLChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	host := monitor.URL
	// Strip protocol if present
	for _, prefix := range []string{"https://", "http://"} {
		if len(host) > len(prefix) && host[:len(prefix)] == prefix {
			host = host[len(prefix):]
			break
		}
	}
	// Strip path
	if idx := findByte(host, '/'); idx >= 0 {
		host = host[:idx]
	}
	// Add port if not present
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = host + ":443"
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: time.Duration(monitor.TimeoutSeconds) * time.Second},
		"tcp", host, &tls.Config{},
	)
	responseTime := time.Since(start)
	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("TLS connection failed: %w", err),
		}
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("no certificates found"),
		}
	}

	cert := certs[0]
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)
	reminderDays := monitor.SSLExpiryReminder
	if reminderDays == 0 {
		reminderDays = 30
	}

	if time.Now().After(cert.NotAfter) {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("SSL certificate expired on %s", cert.NotAfter.Format("2006-01-02")),
		}
	}

	if daysUntilExpiry <= reminderDays {
		return CheckResult{
			Status:       models.StatusDegraded,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("SSL certificate expires in %d days (on %s)", daysUntilExpiry, cert.NotAfter.Format("2006-01-02")),
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: responseTime,
	}
}

func findByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
