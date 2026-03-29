import { useEffect, useRef, useState } from "react";

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
  fetchServiceChecks,
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
import { useBookmarksData } from "./useBookmarksData";
import { useDashboardData } from "./useDashboardData";
import { useSettingsData } from "./useSettingsData";
import { useServerEvents } from "./useServerEvents";
import { useUIBootstrap } from "./useUIBootstrap";

const REFRESH_DEBOUNCE_MS = 1500;
const REFRESH_MAX_WAIT_MS = 5000;

export function useHomelabwatchApp() {
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [hydrated, setHydrated] = useState(false);
  const hydratedRef = useRef(false);
  const refreshStateRef = useRef({
    bookmarks: { dirty: false, maxTimerID: null, timerID: null },
    dashboard: { dirty: false, maxTimerID: null, timerID: null },
    settings: { dirty: false, maxTimerID: null, timerID: null },
  });
  const bootstrap = useUIBootstrap({ onError: setError });
  const dashboardState = useDashboardData({ onError: setError });
  const bookmarksState = useBookmarksData({ onError: setError });
  const settingsState = useSettingsData({
    onError: setError,
    onTrustedNetworkChange: bootstrap.setTrustedNetwork,
  });

  function clearRefreshTimers(kind) {
    const state = refreshStateRef.current[kind];
    if (state.timerID != null) {
      window.clearTimeout(state.timerID);
      state.timerID = null;
    }
    if (state.maxTimerID != null) {
      window.clearTimeout(state.maxTimerID);
      state.maxTimerID = null;
    }
  }

  async function loadDashboard() {
    return Boolean(await dashboardState.loadDashboard());
  }

  async function loadBookmarksWorkspace() {
    return bookmarksState.loadBookmarksWorkspace();
  }

  async function loadSettings() {
    return Boolean(await settingsState.loadSettings());
  }

  async function flushRefresh(kind) {
    const state = refreshStateRef.current[kind];
    clearRefreshTimers(kind);
    if (!state.dirty) {
      return null;
    }
    state.dirty = false;
    switch (kind) {
      case "bookmarks":
        return loadBookmarksWorkspace();
      case "settings":
        return loadSettings();
      default:
        return loadDashboard();
    }
  }

  function scheduleRefresh(kind) {
    const state = refreshStateRef.current[kind];
    state.dirty = true;
    if (!bootstrap.initialized || !hydratedRef.current) {
      return;
    }
    if (state.timerID == null) {
      state.timerID = window.setTimeout(() => {
        void flushRefresh(kind);
      }, REFRESH_DEBOUNCE_MS);
    }
    if (state.maxTimerID == null) {
      state.maxTimerID = window.setTimeout(() => {
        void flushRefresh(kind);
      }, REFRESH_MAX_WAIT_MS);
    }
  }

  function queueRefreshes(...kinds) {
    for (const kind of kinds) {
      scheduleRefresh(kind);
    }
  }

  async function flushPendingRefreshes() {
    if (!hydratedRef.current) {
      return;
    }
    const pendingKinds = Object.entries(refreshStateRef.current)
      .filter(([, state]) => state.dirty)
      .map(([kind]) => kind);
    if (pendingKinds.length == 0) {
      return;
    }
    await Promise.all(pendingKinds.map((kind) => flushRefresh(kind)));
  }

  async function refreshAll() {
    await Promise.all([loadDashboard(), loadSettings(), loadBookmarksWorkspace()]);
  }

  useEffect(() => {
    if (!bootstrap.initialized) {
      hydratedRef.current = false;
      setHydrated(false);
      return;
    }

    let active = true;
    hydratedRef.current = false;
    setHydrated(false);

    async function hydrateApp() {
      await refreshAll();
      if (active) {
        hydratedRef.current = true;
        setHydrated(true);
        await flushPendingRefreshes();
      }
    }

    void hydrateApp();
    return () => {
      active = false;
    };
  }, [bootstrap.initialized]);

  useEffect(() => {
    return () => {
      for (const kind of Object.keys(refreshStateRef.current)) {
        clearRefreshTimers(kind);
      }
    };
  }, []);

  useServerEvents(bootstrap.initialized, {
    bootstrap: async () => {
      const payload = await bootstrap.loadBootstrapState();
      if (payload?.initialized) {
        hydratedRef.current = false;
        setHydrated(false);
        await refreshAll();
        hydratedRef.current = true;
        setHydrated(true);
        await flushPendingRefreshes();
      }
    },
    bookmark: () => queueRefreshes("dashboard", "bookmarks"),
    check: () => queueRefreshes("dashboard", "bookmarks"),
    device: () => queueRefreshes("dashboard", "bookmarks"),
    "discovered-service": () => queueRefreshes("dashboard"),
    "docker-endpoint": () => queueRefreshes("settings"),
    folder: () => queueRefreshes("bookmarks"),
    "scan-target": () => queueRefreshes("settings"),
    service: () => queueRefreshes("dashboard", "bookmarks"),
    "service-definition": () => queueRefreshes("dashboard", "settings"),
  });

  async function submitSetup(payload) {
    return performAction(async () => {
      await initializeSetup(payload, bootstrap.csrfToken);
      bootstrap.markInitialized();
      setHydrated(false);
    }, "Workspace initialized.");
  }

  async function saveManualService(payload) {
    return performAction(async () => {
      await createService(payload, bootstrap.csrfToken);
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
        ? await updateServiceCheck(payload.id, payload, bootstrap.csrfToken)
        : await createServiceCheck(serviceId, payload, bootstrap.csrfToken);
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
      await deleteServiceCheck(id, bootstrap.csrfToken);
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
      return await testServiceCheck(serviceId, payload, bootstrap.csrfToken);
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
        ? await updateServiceDefinition(payload.id, payload, bootstrap.csrfToken)
        : await createServiceDefinition(payload, bootstrap.csrfToken);
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
      await deleteServiceDefinition(id, bootstrap.csrfToken);
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
      await reapplyServiceDefinition(id, bootstrap.csrfToken);
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
        await updateBookmark(payload.id, payload, bootstrap.csrfToken);
      } else {
        await createBookmark(payload, bootstrap.csrfToken);
      }
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmark saved.");
  }

  async function removeBookmark(id) {
    return performAction(async () => {
      await deleteBookmark(id, bootstrap.csrfToken);
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmark deleted.");
  }

  async function saveFolder(payload) {
    return performAction(async () => {
      if (payload.id) {
        await updateFolder(payload.id, payload, bootstrap.csrfToken);
      } else {
        await createFolder(payload, bootstrap.csrfToken);
      }
      await loadBookmarksWorkspace();
    }, "Folder saved.");
  }

  async function removeFolder(id) {
    return performAction(async () => {
      await deleteFolder(id, bootstrap.csrfToken);
      await loadBookmarksWorkspace();
    }, "Folder deleted.");
  }

  async function saveBookmarkFromService(payload) {
    return performAction(async () => {
      await createBookmarkFromService(payload, bootstrap.csrfToken);
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmark created from service.");
  }

  async function saveBookmarkFromDiscoveredService(id, payload) {
    return performAction(async () => {
      await createBookmarkFromDiscoveredService(id, payload, bootstrap.csrfToken);
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmark created from discovery.");
  }

  async function saveBookmarkOrder(items) {
    return performAction(async () => {
      await reorderBookmarks(items, bootstrap.csrfToken);
      await loadBookmarksWorkspace();
    }, "Bookmark order updated.");
  }

  async function saveFolderOrder(items) {
    return performAction(async () => {
      await reorderFolders(items, bootstrap.csrfToken);
      await loadBookmarksWorkspace();
    }, "Folder order updated.");
  }

  async function uploadBookmarkIcon(file) {
    try {
      setError("");
      const asset = await uploadBookmarkAsset(file, bootstrap.csrfToken);
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
      await importBookmarks(payload, bootstrap.csrfToken);
      await Promise.all([loadDashboard(), loadBookmarksWorkspace()]);
    }, "Bookmarks imported.");
  }

  async function saveDockerEndpoint(payload) {
    return performAction(async () => {
      await createDockerEndpoint(payload, bootstrap.csrfToken);
      await loadSettings();
    }, "Docker endpoint saved.");
  }

  async function saveScanTarget(payload) {
    return performAction(async () => {
      await createScanTarget(payload, bootstrap.csrfToken);
      await loadSettings();
    }, "Scan target saved.");
  }

  async function saveDiscoveryPolicy(payload) {
    return performAction(async () => {
      await updateDiscoverySettings(payload, bootstrap.csrfToken);
      await loadSettings();
    }, "Discovery settings saved.");
  }

  async function ignoreSuggestion(id) {
    return performAction(async () => {
      await ignoreDiscoveredService(id, bootstrap.csrfToken);
      await loadDashboard();
    }, "Suggestion ignored.");
  }

  async function restoreSuggestion(id) {
    return performAction(async () => {
      await restoreDiscoveredService(id, bootstrap.csrfToken);
      await loadDashboard();
    }, "Suggestion restored.");
  }

  async function createExternalToken(payload) {
    try {
      setError("");
      setNotice("");
      const created = await createAPIToken(payload, bootstrap.csrfToken);
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
      await revokeAPIToken(id, bootstrap.csrfToken);
      await loadSettings();
    }, "External API token revoked.");
  }

  async function runDiscovery() {
    return performAction(async () => {
      await runDiscoveryJob(bootstrap.csrfToken);
      await refreshAll();
    }, "Discovery run started.");
  }

  async function runMonitoring() {
    return performAction(async () => {
      await runMonitoringJob(bootstrap.csrfToken);
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
    actions: {
      createExternalToken,
      exportBookmarksData,
      ignoreSuggestion,
      importBookmarksData,
      loadServiceHealthChecks,
      refreshAll,
      removeBookmark,
      removeFolder,
      removeServiceDefinitionRecord,
      removeServiceHealthCheck,
      restoreSuggestion,
      revokeExternalToken,
      runDiscovery,
      runMonitoring,
      runServiceCheckTest,
      saveBookmark,
      saveBookmarkFromDiscoveredService,
      saveBookmarkFromService,
      saveBookmarkOrder,
      saveDiscoveryPolicy,
      saveDockerEndpoint,
      saveFolder,
      saveFolderOrder,
      saveManualService,
      saveScanTarget,
      saveServiceDefinitionRecord,
      saveServiceHealthCheck,
      submitSetup,
      uploadBookmarkIcon,
      rerunServiceDefinition,
    },
    alerts: {
      error,
      notice,
    },
    bootstrap: {
      initialized: bootstrap.initialized,
      loading: bootstrap.loading || (bootstrap.initialized && !hydrated),
      trustedNetwork: bootstrap.trustedNetwork,
    },
    data: {
      bookmarks: bookmarksState.bookmarks,
      dashboard: dashboardState.dashboard,
      folders: bookmarksState.folders,
      settings: settingsState.settings,
      tags: bookmarksState.tags,
    },
  };
}
