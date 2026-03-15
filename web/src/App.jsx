import BootstrapScreen from "./components/bootstrap/BootstrapScreen";
import DashboardScreen from "./components/dashboard/DashboardScreen";
import EmptyState from "./components/ui/EmptyState";
import Shell from "./components/ui/Shell";
import { useHomelabwatchApp } from "./hooks/useHomelabwatchApp";

export default function App() {
  const app = useHomelabwatchApp();

  return (
    <Shell>
      {app.loading ? (
        <EmptyState
          title="Loading control plane"
          body="Bootstrapping the dashboard state."
        />
      ) : app.initialized ? (
        <DashboardScreen
          adminToken={app.adminToken}
          dashboard={app.dashboard}
          error={app.error}
          notice={app.notice}
          onAdminTokenChange={app.setAdminToken}
          onRefresh={app.refreshAll}
          onRunDiscovery={app.runDiscovery}
          onRunMonitoring={app.runMonitoring}
          onSaveBookmark={app.saveBookmark}
          onSaveDockerEndpoint={app.saveDockerEndpoint}
          onSaveManualService={app.saveManualService}
          onSaveScanTarget={app.saveScanTarget}
          settings={app.settings}
        />
      ) : (
        <BootstrapScreen
          error={app.error}
          notice={app.notice}
          onSubmit={app.submitBootstrap}
        />
      )}
    </Shell>
  );
}
