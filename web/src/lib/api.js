async function request(path, { token = "", ...options } = {}) {
  const headers = new Headers(options.headers || {});

  if (!headers.has("Content-Type") && options.body) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("X-Admin-Token", token);
  }

  const response = await fetch(path, { ...options, headers });
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

export function fetchBootstrapStatus() {
  return request("/api/v1/bootstrap/status");
}

export function fetchDashboard(token) {
  return request("/api/v1/dashboard", { token });
}

export function fetchSettings(token) {
  return request("/api/v1/settings", { token });
}

export function initializeBootstrap(payload, token) {
  return request("/api/v1/bootstrap/init", {
    body: JSON.stringify(payload),
    method: "POST",
    token,
  });
}

export function createService(payload, token) {
  return request("/api/v1/services", {
    body: JSON.stringify(payload),
    method: "POST",
    token,
  });
}

export function createBookmark(payload, token) {
  return request("/api/v1/bookmarks", {
    body: JSON.stringify(payload),
    method: "POST",
    token,
  });
}

export function createDockerEndpoint(payload, token) {
  return request("/api/v1/discovery/docker-endpoints", {
    body: JSON.stringify(payload),
    method: "POST",
    token,
  });
}

export function createScanTarget(payload, token) {
  return request("/api/v1/discovery/scan-targets", {
    body: JSON.stringify(payload),
    method: "POST",
    token,
  });
}

export function runDiscoveryJob(token) {
  return request("/api/v1/discovery/run", {
    method: "POST",
    token,
  });
}

export function runMonitoringJob(token) {
  return request("/api/v1/monitoring/run", {
    method: "POST",
    token,
  });
}
