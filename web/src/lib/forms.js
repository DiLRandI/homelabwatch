export const defaultBootstrapForm = {
  adminToken: "",
  autoScanEnabled: true,
  defaultScanPorts: "22,80,443,8080,8443",
  seedCIDRs: "",
};

export const defaultServiceForm = {
  name: "",
  url: "",
};

export const defaultBookmarkForm = {
  name: "",
  url: "",
  description: "",
};

export const defaultDockerEndpointForm = {
  name: "Remote Docker",
  kind: "remote",
  address: "tcp://192.168.1.100:2375",
  enabled: true,
  scanIntervalSeconds: 30,
};

export const defaultScanTargetForm = {
  name: "Lab subnet",
  cidr: "192.168.1.0/24",
  commonPorts: "22,80,443,8080,8443",
  enabled: true,
  scanIntervalSeconds: 300,
};

export function parsePorts(raw) {
  return raw
    .split(",")
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isFinite(item) && item > 0);
}

export function parseCIDRTargets(raw, portsRaw) {
  const ports = parsePorts(portsRaw);

  return raw
    .split(/\n|,/)
    .map((item) => item.trim())
    .filter(Boolean)
    .map((cidr) => ({
      name: cidr,
      cidr,
      enabled: true,
      autoDetected: false,
      scanIntervalSeconds: 300,
      commonPorts: ports,
    }));
}
