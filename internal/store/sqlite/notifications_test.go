package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/deleema/homelabwatch/internal/domain"
)

func TestNotificationChannelCRUDRedactsAndPreservesSecrets(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	channel, err := store.SaveNotificationChannel(ctx, domain.NotificationChannel{
		Name: "Webhook",
		Type: domain.NotificationChannelWebhook,
		Config: map[string]any{
			"url":            "https://example.test/hook/token",
			"timeoutSeconds": float64(4),
		},
	})
	if err != nil {
		t.Fatalf("save channel: %v", err)
	}
	if !channel.Enabled {
		t.Fatalf("new channel disabled, want enabled by default")
	}
	if got := channel.Config["url"]; got != domain.RedactedSecret {
		t.Fatalf("webhook url = %v, want redacted", got)
	}

	channel.Name = "Renamed"
	channel.Config = map[string]any{"url": domain.RedactedSecret, "timeoutSeconds": float64(8)}
	patched, err := store.SaveNotificationChannel(ctx, channel)
	if err != nil {
		t.Fatalf("patch channel: %v", err)
	}
	if got := patched.Config["url"]; got != domain.RedactedSecret {
		t.Fatalf("patched webhook url = %v, want redacted", got)
	}
	raw, err := store.GetNotificationChannelForSend(ctx, channel.ID)
	if err != nil {
		t.Fatalf("get channel for send: %v", err)
	}
	if got := raw.Config["url"]; got != "https://example.test/hook/token" {
		t.Fatalf("raw webhook url = %v, want preserved secret", got)
	}
}

func TestNotificationRuleValidationAndDeliveryHistory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	channel, err := store.SaveNotificationChannel(ctx, domain.NotificationChannel{
		Name:   "ntfy",
		Type:   domain.NotificationChannelNtfy,
		Config: map[string]any{"serverUrl": "https://ntfy.sh", "topic": "homelabwatch", "token": "secret"},
	})
	if err != nil {
		t.Fatalf("save channel: %v", err)
	}
	if _, err := store.SaveNotificationRule(ctx, domain.NotificationRule{
		Name:       "bad",
		EventType:  domain.NotificationEventWorkerFailed,
		ChannelIDs: []string{"missing"},
	}); err == nil {
		t.Fatalf("save rule with missing channel succeeded")
	}
	rule, err := store.SaveNotificationRule(ctx, domain.NotificationRule{
		Name:       "worker failures",
		EventType:  domain.NotificationEventWorkerFailed,
		ChannelIDs: []string{channel.ID},
		Filters:    map[string]any{"minConsecutiveFailures": float64(2)},
	})
	if err != nil {
		t.Fatalf("save rule: %v", err)
	}
	delivery, err := store.CreateNotificationDelivery(ctx, domain.NotificationDelivery{
		RuleID:    rule.ID,
		ChannelID: channel.ID,
		EventType: domain.NotificationEventWorkerFailed,
		Status:    domain.NotificationDeliverySent,
		Message:   "sent",
	})
	if err != nil {
		t.Fatalf("create delivery: %v", err)
	}
	items, err := store.ListNotificationDeliveries(ctx, 0)
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	if len(items) != 1 || items[0].ID != delivery.ID || items[0].RuleName != rule.Name || items[0].ChannelName != channel.Name {
		t.Fatalf("delivery history did not include names: %#v", items)
	}
	if err := store.DeleteNotificationRule(ctx, "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("delete missing rule error = %v, want sql.ErrNoRows", err)
	}
}
