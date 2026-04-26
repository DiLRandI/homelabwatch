import AppShell from "./app/AppShell";
import { useAppRoute } from "./app/useAppRoute";
import BookmarksScreen from "./app/screens/BookmarksScreen";
import DefinitionsScreen from "./app/screens/DefinitionsScreen";
import DevicesScreen from "./app/screens/DevicesScreen";
import DiscoveryScreen from "./app/screens/DiscoveryScreen";
import HealthScreen from "./app/screens/HealthScreen";
import NotificationsScreen from "./app/screens/NotificationsScreen";
import OverviewScreen from "./app/screens/OverviewScreen";
import ServicesScreen from "./app/screens/ServicesScreen";
import SettingsScreen from "./app/screens/SettingsScreen";
import StatusPagesScreen from "./app/screens/StatusPagesScreen";
import BootstrapScreen from "./components/bootstrap/BootstrapScreen";
import PublicStatusPageScreen from "./components/status-pages/PublicStatusPageScreen";
import {
  DevicesIcon,
  DiscoveryIcon,
  ServicesIcon,
  ShieldIcon,
  SparklesIcon,
} from "./components/ui/Icons";
import EmptyState from "./components/ui/EmptyState";
import Shell from "./components/ui/Shell";
import { useHomelabwatchApp } from "./hooks/useHomelabwatchApp";
import { usePublicStatusPage } from "./hooks/usePublicStatusPage";
import { useThemePreference } from "./hooks/useThemePreference";
import { isPublicStatusPath, statusSlugFromPath } from "./app/routes";

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

function buildMetrics(summary) {
  return [
    {
      description: "Tracked endpoints across all discovery sources.",
      icon: ServicesIcon,
      iconTone: "bg-accent/10 text-accent-strong",
      label: "Services",
      value: summary.totalServices,
    },
    {
      description: "Passing checks and responding to requests.",
      icon: ShieldIcon,
      iconTone: "bg-ok/10 text-ok-strong",
      label: "Healthy",
      value: summary.healthyServices,
    },
    {
      description: "Running workloads discovered from attached Docker engines.",
      icon: DiscoveryIcon,
      iconTone: "bg-sky-500/12 text-sky-300",
      label: "Containers",
      value: summary.runningContainers,
    },
    {
      description: "Devices known to the control plane inventory.",
      icon: DevicesIcon,
      iconTone: "bg-base text-ink-soft",
      label: "Devices",
      value: summary.devicesSeen,
    },
    {
      description: "Pending discovery suggestions waiting for review.",
      icon: SparklesIcon,
      iconTone: "bg-amber-500/12 text-amber-300",
      label: "Discovered",
      value: summary.discoveredServices,
    },
  ];
}

function countByDevice(items = []) {
  return items.reduce((counts, item) => {
    if (!item.deviceId) {
      return counts;
    }

    return {
      ...counts,
      [item.deviceId]: (counts[item.deviceId] || 0) + 1,
    };
  }, {});
}

export default function App() {
  const { theme, toggleTheme } = useThemePreference();
  const { route } = useAppRoute();
  const pathname = window.location.pathname;

  if (isPublicStatusPath(pathname)) {
    return (
      <Shell onToggleTheme={toggleTheme} theme={theme}>
        <PublicStatusApp slug={statusSlugFromPath(pathname)} />
      </Shell>
    );
  }

  return <ManagementApp route={route} theme={theme} toggleTheme={toggleTheme} />;
}

function PublicStatusApp({ slug }) {
  const statusPage = usePublicStatusPage(slug);
  return <PublicStatusPageScreen {...statusPage} />;
}

function ManagementApp({ route, theme, toggleTheme }) {
  const app = useHomelabwatchApp();
  const { navigate } = useAppRoute();
  const dashboard = app.data.dashboard;
  const settings = app.data.settings;
  const summary = dashboard?.summary ?? DEFAULT_SUMMARY;
  const metrics = buildMetrics(summary);
  const serviceCounts = countByDevice(dashboard?.services ?? []);
  const discoveryCounts = countByDevice(dashboard?.discoveredServices ?? []);

  let content = null;
  switch (route.id) {
    case "bookmarks":
      content = (
        <BookmarksScreen
          bookmarks={app.data.bookmarks}
          canManageUI={app.bootstrap.trustedNetwork}
          dashboard={dashboard}
          folders={app.data.folders}
          onDeleteBookmark={app.actions.removeBookmark}
          onDeleteFolder={app.actions.removeFolder}
          onExportBookmarks={app.actions.exportBookmarksData}
          onImportBookmarks={app.actions.importBookmarksData}
          onReorderBookmarks={app.actions.saveBookmarkOrder}
          onReorderFolders={app.actions.saveFolderOrder}
          onSaveBookmark={app.actions.saveBookmark}
          onSaveFolder={app.actions.saveFolder}
          onUploadBookmarkIcon={app.actions.uploadBookmarkIcon}
          services={dashboard?.services ?? []}
          tags={app.data.tags}
        />
      );
      break;
    case "services":
      content = (
        <ServicesScreen
          bookmarks={app.data.bookmarks}
          canManageUI={app.bootstrap.trustedNetwork}
          dashboard={dashboard}
          onDeleteServiceHealthCheck={app.actions.removeServiceHealthCheck}
          onFetchServiceHealthChecks={app.actions.loadServiceHealthChecks}
          onSaveBookmarkFromService={app.actions.saveBookmarkFromService}
          onSaveManualService={app.actions.saveManualService}
          onSaveServiceHealthCheck={app.actions.saveServiceHealthCheck}
          onTestServiceCheck={app.actions.runServiceCheckTest}
        />
      );
      break;
    case "health":
      content = (
        <HealthScreen
          bookmarks={app.data.bookmarks}
          canManageUI={app.bootstrap.trustedNetwork}
          dashboard={dashboard}
          onDeleteServiceHealthCheck={app.actions.removeServiceHealthCheck}
          onFetchServiceHealthChecks={app.actions.loadServiceHealthChecks}
          onSaveBookmarkFromService={app.actions.saveBookmarkFromService}
          onSaveManualService={app.actions.saveManualService}
          onSaveServiceHealthCheck={app.actions.saveServiceHealthCheck}
          onTestServiceCheck={app.actions.runServiceCheckTest}
        />
      );
      break;
    case "discovery":
      content = (
        <DiscoveryScreen
          canManageUI={app.bootstrap.trustedNetwork}
          dashboard={dashboard}
          folders={app.data.folders}
          onIgnoreDiscoveredService={app.actions.ignoreSuggestion}
          onRestoreDiscoveredService={app.actions.restoreSuggestion}
          onSaveBookmarkFromDiscoveredService={
            app.actions.saveBookmarkFromDiscoveredService
          }
          onSaveDiscoveryPolicy={app.actions.saveDiscoveryPolicy}
          onSaveDockerEndpoint={app.actions.saveDockerEndpoint}
          onSaveScanTarget={app.actions.saveScanTarget}
          settings={settings}
        />
      );
      break;
    case "devices":
      content = (
        <DevicesScreen
          dashboard={dashboard}
          discoveryCounts={discoveryCounts}
          serviceCounts={serviceCounts}
        />
      );
      break;
    case "notifications":
      content = (
        <NotificationsScreen
          canManageUI={app.bootstrap.trustedNetwork}
          notifications={app.data.notifications}
          onDeleteChannel={app.actions.removeNotificationChannel}
          onDeleteRule={app.actions.removeNotificationRule}
          onSaveChannel={app.actions.saveNotificationChannel}
          onSaveRule={app.actions.saveNotificationRule}
          onTestChannel={app.actions.sendNotificationTest}
        />
      );
      break;
    case "status-pages":
      content = (
        <StatusPagesScreen
          canManageUI={app.bootstrap.trustedNetwork}
          dashboard={dashboard}
          onCreatePage={() =>
            void app.actions.saveStatusPage({
              enabled: true,
              slug: `status-${Date.now().toString(36)}`,
              title: "New Status Page",
              description: "",
            })
          }
          onDeleteAnnouncement={app.actions.removeStatusPageAnnouncement}
          onDeletePage={app.actions.removeStatusPage}
          onSaveAnnouncement={app.actions.saveStatusPageAnnouncement}
          onSavePage={app.actions.saveStatusPage}
          onSaveServices={app.actions.saveStatusPageServices}
          onSelectPage={app.actions.loadStatusPage}
          statusPages={app.data.statusPages}
        />
      );
      break;
    case "definitions":
      content = (
        <DefinitionsScreen
          canManageUI={app.bootstrap.trustedNetwork}
          onDeleteServiceDefinition={app.actions.removeServiceDefinitionRecord}
          onReapplyServiceDefinition={app.actions.rerunServiceDefinition}
          onSaveServiceDefinition={app.actions.saveServiceDefinitionRecord}
          settings={settings}
        />
      );
      break;
    case "settings":
      content = (
        <SettingsScreen
          canManageUI={app.bootstrap.trustedNetwork}
          dashboard={dashboard}
          onCreateAPIToken={app.actions.createExternalToken}
          onRevokeAPIToken={app.actions.revokeExternalToken}
          settings={settings}
        />
      );
      break;
    default:
      content = (
        <OverviewScreen
          bookmarks={app.data.bookmarks}
          canManageUI={app.bootstrap.trustedNetwork}
          dashboard={dashboard}
          folders={app.data.folders}
          metrics={metrics}
          onIgnoreDiscoveredService={app.actions.ignoreSuggestion}
          onNavigate={navigate}
          onRestoreDiscoveredService={app.actions.restoreSuggestion}
          onSaveBookmarkFromDiscoveredService={
            app.actions.saveBookmarkFromDiscoveredService
          }
          settings={settings}
        />
      );
      break;
  }

  return (
    <Shell onToggleTheme={toggleTheme} theme={theme}>
      {app.bootstrap.loading ? (
        <EmptyState
          body="Bootstrapping the dashboard state."
          title="Loading control plane"
        />
      ) : app.bootstrap.initialized ? (
        <AppShell
          activeRoute={route}
          bookmarks={app.data.bookmarks}
          canManageUI={app.bootstrap.trustedNetwork}
          dashboard={dashboard}
          error={app.alerts.error}
          notice={app.alerts.notice}
          notifications={app.data.notifications}
          onNavigate={navigate}
          onRefresh={app.actions.refreshAll}
          onRunDiscovery={app.actions.runDiscovery}
          onRunMonitoring={app.actions.runMonitoring}
          settings={settings}
          statusPages={app.data.statusPages}
        >
          {content}
        </AppShell>
      ) : (
        <BootstrapScreen
          error={app.alerts.error}
          notice={app.alerts.notice}
          onSubmit={app.actions.submitSetup}
          trustedNetwork={app.bootstrap.trustedNetwork}
        />
      )}
    </Shell>
  );
}
