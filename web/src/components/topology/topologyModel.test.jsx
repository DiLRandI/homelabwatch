import { describe, expect, it } from "vitest";

import { buildTopologyFlowModel } from "./topologyModel";

describe("buildTopologyFlowModel", () => {
  it("converts topology records into visible React Flow nodes and edges", () => {
    const model = buildTopologyFlowModel({
      routers: [{ id: "router:subnet:a", label: "Inferred gateway" }],
      subnets: [{ id: "subnet:a", name: "LAN", cidr: "192.168.1.0/24" }],
      devices: [
        { id: "device:1", subnetId: "subnet:a", label: "nas", primaryAddress: "192.168.1.10" },
        { id: "device:2", subnetId: "subnet:a", hidden: true, label: "hidden" },
      ],
      edges: [
        { id: "router:subnet:a->subnet:a", sourceId: "router:subnet:a", targetId: "subnet:a", kind: "router-subnet" },
        { id: "subnet:a->device:1", sourceId: "subnet:a", targetId: "device:1", kind: "subnet-device" },
        { id: "subnet:a->device:2", sourceId: "subnet:a", targetId: "device:2", kind: "subnet-device" },
      ],
    });

    expect(model.nodes.map((node) => node.id)).toEqual([
      "router:subnet:a",
      "subnet:a",
      "device:1",
    ]);
    expect(model.edges.map((edge) => edge.id)).toEqual([
      "router:subnet:a->subnet:a",
      "subnet:a->device:1",
    ]);
    expect(model.hiddenDeviceCount).toBe(1);
  });

  it("renders infrastructure and address-group nodes", () => {
    const model = buildTopologyFlowModel({
      infrastructureNodes: [{ id: "infra:core", label: "core", managementAddress: "192.168.1.2" }],
      addressGroups: [{ id: "addrgrp:1", name: "192.168.1.0/26", cidr: "192.168.1.0/26" }],
      devices: [{ id: "device:1", label: "nas" }],
      edges: [
        { id: "infra-device", sourceId: "infra:core", targetId: "device:1", kind: "infrastructure-port-device", source: "bridge", protocol: "snmp" },
      ],
    });

    expect(model.nodes.map((node) => node.id)).toEqual(["infra:core", "addrgrp:1", "device:1"]);
    expect(model.edges.map((edge) => edge.id)).toEqual(["infra-device"]);
  });

  it("filters inferred fallback links in observed and auto modes", () => {
    const topology = {
      infrastructureNodes: [{ id: "infra:core", label: "core" }],
      subnets: [{ id: "subnet:a", name: "LAN" }],
      devices: [{ id: "device:1", label: "nas" }, { id: "device:2", label: "printer" }],
      edges: [
        { id: "observed", sourceId: "infra:core", targetId: "device:1", kind: "infrastructure-port-device", source: "bridge", protocol: "snmp" },
        { id: "fallback-1", sourceId: "subnet:a", targetId: "device:1", kind: "subnet-device", source: "address", inferred: true },
        { id: "fallback-2", sourceId: "subnet:a", targetId: "device:2", kind: "subnet-device", source: "address", inferred: true },
      ],
    };

    expect(buildTopologyFlowModel(topology, { mode: "observed" }).edges.map((edge) => edge.id)).toEqual(["observed"]);
    expect(buildTopologyFlowModel(topology, { mode: "auto" }).edges.map((edge) => edge.id)).toEqual(["observed", "fallback-2"]);
    expect(buildTopologyFlowModel(topology, { mode: "auto", showInferredFallback: true }).edges.map((edge) => edge.id)).toEqual(["observed", "fallback-1", "fallback-2"]);
  });

  it("hides cross-links by default", () => {
    const topology = {
      infrastructureNodes: [{ id: "infra:a", label: "a" }, { id: "infra:b", label: "b" }],
      edges: [
        { id: "tree", sourceId: "infra:a", targetId: "infra:b", kind: "infrastructure-link", source: "lldp" },
        { id: "cross", sourceId: "infra:b", targetId: "infra:a", kind: "cross-link", source: "lldp" },
      ],
    };

    expect(buildTopologyFlowModel(topology).edges.map((edge) => edge.id)).toEqual(["tree"]);
    expect(buildTopologyFlowModel(topology, { showCrossLinks: true }).edges.map((edge) => edge.id)).toEqual(["tree", "cross"]);
  });
});
