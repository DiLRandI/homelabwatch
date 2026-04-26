async function request(path, { csrfToken = "", ...options } = {}) {
  const headers = new Headers(options.headers || {});
  const method = (options.method || "GET").toUpperCase();

  if (
    !headers.has("Content-Type") &&
    options.body &&
    !(options.body instanceof FormData)
  ) {
    headers.set("Content-Type", "application/json");
  }
  if (method !== "GET" && method !== "HEAD" && csrfToken) {
    headers.set("X-Homelabwatch-CSRF", csrfToken);
  }

  const response = await fetch(path, {
    ...options,
    credentials: "same-origin",
    headers,
  });
  if (!response.ok) {
    const payload = await response
      .json()
      .catch(() => ({ error: response.statusText }));
    throw new Error(payload.error || response.statusText);
  }
  if (response.status === 204) {
    return null;
  }
  return response.json();
}

export function fetchUIBootstrap() {
  return request("/api/ui/v1/bootstrap");
}

export function fetchDashboard() {
  return request("/api/ui/v1/dashboard");
}

export function fetchBookmarks(params = {}) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") {
      continue;
    }
    query.set(key, String(value));
  }
  const suffix = query.toString() ? `?${query.toString()}` : "";
  return request(`/api/ui/v1/bookmarks${suffix}`);
}

export function fetchFolders() {
  return request("/api/ui/v1/folders");
}

export function fetchTags() {
  return request("/api/ui/v1/tags");
}

export function fetchSettings() {
  return request("/api/ui/v1/settings");
}

export function initializeSetup(payload, csrfToken) {
  return request("/api/ui/v1/setup", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function createService(payload, csrfToken) {
  return request("/api/ui/v1/services", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function fetchServiceChecks(id) {
  return request(`/api/ui/v1/services/${id}/checks`);
}

export function createServiceCheck(serviceId, payload, csrfToken) {
  return request(`/api/ui/v1/services/${serviceId}/checks`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function updateServiceCheck(id, payload, csrfToken) {
  return request(`/api/ui/v1/checks/${id}`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "PATCH",
  });
}

export function deleteServiceCheck(id, csrfToken) {
  return request(`/api/ui/v1/checks/${id}`, {
    csrfToken,
    method: "DELETE",
  });
}

export function testServiceCheck(serviceId, payload, csrfToken) {
  return request(`/api/ui/v1/services/${serviceId}/checks/test`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function fetchServiceDefinitions() {
  return request("/api/ui/v1/service-definitions");
}

export function createServiceDefinition(payload, csrfToken) {
  return request("/api/ui/v1/service-definitions", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function updateServiceDefinition(id, payload, csrfToken) {
  return request(`/api/ui/v1/service-definitions/${id}`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "PATCH",
  });
}

export function deleteServiceDefinition(id, csrfToken) {
  return request(`/api/ui/v1/service-definitions/${id}`, {
    csrfToken,
    method: "DELETE",
  });
}

export function reapplyServiceDefinition(id, csrfToken) {
  return request(`/api/ui/v1/service-definitions/${id}/reapply`, {
    body: JSON.stringify({}),
    csrfToken,
    method: "POST",
  });
}

export function createBookmark(payload, csrfToken) {
  return request("/api/ui/v1/bookmarks", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function updateBookmark(id, payload, csrfToken) {
  return request(`/api/ui/v1/bookmarks/${id}`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "PUT",
  });
}

export function deleteBookmark(id, csrfToken) {
  return request(`/api/ui/v1/bookmarks/${id}`, {
    csrfToken,
    method: "DELETE",
  });
}

export function createBookmarkFromService(payload, csrfToken) {
  return request("/api/ui/v1/bookmarks/from-service", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function createBookmarkFromDiscoveredService(id, payload, csrfToken) {
  return request(`/api/ui/v1/discovered-services/${id}/bookmark`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function ignoreDiscoveredService(id, csrfToken) {
  return request(`/api/ui/v1/discovered-services/${id}/ignore`, {
    body: JSON.stringify({}),
    csrfToken,
    method: "POST",
  });
}

export function restoreDiscoveredService(id, csrfToken) {
  return request(`/api/ui/v1/discovered-services/${id}/restore`, {
    body: JSON.stringify({}),
    csrfToken,
    method: "POST",
  });
}

export function reorderBookmarks(payload, csrfToken) {
  return request("/api/ui/v1/bookmarks/reorder", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function exportBookmarks() {
  return request("/api/ui/v1/bookmarks/export");
}

export function importBookmarks(payload, csrfToken) {
  return request("/api/ui/v1/bookmarks/import", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function uploadBookmarkAsset(file, csrfToken) {
  const formData = new FormData();
  formData.set("file", file);

  return request("/api/ui/v1/bookmark-assets", {
    body: formData,
    csrfToken,
    headers: {},
    method: "POST",
  });
}

export function bookmarkOpenURL(id) {
  return `/api/ui/v1/bookmarks/${id}/open`;
}

export function createFolder(payload, csrfToken) {
  return request("/api/ui/v1/folders", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function updateFolder(id, payload, csrfToken) {
  return request(`/api/ui/v1/folders/${id}`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "PUT",
  });
}

export function deleteFolder(id, csrfToken) {
  return request(`/api/ui/v1/folders/${id}`, {
    csrfToken,
    method: "DELETE",
  });
}

export function reorderFolders(payload, csrfToken) {
  return request("/api/ui/v1/folders/reorder", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function createDockerEndpoint(payload, csrfToken) {
  return request("/api/ui/v1/discovery/docker-endpoints", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function createScanTarget(payload, csrfToken) {
  return request("/api/ui/v1/discovery/scan-targets", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function updateDiscoverySettings(payload, csrfToken) {
  return request("/api/ui/v1/discovery/settings", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "PATCH",
  });
}

export function createAPIToken(payload, csrfToken) {
  return request("/api/ui/v1/settings/api-tokens", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function revokeAPIToken(id, csrfToken) {
  return request(`/api/ui/v1/settings/api-tokens/${id}`, {
    csrfToken,
    method: "DELETE",
  });
}

export function runDiscoveryJob(csrfToken) {
  return request("/api/ui/v1/discovery/run", {
    csrfToken,
    method: "POST",
  });
}

export function runMonitoringJob(csrfToken) {
  return request("/api/ui/v1/monitoring/run", {
    csrfToken,
    method: "POST",
  });
}

export function fetchNotificationChannels() {
  return request("/api/ui/v1/notifications/channels");
}

export function createNotificationChannel(payload, csrfToken) {
  return request("/api/ui/v1/notifications/channels", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function updateNotificationChannel(id, payload, csrfToken) {
  return request(`/api/ui/v1/notifications/channels/${id}`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "PATCH",
  });
}

export function deleteNotificationChannel(id, csrfToken) {
  return request(`/api/ui/v1/notifications/channels/${id}`, {
    csrfToken,
    method: "DELETE",
  });
}

export function testNotificationChannel(id, csrfToken) {
  return request(`/api/ui/v1/notifications/channels/${id}/test`, {
    body: JSON.stringify({}),
    csrfToken,
    method: "POST",
  });
}

export function fetchNotificationRules() {
  return request("/api/ui/v1/notifications/rules");
}

export function createNotificationRule(payload, csrfToken) {
  return request("/api/ui/v1/notifications/rules", {
    body: JSON.stringify(payload),
    csrfToken,
    method: "POST",
  });
}

export function updateNotificationRule(id, payload, csrfToken) {
  return request(`/api/ui/v1/notifications/rules/${id}`, {
    body: JSON.stringify(payload),
    csrfToken,
    method: "PATCH",
  });
}

export function deleteNotificationRule(id, csrfToken) {
  return request(`/api/ui/v1/notifications/rules/${id}`, {
    csrfToken,
    method: "DELETE",
  });
}

export function fetchNotificationDeliveries() {
  return request("/api/ui/v1/notifications/deliveries");
}
