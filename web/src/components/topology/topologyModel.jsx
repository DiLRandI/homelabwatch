import dagre from "@dagrejs/dagre";

const NODE_WIDTH = 210;
const NODE_HEIGHT = 86;

function labelForKind(kind) {
  if (kind === "router") return "Gateway";
  if (kind === "subnet") return "Subnet";
  return "Device";
}

export function buildTopologyFlowModel(topology, { showHidden = false } = {}) {
  const graph = new dagre.graphlib.Graph();
  graph.setDefaultEdgeLabel(() => ({}));
  graph.setGraph({ marginx: 30, marginy: 30, nodesep: 54, rankdir: "TB", ranksep: 86 });

  const nodes = [];
  const visibleDeviceIds = new Set();
  const addNode = (node) => {
    nodes.push(node);
    graph.setNode(node.id, { height: NODE_HEIGHT, width: NODE_WIDTH });
  };

  for (const router of topology?.routers ?? []) {
    addNode({
      id: router.id,
      data: { item: router, kind: "router", label: router.label || "Inferred gateway" },
      position: { x: 0, y: 0 },
      type: "topology",
    });
  }

  for (const subnet of topology?.subnets ?? []) {
    addNode({
      id: subnet.id,
      data: { item: subnet, kind: "subnet", label: subnet.name || subnet.cidr || "Subnet" },
      position: { x: 0, y: 0 },
      type: "topology",
    });
  }

  for (const device of topology?.devices ?? []) {
    if (device.hidden && !showHidden) continue;
    visibleDeviceIds.add(device.id);
    addNode({
      id: device.id,
      data: { item: device, kind: "device", label: device.label || device.primaryAddress || "Device" },
      position: { x: 0, y: 0 },
      type: "topology",
    });
  }

  const nodeIds = new Set(nodes.map((node) => node.id));
  const edges = (topology?.edges ?? [])
    .filter((edge) => nodeIds.has(edge.sourceId) && nodeIds.has(edge.targetId))
    .map((edge) => ({
      id: edge.id,
      source: edge.sourceId,
      target: edge.targetId,
      type: "smoothstep",
      label: edge.kind === "subnet-subnet" ? "contains" : "",
      className: `topology-edge topology-edge--${edge.kind}`,
    }));

  for (const edge of edges) {
    graph.setEdge(edge.source, edge.target);
  }

  dagre.layout(graph);
  for (const node of nodes) {
    const position = graph.node(node.id);
    node.position = {
      x: (position?.x ?? 0) - NODE_WIDTH / 2,
      y: (position?.y ?? 0) - NODE_HEIGHT / 2,
    };
  }

  return {
    edges,
    hiddenDeviceCount: (topology?.devices ?? []).filter((device) => device.hidden && !visibleDeviceIds.has(device.id)).length,
    nodes,
  };
}

export function TopologyNode({ data }) {
  const item = data.item;
  const kind = data.kind;
  const details =
    kind === "subnet"
      ? item.cidr || item.family
      : kind === "device"
        ? item.primaryAddress || item.primaryMac || "No address"
        : item.address || "Inferred";

  return (
    <div className={`topology-node topology-node--${kind}`}>
      <div className="text-[0.68rem] font-semibold uppercase tracking-[0.08em] text-copy-subtle">
        {labelForKind(kind)}
      </div>
      <div className="mt-1 truncate text-sm font-semibold text-ink" title={data.label}>
        {data.label}
      </div>
      <div className="mt-1 truncate font-mono text-xs text-muted" title={details}>
        {details}
      </div>
      {kind === "device" ? (
        <div className="mt-2 flex gap-2 text-[0.7rem] text-ink-soft">
          <span>{item.serviceCount || 0} svc</span>
          <span>{item.openPorts?.length || 0} ports</span>
          <span>{item.identityConfidence || "unknown"}</span>
        </div>
      ) : null}
      {kind === "subnet" ? (
        <div className="mt-2 text-[0.7rem] text-ink-soft">
          {item.discoveredDeviceCount || 0} devices · {item.utilizationPct || 0}%
        </div>
      ) : null}
    </div>
  );
}
