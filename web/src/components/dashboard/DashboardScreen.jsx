import DashboardHeader from "./DashboardHeader";
import ServicesSection from "./ServicesSection";
import DiscoverySection from "./DiscoverySection";
import DevicesSection from "./DevicesSection";
import BookmarksSection from "./BookmarksSection";
import WorkersSection from "./WorkersSection";

const DEFAULT_SUMMARY = {
  totalServices: 0,
  healthyServices: 0,
  degradedServices: 0,
  unhealthyServices: 0,
  devicesSeen: 0,
  bookmarks: 0,
};

export default function DashboardScreen({
  adminToken,
  dashboard,
  error,
  notice,
  onAdminTokenChange,
  onRefresh,
  onRunDiscovery,
  onRunMonitoring,
  onSaveBookmark,
  onSaveDockerEndpoint,
  onSaveManualService,
  onSaveScanTarget,
  settings,
}) {
  const summary = dashboard?.summary ?? DEFAULT_SUMMARY;
  const metrics = [
    { label: "Services", value: summary.totalServices, tone: "text-ink" },
    { label: "Healthy", value: summary.healthyServices, tone: "text-ok" },
    { label: "Degraded", value: summary.degradedServices, tone: "text-warn" },
    { label: "Unhealthy", value: summary.unhealthyServices, tone: "text-danger" },
    { label: "Devices", value: summary.devicesSeen, tone: "text-accent" },
    { label: "Bookmarks", value: summary.bookmarks, tone: "text-ink" },
  ];

  return (
    <section className="grid gap-6">
      <DashboardHeader
        adminToken={adminToken}
        adminTokenFile={settings?.appSettings?.adminTokenFile ?? ""}
        error={error}
        metrics={metrics}
        notice={notice}
        onAdminTokenChange={onAdminTokenChange}
        onRefresh={onRefresh}
        onRunDiscovery={onRunDiscovery}
        onRunMonitoring={onRunMonitoring}
      />

      <div className="grid gap-6 xl:grid-cols-[1.4fr_0.9fr]">
        <ServicesSection
          onSubmit={onSaveManualService}
          services={dashboard?.services ?? []}
        />
        <DiscoverySection
          dockerEndpoints={settings?.dockerEndpoints ?? []}
          onSaveDockerEndpoint={onSaveDockerEndpoint}
          onSaveScanTarget={onSaveScanTarget}
          scanTargets={settings?.scanTargets ?? []}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr_0.8fr]">
        <DevicesSection devices={dashboard?.devices ?? []} />
        <BookmarksSection
          bookmarks={dashboard?.bookmarks ?? []}
          onSubmit={onSaveBookmark}
        />
        <WorkersSection
          jobState={settings?.jobState ?? []}
          recentEvents={dashboard?.recentEvents ?? []}
        />
      </div>
    </section>
  );
}
