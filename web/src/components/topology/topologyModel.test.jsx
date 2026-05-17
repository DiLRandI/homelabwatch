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
});
