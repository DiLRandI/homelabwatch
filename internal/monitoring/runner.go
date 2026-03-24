package monitoring

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/deleema/homelabwatch/internal/domain"
)

type CheckStore interface {
	GetChecksDue(context.Context) ([]domain.MonitorCheck, error)
	SaveCheckResult(context.Context, domain.CheckResult) error
}

type Runner struct {
	store CheckStore
}

func NewRunner(store CheckStore) *Runner {
	return &Runner{store: store}
}

func (r *Runner) RunDueChecks(ctx context.Context) ([]domain.CheckResult, error) {
	checks, err := r.store.GetChecksDue(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	results := make([]domain.CheckResult, 0, len(checks))
	for _, item := range checks {
		if !isDue(item.Check, now) {
			continue
		}
		result := runCheck(ctx, item)
		if err := r.store.SaveCheckResult(ctx, result); err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func isDue(check domain.ServiceCheck, now time.Time) bool {
	if !check.Enabled {
		return false
	}
	if check.IntervalSeconds <= 0 || check.LastResult == nil || check.LastResult.CheckedAt.IsZero() {
		return true
	}
	return check.LastResult.CheckedAt.Add(time.Duration(check.IntervalSeconds) * time.Second).Before(now)
}

func runCheck(ctx context.Context, item domain.MonitorCheck) domain.CheckResult {
	check := item.Check
	result := domain.CheckResult{
		CheckID:   check.ID,
		ServiceID: check.ServiceID,
		Status:    domain.HealthStatusUnhealthy,
		CheckedAt: time.Now().UTC(),
	}
	switch check.Type {
	case domain.CheckTypeHTTP:
		runHTTPCheck(ctx, item, &result)
	case domain.CheckTypePing:
		runPingCheck(check, &result)
	default:
		runTCPCheck(ctx, check, &result)
	}
	return result
}

func RunAdhocCheck(ctx context.Context, check domain.ServiceCheck) domain.CheckResult {
	return runCheck(ctx, domain.MonitorCheck{Check: check})
}

func runHTTPCheck(ctx context.Context, item domain.MonitorCheck, result *domain.CheckResult) {
	check := item.Check
	timeout := durationFromTimeout(check.TimeoutSeconds)
	target := check.Target
	if target == "" {
		target = buildCheckTarget(item)
	}
	method := strings.ToUpper(strings.TrimSpace(check.Method))
	if method == "" {
		method = http.MethodGet
	}
	request, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		result.Message = err.Error()
		return
	}
	start := time.Now()
	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		result.LatencyMS = time.Since(start).Milliseconds()
		result.Message = err.Error()
		return
	}
	defer response.Body.Close()
	result.LatencyMS = time.Since(start).Milliseconds()
	result.HTTPStatusCode = response.StatusCode
	result.ResolvedTarget = request.URL.String()
	limited, _ := io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	result.ResponseSizeBytes = limited
	if response.StatusCode < check.ExpectedStatusMin || response.StatusCode > check.ExpectedStatusMax {
		result.Message = response.Status
		return
	}
	result.Status = domain.HealthStatusHealthy
	result.Message = response.Status
}

func runTCPCheck(ctx context.Context, check domain.ServiceCheck, result *domain.CheckResult) {
	timeout := durationFromTimeout(check.TimeoutSeconds)
	start := time.Now()
	target := check.Target
	if target == "" {
		target = fmt.Sprintf("%s:%d", firstNonEmpty(check.Host, check.HostValue), check.Port)
	}
	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", target)
	if err != nil {
		result.LatencyMS = time.Since(start).Milliseconds()
		result.Message = err.Error()
		return
	}
	_ = conn.Close()
	result.Status = domain.HealthStatusHealthy
	result.LatencyMS = time.Since(start).Milliseconds()
	result.ResolvedTarget = target
	result.Message = "tcp ok"
}

func runPingCheck(check domain.ServiceCheck, result *domain.CheckResult) {
	target := check.Target
	if target == "" {
		target = firstNonEmpty(check.Host, check.HostValue)
	}
	host := trimSchemeHost(target)
	pinger, err := probing.NewPinger(host)
	if err != nil {
		result.Message = err.Error()
		return
	}
	pinger.Count = 1
	pinger.Timeout = durationFromTimeout(check.TimeoutSeconds)
	pinger.SetPrivileged(true)
	if err := pinger.Run(); err != nil {
		result.Message = err.Error()
		return
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		result.Message = "no reply"
		return
	}
	result.Status = domain.HealthStatusHealthy
	result.LatencyMS = stats.AvgRtt.Milliseconds()
	result.ResolvedTarget = host
	result.Message = fmt.Sprintf("%d/%d packets", stats.PacketsRecv, stats.PacketsSent)
}

func durationFromTimeout(timeoutSeconds int) time.Duration {
	if timeoutSeconds <= 0 {
		return 5 * time.Second
	}
	return time.Duration(timeoutSeconds) * time.Second
}

func trimSchemeHost(value string) string {
	if strings.Contains(value, "://") {
		parts := strings.SplitN(value, "://", 2)
		value = parts[1]
	}
	if index := strings.Index(value, "/"); index >= 0 {
		value = value[:index]
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		return host
	}
	return value
}

func buildCheckTarget(item domain.MonitorCheck) string {
	check := item.Check
	switch check.Type {
	case domain.CheckTypeTCP:
		if firstNonEmpty(check.Host, check.HostValue) == "" || check.Port == 0 {
			return ""
		}
		return fmt.Sprintf("%s:%d", firstNonEmpty(check.Host, check.HostValue), check.Port)
	case domain.CheckTypePing:
		return firstNonEmpty(check.Host, check.HostValue)
	default:
		scheme := strings.TrimSpace(check.Protocol)
		if scheme == "" {
			scheme = firstNonEmpty(item.Service.Scheme, "http")
		}
		host := firstNonEmpty(check.Host, check.HostValue, item.Service.Host, item.Service.HostValue)
		path := strings.TrimSpace(check.Path)
		if path == "" {
			path = item.Service.Path
		}
		if host == "" {
			return ""
		}
		base := fmt.Sprintf("%s://%s", scheme, host)
		if check.Port > 0 && check.Port != 80 && check.Port != 443 {
			base = fmt.Sprintf("%s:%d", base, check.Port)
		}
		if path == "" {
			return base
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return base + path
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
