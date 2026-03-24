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
          bookmarks={app.bookmarks}
          canManageUI={app.trustedNetwork}
          dashboard={app.dashboard}
          error={app.error}
          folders={app.folders}
          notice={app.notice}
          onCreateAPIToken={app.createExternalToken}
          onDeleteBookmark={app.removeBookmark}
          onDeleteFolder={app.removeFolder}
          onDeleteServiceDefinition={app.removeServiceDefinitionRecord}
          onDeleteServiceHealthCheck={app.removeServiceHealthCheck}
          onExportBookmarks={app.exportBookmarksData}
          onFetchServiceHealthChecks={app.loadServiceHealthChecks}
          onIgnoreDiscoveredService={app.ignoreSuggestion}
          onImportBookmarks={app.importBookmarksData}
          onRefresh={app.refreshAll}
          onReapplyServiceDefinition={app.rerunServiceDefinition}
          onReorderBookmarks={app.saveBookmarkOrder}
          onReorderFolders={app.saveFolderOrder}
          onRevokeAPIToken={app.revokeExternalToken}
          onRunDiscovery={app.runDiscovery}
          onRunMonitoring={app.runMonitoring}
          onSaveBookmark={app.saveBookmark}
          onSaveBookmarkFromDiscoveredService={app.saveBookmarkFromDiscoveredService}
          onSaveBookmarkFromService={app.saveBookmarkFromService}
          onSaveDiscoveryPolicy={app.saveDiscoveryPolicy}
          onSaveDockerEndpoint={app.saveDockerEndpoint}
          onSaveFolder={app.saveFolder}
          onSaveManualService={app.saveManualService}
          onSaveServiceDefinition={app.saveServiceDefinitionRecord}
          onSaveServiceHealthCheck={app.saveServiceHealthCheck}
          onSaveScanTarget={app.saveScanTarget}
          onRestoreDiscoveredService={app.restoreSuggestion}
          onTestServiceCheck={app.runServiceCheckTest}
          settings={app.settings}
          tags={app.tags}
          onUploadBookmarkIcon={app.uploadBookmarkIcon}
        />
      ) : (
        <BootstrapScreen
          error={app.error}
          notice={app.notice}
          onSubmit={app.submitSetup}
          trustedNetwork={app.trustedNetwork}
        />
      )}
    </Shell>
  );
}
