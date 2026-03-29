import DashboardLayout from "../components/layout/DashboardLayout";
import Alerts from "../components/ui/Alerts";
import Button from "../components/ui/Button";
import { DiscoveryIcon, RefreshIcon, ShieldIcon } from "../components/ui/Icons";
import { APP_ROUTES } from "./routes";

const DEFAULT_SUMMARY = {
  bookmarks: 0,
  degradedServices: 0,
  devicesSeen: 0,
  discoveredServices: 0,
  healthyServices: 0,
  runningContainers: 0,
  totalServices: 0,
  unhealthyServices: 0,
};

export default function AppShell({
  activeRoute,
  bookmarks = [],
  canManageUI,
  children,
  dashboard,
  error,
  notice,
  onNavigate,
  onRefresh,
  onRunDiscovery,
  onRunMonitoring,
  settings,
}) {
  const summary = dashboard?.summary ?? DEFAULT_SUMMARY;
  const issuesCount = summary.degradedServices + summary.unhealthyServices;
  const navCounts = {
    bookmarks: bookmarks.length,
    definitions: settings?.serviceDefinitions?.length ?? 0,
    devices: summary.devicesSeen,
    discovery: summary.discoveredServices + (settings?.dockerEndpoints?.length ?? 0),
    health: issuesCount,
    services: summary.totalServices,
    settings: settings?.apiAccess?.tokens?.length ?? 0,
  };

  const navItems = APP_ROUTES.map((route) => ({
    count: route.countKey === "dashboard" ? null : navCounts[route.countKey] ?? null,
    href: route.path,
    icon: route.icon,
    label: route.label,
  }));

  const statusItems = [
    {
      className: "border-line bg-panel-strong text-muted",
      icon: DiscoveryIcon,
      label: "Realtime updates",
    },
    {
      className:
        issuesCount > 0
          ? "border-warn/20 bg-warn/10 text-warn-strong"
          : "border-ok/15 bg-ok/10 text-ok-strong",
      icon: ShieldIcon,
      label:
        issuesCount > 0
          ? `${issuesCount} services need attention`
          : "Service health clear",
    },
    {
      className: canManageUI
        ? "border-accent/15 bg-accent/10 text-accent-strong"
        : "border-line bg-panel-strong text-muted",
      icon: ShieldIcon,
      label: canManageUI ? "Trusted LAN writes enabled" : "Read-only network",
    },
  ];

  return (
    <DashboardLayout
      activeHref={activeRoute.path}
      alerts={<Alerts error={error} notice={notice} />}
      navItems={navItems}
      onNavigate={onNavigate}
      sidebarMeta={{
        applianceName: settings?.appSettings?.applianceName,
      }}
      statusItems={statusItems}
      subtitle={activeRoute.subtitle}
      title={activeRoute.title}
      toolbar={
        <div className="grid gap-3 xl:grid-cols-[repeat(3,auto)]">
          <Button
            disabled={!canManageUI}
            leadingIcon={DiscoveryIcon}
            onClick={() => void onRunDiscovery()}
            variant="secondary"
          >
            Run discovery
          </Button>
          <Button
            disabled={!canManageUI}
            leadingIcon={ShieldIcon}
            onClick={() => void onRunMonitoring()}
            variant="secondary"
          >
            Run checks
          </Button>
          <Button
            leadingIcon={RefreshIcon}
            onClick={() => void onRefresh()}
            variant="ghost"
          >
            Refresh
          </Button>
        </div>
      }
    >
      {children}
    </DashboardLayout>
  );
}
