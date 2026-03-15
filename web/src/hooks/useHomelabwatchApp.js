import { useEffect, useState } from "react";

import {
  createBookmark,
  createDockerEndpoint,
  createScanTarget,
  createService,
  fetchBootstrapStatus,
  fetchDashboard,
  fetchSettings,
  initializeBootstrap,
  runDiscoveryJob,
  runMonitoringJob,
} from "../lib/api";
import { useServerEvents } from "./useServerEvents";

export function useHomelabwatchApp() {
  const [initialized, setInitialized] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [dashboard, setDashboard] = useState(null);
  const [settings, setSettings] = useState(null);
  const [adminToken, setAdminToken] = useState(
    () => localStorage.getItem("homelabwatch-admin-token") || "",
  );

  useEffect(() => {
    localStorage.setItem("homelabwatch-admin-token", adminToken);
  }, [adminToken]);

  useEffect(() => {
    void loadBootstrapStatus();
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

  async function loadBootstrapStatus() {
    try {
      setLoading(true);
      setError("");
      const payload = await fetchBootstrapStatus();
      setInitialized(Boolean(payload.initialized));
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setLoading(false);
    }
  }

  async function loadDashboard() {
    try {
      const payload = await fetchDashboard(adminToken);
      setDashboard(payload);
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  async function loadSettings() {
    try {
      const payload = await fetchSettings(adminToken);
      setSettings(payload);
      return true;
    } catch (requestError) {
      setError(requestError.message);
      return false;
    }
  }

  async function refreshAll() {
    await Promise.all([loadDashboard(), loadSettings()]);
  }

  async function submitBootstrap(payload) {
    return performAction(async () => {
      await initializeBootstrap(payload, adminToken);
      setAdminToken(payload.adminToken);
      setInitialized(true);
      await refreshAll();
    }, "Bootstrap completed.");
  }

  async function saveManualService(payload) {
    return performAction(async () => {
      await createService(payload, adminToken);
      await loadDashboard();
    }, "Manual service saved.");
  }

  async function saveBookmark(payload) {
    return performAction(async () => {
      await createBookmark(payload, adminToken);
      await loadDashboard();
    }, "Bookmark saved.");
  }

  async function saveDockerEndpoint(payload) {
    return performAction(async () => {
      await createDockerEndpoint(payload, adminToken);
      await loadSettings();
    }, "Docker endpoint saved.");
  }

  async function saveScanTarget(payload) {
    return performAction(async () => {
      await createScanTarget(payload, adminToken);
      await loadSettings();
    }, "Scan target saved.");
  }

  async function runDiscovery() {
    return performAction(async () => {
      await runDiscoveryJob(adminToken);
      await refreshAll();
    }, "Discovery run started.");
  }

  async function runMonitoring() {
    return performAction(async () => {
      await runMonitoringJob(adminToken);
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
    adminToken,
    dashboard,
    error,
    initialized,
    loading,
    notice,
    refreshAll,
    runDiscovery,
    runMonitoring,
    saveBookmark,
    saveDockerEndpoint,
    saveManualService,
    saveScanTarget,
    setAdminToken,
    settings,
    submitBootstrap,
  };
}
