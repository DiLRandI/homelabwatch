import DockerEndpointForm from "../forms/DockerEndpointForm";
import ScanTargetForm from "../forms/ScanTargetForm";
import CardList from "../ui/CardList";
import Section from "../ui/Section";

export default function DiscoverySection({
  dockerEndpoints,
  onSaveDockerEndpoint,
  onSaveScanTarget,
  scanTargets,
}) {
  return (
    <Section
      title="Discovery"
      subtitle="Seed Docker engines and scan targets without leaving the dashboard."
    >
      <div className="grid gap-4">
        <CardList
          items={dockerEndpoints}
          renderItem={(item) => (
            <div
              key={item.id}
              className="rounded-3xl border border-white/10 bg-base/70 p-4"
            >
              <p className="font-semibold text-ink">{item.name}</p>
              <p className="mt-1 text-sm text-muted">{item.address}</p>
              <p className="mt-2 text-xs uppercase tracking-[0.2em] text-muted/80">
                {item.enabled ? "enabled" : "disabled"} • every{" "}
                {item.scanIntervalSeconds}s
              </p>
            </div>
          )}
          title="Docker endpoints"
        />
        <DockerEndpointForm onSubmit={onSaveDockerEndpoint} />
        <CardList
          items={scanTargets}
          renderItem={(item) => (
            <div
              key={item.id}
              className="rounded-3xl border border-white/10 bg-base/70 p-4"
            >
              <div className="flex items-center justify-between gap-4">
                <p className="font-semibold text-ink">{item.name}</p>
                <span className="text-xs uppercase tracking-[0.2em] text-muted">
                  {item.autoDetected ? "auto" : "manual"}
                </span>
              </div>
              <p className="mt-1 text-sm text-muted">{item.cidr}</p>
              <p className="mt-2 text-xs uppercase tracking-[0.2em] text-muted/80">
                ports {item.commonPorts.join(", ")} • every{" "}
                {item.scanIntervalSeconds}s
              </p>
            </div>
          )}
          title="Scan targets"
        />
        <ScanTargetForm onSubmit={onSaveScanTarget} />
      </div>
    </Section>
  );
}
