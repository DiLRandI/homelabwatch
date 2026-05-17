import { useMemo, useState } from "react";
import { Background, Controls, MiniMap, ReactFlow } from "@xyflow/react";

import EmptyState from "../ui/EmptyState";
import { buildTopologyFlowModel, TopologyNode } from "./topologyModel";

const nodeTypes = { topology: TopologyNode };

function DetailPanel({ selected, services }) {
  if (!selected) {
    return (
      <aside className="topology-detail">
        <h2 className="text-sm font-semibold text-ink">Selection</h2>
        <p className="mt-2 text-sm text-muted">Select a subnet or device to inspect addresses, utilization, ports, and linked services.</p>
      </aside>
    );
  }
  const { item, kind } = selected.data;
  if (kind === "edge") {
    return (
      <aside className="topology-detail">
        <h2 className="text-sm font-semibold text-ink">Link</h2>
        <dl className="mt-4 grid gap-3 text-sm">
          <Detail label="Kind" value={item.kind || "n/a"} />
          <Detail label="Source" value={item.source || "n/a"} />
          <Detail label="Protocol" value={item.protocol || "n/a"} />
          <Detail label="Confidence" value={item.confidence || "n/a"} />
          <Detail label="Port" value={item.label || "n/a"} />
          <Detail label="Observed" value={item.observedAt ? new Date(item.observedAt).toLocaleString() : item.inferred ? "inferred" : "n/a"} />
        </dl>
      </aside>
    );
  }
  if (kind === "infrastructure") {
    return (
      <aside className="topology-detail">
        <h2 className="text-sm font-semibold text-ink">{item.label}</h2>
        <dl className="mt-4 grid gap-3 text-sm">
          <Detail label="Management address" value={item.managementAddress || "n/a"} />
          <Detail label="Chassis ID" value={item.chassisId || "n/a"} />
          <Detail label="System name" value={item.systemName || "n/a"} />
          <Detail label="Role" value={item.role || item.kind || "unknown"} />
          <Detail label="Source" value={item.sourceId || "observed neighbor"} />
          <Detail label="Last seen" value={item.lastSeenAt ? new Date(item.lastSeenAt).toLocaleString() : "n/a"} />
        </dl>
      </aside>
    );
  }
  if (kind === "address-group") {
    return (
      <aside className="topology-detail">
        <h2 className="text-sm font-semibold text-ink">{item.name}</h2>
        <dl className="mt-4 grid gap-3 text-sm">
          <Detail label="CIDR" value={item.cidr} />
          <Detail label="Network" value={item.networkAddress || "n/a"} />
          <Detail label="Broadcast" value={item.broadcastAddress || "n/a"} />
          <Detail label="Usable range" value={`${item.firstUsableAddress || "n/a"} - ${item.lastUsableAddress || "n/a"}`} />
          <Detail label="Utilization" value={`${item.discoveredAddressCount || 0}/${item.usableAddressCount || 0} addresses (${item.utilizationPct || 0}%)`} />
          <Detail label="Services" value={item.serviceCount || 0} />
        </dl>
      </aside>
    );
  }
  if (kind === "subnet") {
    return (
      <aside className="topology-detail">
        <h2 className="text-sm font-semibold text-ink">{item.name}</h2>
        <dl className="mt-4 grid gap-3 text-sm">
          <Detail label="CIDR" value={item.cidr || "Unmapped"} />
          <Detail label="Network" value={item.networkAddress || "n/a"} />
          <Detail label="Broadcast" value={item.broadcastAddress || "n/a"} />
          <Detail label="Usable range" value={`${item.firstUsableAddress || "n/a"} - ${item.lastUsableAddress || "n/a"}`} />
          <Detail label="Gateway" value={item.gatewayAddress ? `${item.gatewayAddress}${item.gatewayInferred ? " (inferred)" : ""}` : "n/a"} />
          <Detail label="Utilization" value={`${item.discoveredAddressCount || 0}/${item.usableAddressCount || 0} addresses (${item.utilizationPct || 0}%)`} />
        </dl>
        {item.warnings?.length ? <Warnings warnings={item.warnings} /> : null}
      </aside>
    );
  }
  if (kind === "device") {
    const linkedServices = services.filter((service) => service.deviceId === item.id);
    return (
      <aside className="topology-detail">
        <h2 className="text-sm font-semibold text-ink">{item.label}</h2>
        <dl className="mt-4 grid gap-3 text-sm">
          <Detail label="Primary address" value={item.primaryAddress || "n/a"} />
          <Detail label="Addresses" value={item.addresses?.join(", ") || "n/a"} />
          <Detail label="MAC" value={item.primaryMac || "n/a"} />
          <Detail label="Open ports" value={item.openPorts?.join(", ") || "None"} />
          <Detail label="Confidence" value={item.identityConfidence || "unknown"} />
          <Detail label="Last seen" value={item.lastSeenAt ? new Date(item.lastSeenAt).toLocaleString() : "n/a"} />
        </dl>
        <h3 className="mt-5 text-xs font-semibold uppercase tracking-[0.08em] text-copy-subtle">Services</h3>
        <div className="mt-2 grid gap-2">
          {linkedServices.length ? linkedServices.map((service) => (
            <div className="rounded-md border border-line bg-panel-soft px-3 py-2" key={service.id}>
              <div className="text-sm font-semibold text-ink">{service.name}</div>
              <div className="font-mono text-xs text-muted">{service.host}:{service.port}</div>
            </div>
          )) : <p className="text-sm text-muted">No linked services.</p>}
        </div>
      </aside>
    );
  }
  return (
    <aside className="topology-detail">
      <h2 className="text-sm font-semibold text-ink">{item.label}</h2>
      <p className="mt-2 text-sm text-muted">{item.address || "Gateway address inferred from the subnet."}</p>
    </aside>
  );
}

function Detail({ label, value }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-[0.08em] text-copy-subtle">{label}</dt>
      <dd className="mt-1 break-words text-ink-soft">{value}</dd>
    </div>
  );
}

function Warnings({ warnings }) {
  return (
    <div className="mt-5 rounded-md border border-warn/30 bg-warn/10 p-3 text-sm text-warn-strong">
      {warnings.map((warning) => <div key={warning}>{warning}</div>)}
    </div>
  );
}

export default function TopologyScreen({ topology }) {
  const [mode, setMode] = useState("auto");
  const [showCrossLinks, setShowCrossLinks] = useState(false);
  const [showHidden, setShowHidden] = useState(false);
  const [showInferredFallback, setShowInferredFallback] = useState(false);
  const [selected, setSelected] = useState(null);
  const model = useMemo(
    () => buildTopologyFlowModel(topology, { mode, showCrossLinks, showHidden, showInferredFallback }),
    [mode, showCrossLinks, showHidden, showInferredFallback, topology],
  );
  const isEmpty = !topology || ((topology.subnets?.length || 0) === 0 && (topology.devices?.length || 0) === 0);

  if (isEmpty) {
    return <EmptyState body="Topology appears after scan targets and LAN discovery create devices." title="No network topology yet" />;
  }

  return (
    <section className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
      <div className="overflow-hidden rounded-lg border border-line bg-panel shadow-card">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-4 py-3">
          <div>
            <h1 className="text-base font-semibold text-ink">Network topology</h1>
            <p className="text-sm text-muted">
              {topology.summary?.routerCount || 0} gateways · {topology.summary?.subnetCount || 0} subnets · {topology.infrastructureNodes?.length || 0} infrastructure · {topology.summary?.deviceCount || 0} devices
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <label className="inline-flex items-center gap-2 text-sm text-ink-soft">
              Mode
              <select
                className="rounded-xl border border-line bg-panel-strong px-3 py-2 text-sm text-ink outline-hidden focus:border-accent focus-visible:ring-4 focus-visible:ring-accent/15"
                onChange={(event) => setMode(event.target.value)}
                value={mode}
              >
                <option value="auto">Auto</option>
                <option value="observed">Observed</option>
                <option value="address">Address</option>
              </select>
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-ink-soft">
              <input className="h-4 w-4 accent-accent" checked={showHidden} onChange={(event) => setShowHidden(event.target.checked)} type="checkbox" />
              Show hidden devices
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-ink-soft">
              <input className="h-4 w-4 accent-accent" checked={showCrossLinks} onChange={(event) => setShowCrossLinks(event.target.checked)} type="checkbox" />
              Show cross-links
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-ink-soft">
              <input className="h-4 w-4 accent-accent" checked={showInferredFallback} onChange={(event) => setShowInferredFallback(event.target.checked)} type="checkbox" />
              Show inferred fallback links
            </label>
          </div>
        </div>
        <div className="h-[68vh] min-h-[560px]">
          <ReactFlow
            className="topology-flow"
            edges={model.edges}
            fitView
            nodes={model.nodes}
            nodeTypes={nodeTypes}
            onEdgeClick={(_, edge) => setSelected({ data: edge.data })}
            onNodeClick={(_, node) => setSelected(node)}
            onPaneClick={() => setSelected(null)}
          >
            <Background gap={24} size={1} />
            <Controls />
            <MiniMap
              bgColor="var(--color-panel-strong)"
              maskColor="rgba(2, 6, 23, 0.48)"
              nodeBorderRadius={8}
              nodeColor="var(--color-panel-soft)"
              nodeStrokeColor="var(--color-line-strong)"
              pannable
              zoomable
            />
          </ReactFlow>
        </div>
      </div>
      <DetailPanel selected={selected} services={topology.services ?? []} />
      {topology.warnings?.length ? <div className="xl:col-span-2"><Warnings warnings={topology.warnings} /></div> : null}
    </section>
  );
}
