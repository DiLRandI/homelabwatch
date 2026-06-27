package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

func TestTopologySourceCRUDRedactsSecretsAndClearsExplicitEmpty(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	saved, err := store.SaveTopologySource(ctx, domain.TopologySource{
		Name:              "Core",
		Address:           "192.168.1.2",
		Enabled:           true,
		SNMPVersion:       "v3",
		Community:         "public",
		Username:          "snmpuser",
		AuthProtocol:      "sha256",
		AuthPassphrase:    "authsecret",
		PrivacyProtocol:   "aes",
		PrivacyPassphrase: "privsecret",
		Role:              "switch",
		Root:              true,
	})
	if err != nil {
		t.Fatalf("save topology source: %v", err)
	}

	items, err := store.ListTopologySources(ctx)
	if err != nil {
		t.Fatalf("list topology sources: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one source, got %#v", items)
	}
	if items[0].Community != "" || items[0].AuthPassphrase != "" || items[0].PrivacyPassphrase != "" {
		t.Fatalf("expected redacted secrets, got %#v", items[0])
	}
	if !items[0].HasCommunity || !items[0].HasAuthPassphrase || !items[0].HasPrivacyPassphrase {
		t.Fatalf("expected credential flags, got %#v", items[0])
	}

	discovery, err := store.GetTopologySourceForDiscovery(ctx, saved.ID)
	if err != nil {
		t.Fatalf("get source for discovery: %v", err)
	}
	if discovery.Community != "public" || discovery.AuthPassphrase != "authsecret" || discovery.PrivacyPassphrase != "privsecret" {
		t.Fatalf("expected discovery secrets, got %#v", discovery)
	}

	discovery.Community = ""
	discovery.AuthPassphrase = ""
	discovery.PrivacyPassphrase = ""
	if _, err := store.SaveTopologySource(ctx, discovery); err != nil {
		t.Fatalf("clear source secrets: %v", err)
	}
	cleared, err := store.ListTopologySources(ctx)
	if err != nil {
		t.Fatalf("list cleared source: %v", err)
	}
	if cleared[0].HasCommunity || cleared[0].HasAuthPassphrase || cleared[0].HasPrivacyPassphrase {
		t.Fatalf("expected cleared credential flags, got %#v", cleared[0])
	}

	if err := store.DeleteTopologySource(ctx, saved.ID); err != nil {
		t.Fatalf("delete source: %v", err)
	}
	afterDelete, err := store.ListTopologySources(ctx)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(afterDelete) != 0 {
		t.Fatalf("expected no sources, got %#v", afterDelete)
	}
}

func TestReplaceTopologyObservationsScopesRowsBySource(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	sourceA, err := store.SaveTopologySource(ctx, domain.TopologySource{Name: "A", Address: "192.168.1.2", Enabled: true, SNMPVersion: "v2c", Community: "public"})
	if err != nil {
		t.Fatalf("save source a: %v", err)
	}
	sourceB, err := store.SaveTopologySource(ctx, domain.TopologySource{Name: "B", Address: "192.168.1.3", Enabled: true, SNMPVersion: "v2c", Community: "public"})
	if err != nil {
		t.Fatalf("save source b: %v", err)
	}
	seen := time.Now().UTC()
	if err := store.ReplaceTopologyObservations(ctx, sourceA.ID, domain.TopologySourceObservation{
		ObservedAt: seen,
		Interfaces: []domain.TopologyInterfaceObservation{{
			IfIndex:       1,
			IfName:        "gi1",
			IfDescription: "uplink",
		}},
		MACLinks: []domain.TopologyMACLinkObservation{{
			MACAddress: "aa:bb:cc:dd:ee:ff",
			IfIndex:    1,
		}},
	}); err != nil {
		t.Fatalf("replace source a observations: %v", err)
	}
	if err := store.ReplaceTopologyObservations(ctx, sourceB.ID, domain.TopologySourceObservation{
		ObservedAt: seen,
		LLDPLinks: []domain.TopologyLLDPLinkObservation{{
			LocalChassisID:  "b",
			LocalPortID:     "gi1",
			RemoteChassisID: "a",
			RemotePortID:    "gi1",
		}},
	}); err != nil {
		t.Fatalf("replace source b observations: %v", err)
	}
	if err := store.ReplaceTopologyObservations(ctx, sourceA.ID, domain.TopologySourceObservation{
		ObservedAt: seen,
		Interfaces: []domain.TopologyInterfaceObservation{{
			IfIndex: 2,
			IfName:  "gi2",
		}},
	}); err != nil {
		t.Fatalf("replace stale source a observations: %v", err)
	}
	observations, err := store.ListTopologyObservations(ctx)
	if err != nil {
		t.Fatalf("list observations: %v", err)
	}
	if len(observations) != 2 {
		t.Fatalf("expected two source observations, got %#v", observations)
	}
	if observations[0].SourceID != sourceA.ID || len(observations[0].Interfaces) != 1 || observations[0].Interfaces[0].IfIndex != 2 || len(observations[0].MACLinks) != 0 {
		t.Fatalf("expected source a stale rows replaced, got %#v", observations[0])
	}
	if observations[1].SourceID != sourceB.ID || len(observations[1].LLDPLinks) != 1 {
		t.Fatalf("expected source b rows preserved, got %#v", observations[1])
	}
}
