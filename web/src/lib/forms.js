export const defaultSetupForm = {
  applianceName: "HomelabWatch",
  autoScanEnabled: true,
  defaultScanPorts: "22,80,443,8080,8443",
  runDiscovery: true,
  seedCIDRs: "",
};

export const defaultServiceForm = {
  healthUrl: "",
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

export const defaultTopologySourceForm = {
  name: "Core switch",
  address: "192.168.1.2",
  port: 161,
  enabled: true,
  pollIntervalSeconds: 300,
  timeoutMs: 1500,
  retries: 1,
  snmpVersion: "v2c",
  community: "",
  username: "",
  authProtocol: "none",
  authPassphrase: "",
  privacyProtocol: "none",
  privacyPassphrase: "",
  role: "switch",
  root: false,
};

export const defaultAPITokenForm = {
  name: "Automation token",
  scope: "write",
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
