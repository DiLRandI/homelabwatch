package notifications

import (
	"fmt"
	"slices"

	"github.com/deleema/homelabwatch/internal/domain"
)

type NotificationEvent struct {
	Type       domain.NotificationEventType `json:"eventType"`
	Title      string                       `json:"title"`
	Message    string                       `json:"message"`
	OccurredAt string                       `json:"occurredAt"`
	Resource   map[string]any               `json:"resource"`
	Payload    map[string]any               `json:"payload"`
}

func matchesRule(rule domain.NotificationRule, event NotificationEvent) bool {
	if !rule.Enabled || rule.EventType != event.Type {
		return false
	}
	switch event.Type {
	case domain.NotificationEventServiceHealthChanged:
		statuses := stringSliceFilter(rule.Filters["statuses"])
		if len(statuses) == 0 {
			return true
		}
		current := fmt.Sprint(event.Payload["currentStatus"])
		return slices.Contains(statuses, current)
	case domain.NotificationEventWorkerFailed:
		minFailures := intFilter(rule.Filters["minConsecutiveFailures"], 3)
		actual, _ := event.Payload["consecutiveFailures"].(int)
		if actual == 0 {
			if typed, ok := event.Payload["consecutiveFailures"].(float64); ok {
				actual = int(typed)
			}
		}
		return actual >= minFailures
	default:
		return true
	}
}

func stringSliceFilter(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := fmt.Sprint(item); text != "" {
				items = append(items, text)
			}
		}
		return items
	default:
		return nil
	}
}

func intFilter(value any, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case float64:
		return int(typed)
	case jsonNumber:
		if parsed, err := typed.Int64(); err == nil {
			return int(parsed)
		}
	}
	return fallback
}

type jsonNumber interface {
	Int64() (int64, error)
}
