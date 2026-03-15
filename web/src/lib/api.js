async function request(path, { csrfToken = "", ...options } = {}) {
  const headers = new Headers(options.headers || {});
  const method = (options.method || "GET").toUpperCase();

  if (!headers.has("Content-Type") && options.body) {
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

export function createBookmark(payload, csrfToken) {
  return request("/api/ui/v1/bookmarks", {
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
