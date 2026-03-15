import { useEffect, useState } from "react";

import {
  createAPIToken,
  createBookmark,
  createDockerEndpoint,
  createScanTarget,
  createService,
  fetchDashboard,
  fetchSettings,
  fetchUIBootstrap,
  initializeSetup,
  revokeAPIToken,
  runDiscoveryJob,
  runMonitoringJob,
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
    await Promise.all([loadDashboard(), loadSettings()]);
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

  async function saveBookmark(payload) {
    return performAction(async () => {
      await createBookmark(payload, csrfToken);
      await loadDashboard();
    }, "Bookmark saved.");
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
    initialized,
    loading,
    notice,
    refreshAll,
    revokeExternalToken,
    runDiscovery,
    runMonitoring,
    saveBookmark,
    saveDockerEndpoint,
    saveManualService,
    saveScanTarget,
    settings,
    submitSetup,
    trustedNetwork,
  };
}
