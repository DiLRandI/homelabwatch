package app

import (
	"context"
	"testing"
	"time"

	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/domain"
)

func TestStatusPageAppPublishesEvents(t *testing.T) {
	application, _, _ := newTestApp(t, config.Config{DefaultScanPorts: []int{22, 80}})
	ctx := context.Background()
	if err := application.Setup(ctx, domain.SetupInput{ApplianceName: "Lab", DefaultScanPorts: []int{22, 80}}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	events := application.SubscribeEvents(10)
	defer application.UnsubscribeEvents(events)

	page, err := application.SaveStatusPage(ctx, domain.StatusPageInput{Slug: "ops", Title: "Ops"})
	if err != nil {
		t.Fatalf("save page: %v", err)
	}
	expectEvent(t, events, "status-page", page.ID, "upserted")

	if err := application.DeleteStatusPage(ctx, page.ID); err != nil {
		t.Fatalf("delete page: %v", err)
	}
	expectEvent(t, events, "status-page", page.ID, "deleted")
}

func expectEvent(t *testing.T, events <-chan domain.EventEnvelope, eventType, id, action string) {
	t.Helper()
	select {
	case event := <-events:
		if event.Type != eventType || event.ID != id || event.Action != action {
			t.Fatalf("event = %+v, want %s %s %s", event, eventType, id, action)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for event %s %s %s", eventType, id, action)
	}
}
