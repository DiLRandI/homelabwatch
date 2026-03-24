export function formatDate(value) {
  if (!value) {
    return "never";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString();
}

export function formatBytes(value) {
  const size = Number(value || 0);
  if (!Number.isFinite(size) || size <= 0) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB"];
  let unitIndex = 0;
  let current = size;

  while (current >= 1024 && unitIndex < units.length - 1) {
    current /= 1024;
    unitIndex += 1;
  }

  const precision = current >= 10 || unitIndex === 0 ? 0 : 1;
  return `${current.toFixed(precision)} ${units[unitIndex]}`;
}

export function formatLatency(value) {
  const latency = Number(value || 0);
  if (!Number.isFinite(latency) || latency <= 0) {
    return "0 ms";
  }
  if (latency < 1000) {
    return `${Math.round(latency)} ms`;
  }
  return `${(latency / 1000).toFixed(1)} s`;
}
