import Alerts from "../ui/Alerts";
import ActionButton from "../ui/ActionButton";
import Input from "../ui/Input";
import MetricCard from "../ui/MetricCard";

export default function DashboardHeader({
  adminToken,
  error,
  metrics,
  notice,
  onAdminTokenChange,
  onRefresh,
  onRunDiscovery,
  onRunMonitoring,
}) {
  return (
    <header className="animate-floatIn rounded-4xl border border-white/10 bg-panel/80 p-6 shadow-halo backdrop-blur-sm">
      <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p className="text-sm uppercase tracking-[0.35em] text-accent">
            Homelabwatch
          </p>
          <h1 className="mt-2 font-display text-4xl font-semibold text-ink">
            Discover, monitor, and reach everything in the lab.
          </h1>
          <p className="mt-3 max-w-3xl text-sm leading-7 text-muted">
            The dashboard tracks devices by MAC identity, discovers Docker
            workloads and LAN services, and streams health changes over a single
            embedded control plane.
          </p>
        </div>
        <div className="grid gap-3 sm:grid-cols-[minmax(0,18rem)_auto_auto_auto]">
          <Input
            compact
            label="Admin token"
            onChange={onAdminTokenChange}
            placeholder="required for writes"
            type="password"
            value={adminToken}
          />
          <ActionButton onClick={() => void onRunDiscovery()}>
            Run discovery
          </ActionButton>
          <ActionButton onClick={() => void onRunMonitoring()}>
            Run checks
          </ActionButton>
          <ActionButton onClick={() => void onRefresh()}>Refresh</ActionButton>
        </div>
      </div>
      <Alerts error={error} notice={notice} />
      <div className="mt-6 grid gap-4 md:grid-cols-3 xl:grid-cols-6">
        {metrics.map((metric) => (
          <MetricCard key={metric.label} {...metric} />
        ))}
      </div>
    </header>
  );
}
