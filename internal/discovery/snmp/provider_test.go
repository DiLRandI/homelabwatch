package snmp

import (
	"context"
	"errors"
	"testing"
	"time"

	g "github.com/gosnmp/gosnmp"

	"github.com/deleema/homelabwatch/internal/domain"
)

type fakeSession struct {
	errs  map[string]error
	walks map[string][]g.SnmpPDU
}

func (s fakeSession) WalkAll(oid string) ([]g.SnmpPDU, error) {
	if err := s.errs[oid]; err != nil {
		return nil, err
	}
	return s.walks[oid], nil
}

func (s fakeSession) Close() error { return nil }

func TestPollInterfacesParsesNamesDescriptionsAndStatus(t *testing.T) {
	seen := time.Now().UTC()
	items, err := pollInterfaces(fakeSession{walks: map[string][]g.SnmpPDU{
		oidIfName:       {{Name: oidIfName + ".10", Value: []byte("gi1")}},
		oidIfDescr:      {{Name: oidIfDescr + ".10", Value: []byte("Uplink")}},
		oidIfAlias:      {{Name: oidIfAlias + ".10", Value: "Rack uplink"}},
		oidIfOperStatus: {{Name: oidIfOperStatus + ".10", Value: 1}},
		oidIfHighSpeed:  {{Name: oidIfHighSpeed + ".10", Value: 1000}},
	}}, "src1", seen)
	if err != nil {
		t.Fatalf("poll interfaces: %v", err)
	}
	if len(items) != 1 || items[0].IfName != "gi1" || items[0].IfDescription != "Uplink" || items[0].IfAlias != "Rack uplink" || items[0].OperStatus != "up" || items[0].SpeedBPS != 1_000_000_000 {
		t.Fatalf("unexpected interfaces: %#v", items)
	}
}

func TestPollLLDPParsesRemoteNeighbors(t *testing.T) {
	seen := time.Now().UTC()
	items, err := pollLLDP(fakeSession{walks: map[string][]g.SnmpPDU{
		oidLLDPLocChassisID: {{Name: oidLLDPLocChassisID, Value: []byte{0, 1, 2, 3, 4, 5}}},
		oidLLDPLocSysName:   {{Name: oidLLDPLocSysName, Value: "core"}},
		oidLLDPLocPortID:    {{Name: oidLLDPLocPortID + ".7", Value: "gi7"}},
		oidLLDPLocPortDesc:  {{Name: oidLLDPLocPortDesc + ".7", Value: "uplink"}},
		oidLLDPRemChassisID: {{Name: oidLLDPRemChassisID + ".0.7.1", Value: []byte{10, 11, 12, 13, 14, 15}}},
		oidLLDPRemPortID:    {{Name: oidLLDPRemPortID + ".0.7.1", Value: "gi48"}},
		oidLLDPRemPortDesc:  {{Name: oidLLDPRemPortDesc + ".0.7.1", Value: "downlink"}},
		oidLLDPRemSysName:   {{Name: oidLLDPRemSysName + ".0.7.1", Value: "access"}},
	}}, "src1", seen, []domain.TopologyInterfaceObservation{{IfIndex: 7, IfName: "gi7"}})
	if err != nil {
		t.Fatalf("poll lldp: %v", err)
	}
	if len(items) != 1 || items[0].LocalChassisID != "00:01:02:03:04:05" || items[0].LocalPortID != "gi7" || items[0].LocalIfIndex != 7 || items[0].RemoteChassisID != "0a:0b:0c:0d:0e:0f" || items[0].RemoteSystemName != "access" {
		t.Fatalf("unexpected lldp links: %#v", items)
	}
}

func TestPollQBridgeParsesVLANAndMAC(t *testing.T) {
	seen := time.Now().UTC()
	items, err := pollQBridge(fakeSession{walks: map[string][]g.SnmpPDU{
		oidBasePortIfIndex:  {{Name: oidBasePortIfIndex + ".5", Value: 10}},
		oidQBridgeFDBPort:   {{Name: oidQBridgeFDBPort + ".20.170.187.204.221.238.255", Value: 5}},
		oidQBridgeFDBStatus: {{Name: oidQBridgeFDBStatus + ".20.170.187.204.221.238.255", Value: 3}},
	}}, "src1", seen, []domain.TopologyInterfaceObservation{{IfIndex: 10, IfName: "gi10", IfDescription: "edge"}})
	if err != nil {
		t.Fatalf("poll q bridge: %v", err)
	}
	if len(items) != 1 || items[0].VLAN != 20 || items[0].MACAddress != "aa:bb:cc:dd:ee:ff" || items[0].IfIndex != 10 || items[0].IfName != "gi10" || items[0].Status != "learned" {
		t.Fatalf("unexpected q bridge links: %#v", items)
	}
}

func TestDiscoverFallsBackToBridgeAndReturnsPartialObservations(t *testing.T) {
	provider := &Provider{
		now: func() time.Time { return time.Unix(100, 0).UTC() },
		open: func(context.Context, domain.TopologySource) (session, error) {
			return fakeSession{
				errs: map[string]error{
					oidLLDPRemChassisID: errors.New("lldp unavailable"),
					oidLLDPRemPortID:    errors.New("lldp unavailable"),
					oidLLDPRemPortDesc:  errors.New("lldp unavailable"),
					oidLLDPRemSysName:   errors.New("lldp unavailable"),
					oidQBridgeFDBPort:   errors.New("q bridge unavailable"),
				},
				walks: map[string][]g.SnmpPDU{
					oidIfName:          {{Name: oidIfName + ".3", Value: "eth3"}},
					oidBasePortIfIndex: {{Name: oidBasePortIfIndex + ".9", Value: 3}},
					oidBridgeFDBPort:   {{Name: oidBridgeFDBPort + ".1.2.3.4.5.6", Value: 9}},
					oidBridgeFDBStatus: {{Name: oidBridgeFDBStatus + ".1.2.3.4.5.6", Value: 3}},
				},
			}, nil
		},
	}
	results := provider.Discover(context.Background(), []domain.TopologySource{{ID: "src1", Name: "Core", Address: "192.168.1.2", Enabled: true, SNMPVersion: "v2c"}})
	if len(results) != 1 || results[0].Error != nil {
		t.Fatalf("expected partial success, got %#v", results)
	}
	if len(results[0].Observation.Interfaces) != 1 || len(results[0].Observation.MACLinks) != 1 || results[0].Observation.MACLinks[0].MACAddress != "01:02:03:04:05:06" {
		t.Fatalf("expected interface and bridge fallback observations, got %#v", results[0].Observation)
	}
}

func TestDiscoverReportsConnectFailure(t *testing.T) {
	provider := &Provider{
		now: func() time.Time { return time.Now().UTC() },
		open: func(context.Context, domain.TopologySource) (session, error) {
			return nil, errors.New("authentication failed")
		},
	}
	results := provider.Discover(context.Background(), []domain.TopologySource{{ID: "src1", Enabled: true}})
	if len(results) != 1 || results[0].Error == nil {
		t.Fatalf("expected source error, got %#v", results)
	}
}
