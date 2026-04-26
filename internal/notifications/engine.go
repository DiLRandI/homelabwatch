package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/events"
)

type Store interface {
	ListEnabledNotificationRules(context.Context, domain.NotificationEventType) ([]domain.NotificationRule, error)
	GetNotificationChannelForSend(context.Context, string) (domain.NotificationChannel, error)
	CreateNotificationDelivery(context.Context, domain.NotificationDelivery) (domain.NotificationDelivery, error)
	UpdateNotificationDelivery(context.Context, domain.NotificationDelivery) (domain.NotificationDelivery, error)
}

type Engine struct {
	store  Store
	bus    *events.Bus
	logger *slog.Logger
}

func NewEngine(store Store, bus *events.Bus, logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	return &Engine{store: store, bus: bus, logger: logger.With("component", "notifications")}
}

func (e *Engine) Start(ctx context.Context) {
	ch := e.bus.Subscribe(128)
	go func() {
		defer e.bus.Unsubscribe(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case envelope, ok := <-ch:
				if !ok {
					return
				}
				if envelope.Type == "notification" {
					continue
				}
				e.handleEnvelope(ctx, envelope)
			}
		}
	}()
}

func (e *Engine) handleEnvelope(ctx context.Context, envelope domain.EventEnvelope) {
	notification, ok := translateEnvelope(envelope)
	if !ok {
		return
	}
	rules, err := e.store.ListEnabledNotificationRules(ctx, notification.Type)
	if err != nil {
		e.logger.Warn("notification rule lookup failed", "err", err)
		return
	}
	for _, rule := range rules {
		if !matchesRule(rule, notification) {
			continue
		}
		for _, channelID := range rule.ChannelIDs {
			channel, err := e.store.GetNotificationChannelForSend(ctx, channelID)
			if err != nil || !channel.Enabled {
				continue
			}
			delivery, err := e.store.CreateNotificationDelivery(ctx, domain.NotificationDelivery{
				RuleID:      rule.ID,
				ChannelID:   channel.ID,
				EventType:   notification.Type,
				Status:      domain.NotificationDeliveryPending,
				Message:     "pending",
				AttemptedAt: time.Now().UTC(),
			})
			if err != nil {
				e.logger.Warn("notification delivery create failed", "err", err)
				continue
			}
			if err := Send(ctx, channel, notification); err != nil {
				delivery.Status = domain.NotificationDeliveryFailed
				delivery.Message = err.Error()
			} else {
				delivery.Status = domain.NotificationDeliverySent
				delivery.Message = "sent"
			}
			saved, err := e.store.UpdateNotificationDelivery(ctx, delivery)
			if err != nil {
				e.logger.Warn("notification delivery update failed", "err", err)
				continue
			}
			e.bus.Publish(domain.EventEnvelope{
				Type:       "notification",
				Resource:   "notification",
				ID:         saved.ID,
				Action:     "delivered",
				Payload:    saved,
				OccurredAt: time.Now().UTC(),
			})
		}
	}
}

func translateEnvelope(envelope domain.EventEnvelope) (NotificationEvent, bool) {
	switch envelope.Type + ":" + envelope.Action {
	case "service:health_changed":
		outcome, ok := envelope.Payload.(domain.CheckResultOutcome)
		if !ok {
			return NotificationEvent{}, false
		}
		return baseEvent(domain.NotificationEventServiceHealthChanged, fmt.Sprintf("%s health changed", outcome.Service.Name), fmt.Sprintf("%s changed from %s to %s", outcome.Service.Name, outcome.PreviousServiceStatus, outcome.CurrentServiceStatus), envelope, map[string]any{"type": "service", "id": outcome.Service.ID, "action": "health_changed"}, map[string]any{"previousStatus": outcome.PreviousServiceStatus, "currentStatus": outcome.CurrentServiceStatus, "service": outcome.Service, "check": outcome.Check, "result": outcome.Result}), true
	case "check:failed":
		outcome, ok := envelope.Payload.(domain.CheckResultOutcome)
		if !ok {
			return NotificationEvent{}, false
		}
		return baseEvent(domain.NotificationEventCheckFailed, fmt.Sprintf("%s check failed", firstNonEmpty(outcome.Check.Name, outcome.Service.Name)), firstNonEmpty(outcome.Result.Message, "check failed"), envelope, map[string]any{"type": "check", "id": outcome.Check.ID, "action": "failed"}, map[string]any{"service": outcome.Service, "check": outcome.Check, "result": outcome.Result}), true
	case "check:recovered":
		outcome, ok := envelope.Payload.(domain.CheckResultOutcome)
		if !ok {
			return NotificationEvent{}, false
		}
		return baseEvent(domain.NotificationEventCheckRecovered, fmt.Sprintf("%s check recovered", firstNonEmpty(outcome.Check.Name, outcome.Service.Name)), firstNonEmpty(outcome.Result.Message, "check recovered"), envelope, map[string]any{"type": "check", "id": outcome.Check.ID, "action": "recovered"}, map[string]any{"service": outcome.Service, "check": outcome.Check, "result": outcome.Result}), true
	case "device:created":
		device, ok := envelope.Payload.(domain.Device)
		if !ok {
			return NotificationEvent{}, false
		}
		return baseEvent(domain.NotificationEventDeviceCreated, "New device discovered", firstNonEmpty(device.DisplayName, device.Hostname, device.IdentityKey), envelope, map[string]any{"type": "device", "id": device.ID, "action": "created"}, map[string]any{"device": device}), true
	case "discovered-service:created":
		service, ok := envelope.Payload.(domain.DiscoveredService)
		if !ok {
			return NotificationEvent{}, false
		}
		return baseEvent(domain.NotificationEventDiscoveredServiceCreated, "New service discovered", firstNonEmpty(service.Name, service.URL), envelope, map[string]any{"type": "discovered-service", "id": service.ID, "action": "created"}, map[string]any{"discoveredService": service}), true
	case "worker:failed":
		outcome, ok := envelope.Payload.(domain.JobRunOutcome)
		if !ok {
			return NotificationEvent{}, false
		}
		return baseEvent(domain.NotificationEventWorkerFailed, fmt.Sprintf("%s worker failed", outcome.JobName), outcome.LastError, envelope, map[string]any{"type": "worker", "id": outcome.JobName, "action": "failed"}, map[string]any{"jobName": outcome.JobName, "consecutiveFailures": outcome.ConsecutiveFailures, "lastError": outcome.LastError}), true
	default:
		return NotificationEvent{}, false
	}
}

func baseEvent(eventType domain.NotificationEventType, title, message string, envelope domain.EventEnvelope, resource map[string]any, payload map[string]any) NotificationEvent {
	occurredAt := envelope.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	return NotificationEvent{
		Type:       eventType,
		Title:      title,
		Message:    message,
		OccurredAt: occurredAt.Format(time.RFC3339Nano),
		Resource:   resource,
		Payload:    payload,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
