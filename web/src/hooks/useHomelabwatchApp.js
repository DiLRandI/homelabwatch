import { useEffect, useState } from "react";

import {
  createAPIToken,
  createBookmark,
  createBookmarkFromDiscoveredService,
  createBookmarkFromService,
  createServiceCheck,
  createServiceDefinition,
  createFolder,
  createDockerEndpoint,
  createScanTarget,
  createService,
  deleteServiceCheck,
  deleteServiceDefinition,
  deleteBookmark,
  deleteFolder,
  exportBookmarks,
  fetchDashboard,
  fetchBookmarks,
  fetchFolders,
  fetchServiceChecks,
  fetchSettings,
  fetchTags,
  fetchUIBootstrap,
  ignoreDiscoveredService,
  importBookmarks,
  initializeSetup,
  reapplyServiceDefinition,
  reorderBookmarks,
  reorderFolders,
  restoreDiscoveredService,
  revokeAPIToken,
  runDiscoveryJob,
  runMonitoringJob,
  testServiceCheck,
  updateServiceCheck,
  updateServiceDefinition,
  updateDiscoverySettings,
  updateBookmark,
  updateFolder,
  uploadBookmarkAsset,
} from "../lib/api";
import { useServerEvents } from "./useServerEvents";

export function useHomelabwatchApp() {
  const [initialized, setInitialized] = useState(false);
  const [trustedNetwork, setTrustedNetwork] = useState(false);
  const [csrfToken, setCsrfToken] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [dashboard, setDashboard] = useState(null);
  const [bookmarks, setBookmarks] = useState([]);
  const [folders, setFolders] = useState([]);
  const [tags, setTags] = useState([]);
  const [settings, setSettings] = useState(null);

  useEffect(() => {
    void loadBootstrapState();
  }, []);

  useEffect(() => {
    if (!initialized) {
      return;
    }
    void refreshAll();
  }, [initialized]);

  useServerEvents(initialized, () => {
    void refreshAll();
  });

  async function loadBootstrapState() {
    try {
      setLoading(true);
      setError("");
      const payload = await fetchUIBootstrap();
      setInitialized(Boolean(payload.initialized));
      setTrustedNetwork(Boolean(payload.trustedNetwork));
      setCsrfToken(payload.csrfToken || "");
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setLoading(false);
    }
  }

  async function loadDashboard() {
    try {
      const payload = await fetchDashboard();
      setDashboard(payload);
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  async function loadBookmarksWorkspace() {
    try {
      const [bookmarkItems, folderItems, tagItems] = await Promise.all([
        fetchBookmarks(),
        fetchFolders(),
        fetchTags(),
      ]);
      setBookmarks(Array.isArray(bookmarkItems) ? bookmarkItems : []);
      setFolders(Array.isArray(folderItems) ? folderItems : []);
      setTags(Array.isArray(tagItems) ? tagItems : []);
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  async function loadSettings() {
    try {
      const payload = await fetchSettings();
      setSettings(payload);
      setTrustedNetwork(Boolean(payload?.appSettings?.trustedNetwork));
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  async function refreshAll() {
    await Promise.all([loadDashboard(), loadSettings(), loadBookmarksWorkspace()]);
  }

  async function submitSetup(payload) {
    return performAction(async () => {
      await initializeSetup(payload, csrfToken);
      setInitialized(true);
      await refreshAll();
    }, "Workspace initialized.");
  }

  async function saveManualService(payload) {
    return performAction(async () => {
      await createService(payload, csrfToken);
      await loadDashboard();
    }, "Manual service saved.");
  }

  async function loadServiceHealthChecks(serviceId) {
    try {
      setError("");
      return await fetchServiceChecks(serviceId);
    } catch (requestError) {
      setError(requestError.message);
      return [];
    }
  }

  async function saveServiceHealthCheck(serviceId, payload) {
    try {
      setError("");
      setNotice("");
      const saved = payload.id
        ? await updateServiceCheck(payload.id, payload, csrfToken)
        : await createServiceCheck(serviceId, payload, csrfToken);
      await loadDashboard();
      setNotice(payload.id ? "Health check saved." : "Health check created.");
      return saved;
    } catch (requestError) {
      setError(requestError.message);
      return null;
    }
  }

  async function removeServiceHealthCheck(id) {
    try {
      setError("");
      setNotice("");
      await deleteServiceCheck(id, csrfToken);
      await loadDashboard();
      setNotice("Health check deleted.");
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  async function runServiceCheckTest(serviceId, payload) {
    try {
      setError("");
      return await testServiceCheck(serviceId, payload, csrfToken);
    } catch (requestError) {
      setError(requestError.message);
      return null;
    }
  }

  async function saveServiceDefinitionRecord(payload) {
    try {
      setError("");
      setNotice("");
      const saved = payload.id
        ? await updateServiceDefinition(payload.id, payload, csrfToken)
        : await createServiceDefinition(payload, csrfToken);
      await Promise.all([loadDashboard(), loadSettings()]);
      setNotice(payload.id ? "Service definition saved." : "Service definition created.");
      return saved;
    } catch (requestError) {
      setError(requestError.message);
      return null;
    }
  }

  async function removeServiceDefinitionRecord(id) {
    try {
      setError("");
      setNotice("");
      await deleteServiceDefinition(id, csrfToken);
      await Promise.all([loadDashboard(), loadSettings()]);
      setNotice("Service definition deleted.");
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  async function rerunServiceDefinition(id) {
    try {
      setError("");
      setNotice("");
      await reapplyServiceDefinition(id, csrfToken);
      await Promise.all([loadDashboard(), loadSettings()]);
      setNotice("Service definition reapplied.");
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  async function saveBookmark(payload) {
    return performAction(async () => {
      if (payload.id) {
        await updateBookmark(payload.id, payload, csrfToken);
      } else {
        await createBookmark(payload, csrfToken);
      }
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmark saved.");
  }

  async function removeBookmark(id) {
    return performAction(async () => {
      await deleteBookmark(id, csrfToken);
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmark deleted.");
  }

  async function saveFolder(payload) {
    return performAction(async () => {
      if (payload.id) {
        await updateFolder(payload.id, payload, csrfToken);
      } else {
        await createFolder(payload, csrfToken);
      }
      await loadBookmarksWorkspace();
    }, "Folder saved.");
  }

  async function removeFolder(id) {
    return performAction(async () => {
      await deleteFolder(id, csrfToken);
      await loadBookmarksWorkspace();
    }, "Folder deleted.");
  }

  async function saveBookmarkFromService(payload) {
    return performAction(async () => {
      await createBookmarkFromService(payload, csrfToken);
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmark created from service.");
  }

  async function saveBookmarkFromDiscoveredService(id, payload) {
    return performAction(async () => {
      await createBookmarkFromDiscoveredService(id, payload, csrfToken);
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmark created from discovery.");
  }

  async function saveBookmarkOrder(items) {
    return performAction(async () => {
      await reorderBookmarks(items, csrfToken);
      await loadBookmarksWorkspace();
    }, "Bookmark order updated.");
  }

  async function saveFolderOrder(items) {
    return performAction(async () => {
      await reorderFolders(items, csrfToken);
      await loadBookmarksWorkspace();
    }, "Folder order updated.");
  }

  async function uploadBookmarkIcon(file) {
    try {
      setError("");
      const asset = await uploadBookmarkAsset(file, csrfToken);
      return asset;
    } catch (requestError) {
      setError(requestError.message);
      return null;
    }
  }

  async function exportBookmarksData() {
    try {
      setError("");
      return await exportBookmarks();
    } catch (requestError) {
      setError(requestError.message);
      return null;
    }
  }

  async function importBookmarksData(payload) {
    return performAction(async () => {
      await importBookmarks(payload, csrfToken);
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmarks imported.");
  }

  async function saveDockerEndpoint(payload) {
    return performAction(async () => {
      await createDockerEndpoint(payload, csrfToken);
      await loadSettings();
    }, "Docker endpoint saved.");
  }

  async function saveScanTarget(payload) {
    return performAction(async () => {
      await createScanTarget(payload, csrfToken);
      await loadSettings();
    }, "Scan target saved.");
  }

  async function saveDiscoveryPolicy(payload) {
    return performAction(async () => {
      await updateDiscoverySettings(payload, csrfToken);
      await loadSettings();
    }, "Discovery settings saved.");
  }

  async function ignoreSuggestion(id) {
    return performAction(async () => {
      await ignoreDiscoveredService(id, csrfToken);
      await loadDashboard();
    }, "Suggestion ignored.");
  }

  async function restoreSuggestion(id) {
    return performAction(async () => {
      await restoreDiscoveredService(id, csrfToken);
      await loadDashboard();
    }, "Suggestion restored.");
  }

  async function createExternalToken(payload) {
    try {
      setError("");
      setNotice("");
      const created = await createAPIToken(payload, csrfToken);
      await loadSettings();
      setNotice("External API token created.");
      return created;
    } catch (requestError) {
      setError(requestError.message);
      return null;
    }
  }

  async function revokeExternalToken(id) {
    return performAction(async () => {
      await revokeAPIToken(id, csrfToken);
      await loadSettings();
    }, "External API token revoked.");
  }

  async function runDiscovery() {
    return performAction(async () => {
      await runDiscoveryJob(csrfToken);
      await refreshAll();
    }, "Discovery run started.");
  }

  async function runMonitoring() {
    return performAction(async () => {
      await runMonitoringJob(csrfToken);
      await refreshAll();
    }, "Health checks started.");
  }

  async function performAction(action, successMessage) {
    try {
      setError("");
      setNotice("");
      await action();
      setNotice(successMessage);
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  return {
    createExternalToken,
    dashboard,
    error,
    bookmarks,
    folders,
    initialized,
    importBookmarksData,
    loadServiceHealthChecks,
    loading,
    notice,
    removeBookmark,
    removeFolder,
    removeServiceDefinitionRecord,
    removeServiceHealthCheck,
    refreshAll,
    revokeExternalToken,
    runDiscovery,
    runMonitoring,
    runServiceCheckTest,
    rerunServiceDefinition,
    saveBookmark,
    saveBookmarkFromDiscoveredService,
    saveBookmarkFromService,
    saveBookmarkOrder,
    saveDiscoveryPolicy,
    saveDockerEndpoint,
    saveFolder,
    saveFolderOrder,
    saveServiceDefinitionRecord,
    saveServiceHealthCheck,
    saveManualService,
    saveScanTarget,
    settings,
    submitSetup,
    tags,
    trustedNetwork,
    ignoreSuggestion,
    restoreSuggestion,
    uploadBookmarkIcon,
    exportBookmarksData,
  };
}
