package app

import (
	"context"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/notifications"
)

func (a *App) ListNotificationChannels(ctx context.Context) ([]domain.NotificationChannel, error) {
	return a.store.ListNotificationChannels(ctx)
}

func (a *App) SaveNotificationChannel(ctx context.Context, input domain.NotificationChannel) (domain.NotificationChannel, error) {
	item, err := a.store.SaveNotificationChannel(ctx, input)
	if err != nil {
		return domain.NotificationChannel{}, err
	}
	a.publish("notification", item.ID, "channel_saved", item)
	return item, nil
}

func (a *App) DeleteNotificationChannel(ctx context.Context, id string) error {
	if err := a.store.DeleteNotificationChannel(ctx, id); err != nil {
		return err
	}
	a.publish("notification", id, "channel_deleted", nil)
	return nil
}

func (a *App) TestNotificationChannel(ctx context.Context, id string) (domain.NotificationDelivery, error) {
	channel, err := a.store.GetNotificationChannelForSend(ctx, id)
	if err != nil {
		return domain.NotificationDelivery{}, err
	}
	delivery, err := a.store.CreateNotificationDelivery(ctx, domain.NotificationDelivery{
		ChannelID:   id,
		EventType:   domain.NotificationEventServiceHealthChanged,
		Status:      domain.NotificationDeliveryPending,
		Message:     "pending",
		AttemptedAt: time.Now().UTC(),
	})
	if err != nil {
		return domain.NotificationDelivery{}, err
	}
	event := notifications.NotificationEvent{
		Type:       domain.NotificationEventServiceHealthChanged,
		Title:      "HomelabWatch test notification",
		Message:    "This is a HomelabWatch test notification.",
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		Resource:   map[string]any{"type": "notification", "id": id, "action": "test"},
		Payload:    map[string]any{},
	}
	if err := notifications.Send(ctx, channel, event); err != nil {
		delivery.Status = domain.NotificationDeliveryFailed
		delivery.Message = err.Error()
	} else {
		delivery.Status = domain.NotificationDeliverySent
		delivery.Message = "sent"
	}
	saved, err := a.store.UpdateNotificationDelivery(ctx, delivery)
	if err != nil {
		return domain.NotificationDelivery{}, err
	}
	a.publish("notification", saved.ID, "delivered", saved)
	return saved, nil
}

func (a *App) ListNotificationRules(ctx context.Context) ([]domain.NotificationRule, error) {
	return a.store.ListNotificationRules(ctx)
}

func (a *App) SaveNotificationRule(ctx context.Context, input domain.NotificationRule) (domain.NotificationRule, error) {
	item, err := a.store.SaveNotificationRule(ctx, input)
	if err != nil {
		return domain.NotificationRule{}, err
	}
	a.publish("notification", item.ID, "rule_saved", item)
	return item, nil
}

func (a *App) DeleteNotificationRule(ctx context.Context, id string) error {
	if err := a.store.DeleteNotificationRule(ctx, id); err != nil {
		return err
	}
	a.publish("notification", id, "rule_deleted", nil)
	return nil
}

func (a *App) ListNotificationDeliveries(ctx context.Context, limit int) ([]domain.NotificationDelivery, error) {
	return a.store.ListNotificationDeliveries(ctx, limit)
}
