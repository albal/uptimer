package monitor

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// DNSChecker performs DNS record monitoring.
type DNSChecker struct{}

func (c *DNSChecker) Type() string { return models.MonitorDNS }

func (c *DNSChecker) Check(ctx context.Context, monitor *models.Monitor) CheckResult {
	start := time.Now()

	host := monitor.URL
	recordType := monitor.DNSRecordType
	if recordType == "" {
		recordType = "A"
	}

	resolver := net.Resolver{
		PreferGo: true,
	}

	var records []string
	var err error

	switch strings.ToUpper(recordType) {
	case "A":
		var ips []net.IP
		ips, err = resolver.LookupIP(ctx, "ip4", host)
		for _, ip := range ips {
			records = append(records, ip.String())
		}
	case "AAAA":
		var ips []net.IP
		ips, err = resolver.LookupIP(ctx, "ip6", host)
		for _, ip := range ips {
			records = append(records, ip.String())
		}
	case "CNAME":
		var cname string
		cname, err = resolver.LookupCNAME(ctx, host)
		records = append(records, cname)
	case "MX":
		var mxRecords []*net.MX
		mxRecords, err = resolver.LookupMX(ctx, host)
		for _, mx := range mxRecords {
			records = append(records, mx.Host)
		}
	case "NS":
		var nsRecords []*net.NS
		nsRecords, err = resolver.LookupNS(ctx, host)
		for _, ns := range nsRecords {
			records = append(records, ns.Host)
		}
	case "TXT":
		records, err = resolver.LookupTXT(ctx, host)
	default:
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: time.Since(start),
			Error:        fmt.Errorf("unsupported DNS record type: %s", recordType),
		}
	}

	responseTime := time.Since(start)

	if err != nil {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("DNS lookup failed: %w", err),
		}
	}

	if len(records) == 0 {
		return CheckResult{
			Status:       models.StatusDown,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("no %s records found for %s", recordType, host),
		}
	}

	// If expected value is set, validate
	if monitor.DNSExpectedValue != "" {
		found := false
		for _, record := range records {
			if strings.TrimSuffix(record, ".") == strings.TrimSuffix(monitor.DNSExpectedValue, ".") {
				found = true
				break
			}
		}
		if !found {
			return CheckResult{
				Status:       models.StatusDown,
				ResponseTime: responseTime,
				Error:        fmt.Errorf("expected %s record '%s' not found (got: %s)", recordType, monitor.DNSExpectedValue, strings.Join(records, ", ")),
			}
		}
	}

	return CheckResult{
		Status:       models.StatusUp,
		ResponseTime: responseTime,
	}
}
