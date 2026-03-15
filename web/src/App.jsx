import { useEffect, useMemo, useState } from "react";

const defaultBootstrap = {
  adminToken: "",
  autoScanEnabled: true,
  defaultScanPorts: "22,80,443,8080,8443",
  seedCIDRs: "",
};

const defaultService = {
  name: "",
  url: "",
};

const defaultBookmark = {
  name: "",
  url: "",
  description: "",
};

const defaultDockerEndpoint = {
  name: "Remote Docker",
  kind: "remote",
  address: "tcp://192.168.1.100:2375",
  enabled: true,
  scanIntervalSeconds: 30,
};

const defaultScanTarget = {
  name: "Lab subnet",
  cidr: "192.168.1.0/24",
  commonPorts: "22,80,443,8080,8443",
  enabled: true,
  scanIntervalSeconds: 300,
};

export default function App() {
  const [initialized, setInitialized] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [dashboard, setDashboard] = useState(null);
  const [settings, setSettings] = useState(null);
  const [bootstrapForm, setBootstrapForm] = useState(defaultBootstrap);
  const [serviceForm, setServiceForm] = useState(defaultService);
  const [bookmarkForm, setBookmarkForm] = useState(defaultBookmark);
  const [dockerForm, setDockerForm] = useState(defaultDockerEndpoint);
  const [scanTargetForm, setScanTargetForm] = useState(defaultScanTarget);
  const [adminToken, setAdminToken] = useState(() => localStorage.getItem("homelabwatch-admin-token") || "");

  const summary = dashboard?.summary ?? {
    totalServices: 0,
    healthyServices: 0,
    degradedServices: 0,
    unhealthyServices: 0,
    devicesSeen: 0,
    bookmarks: 0,
  };

  useEffect(() => {
    localStorage.setItem("homelabwatch-admin-token", adminToken);
  }, [adminToken]);

  useEffect(() => {
    bootstrapStatus();
  }, []);

  useEffect(() => {
    if (!initialized) {
      return undefined;
    }
    loadDashboard();
    loadSettings();
    const events = new EventSource("/api/v1/events");
    const refresh = () => {
      loadDashboard();
      loadSettings();
    };
    events.addEventListener("service", refresh);
    events.addEventListener("device", refresh);
    events.addEventListener("check", refresh);
    events.addEventListener("bookmark", refresh);
    events.addEventListener("docker-endpoint", refresh);
    events.addEventListener("scan-target", refresh);
    return () => events.close();
  }, [initialized]);

  const services = dashboard?.services ?? [];
  const devices = dashboard?.devices ?? [];
  const bookmarks = dashboard?.bookmarks ?? [];
  const recentEvents = dashboard?.recentEvents ?? [];
  const dockerEndpoints = settings?.dockerEndpoints ?? [];
  const scanTargets = settings?.scanTargets ?? [];
  const jobState = settings?.jobState ?? [];

  const metrics = useMemo(
    () => [
      { label: "Services", value: summary.totalServices, tone: "text-ink" },
      { label: "Healthy", value: summary.healthyServices, tone: "text-ok" },
      { label: "Degraded", value: summary.degradedServices, tone: "text-warn" },
      { label: "Unhealthy", value: summary.unhealthyServices, tone: "text-danger" },
      { label: "Devices", value: summary.devicesSeen, tone: "text-accent" },
      { label: "Bookmarks", value: summary.bookmarks, tone: "text-ink" },
    ],
    [summary],
  );

  async function bootstrapStatus() {
    try {
      setLoading(true);
      setError("");
      const response = await fetch("/api/v1/bootstrap/status");
      const payload = await response.json();
      setInitialized(Boolean(payload.initialized));
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setLoading(false);
    }
  }

  async function api(path, options = {}) {
    const headers = new Headers(options.headers || {});
    if (!headers.has("Content-Type") && options.body) {
      headers.set("Content-Type", "application/json");
    }
    if (adminToken) {
      headers.set("X-Admin-Token", adminToken);
    }
    const response = await fetch(path, { ...options, headers });
    if (!response.ok) {
      const payload = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(payload.error || response.statusText);
    }
    if (response.status === 204) {
      return null;
    }
    return response.json();
  }

  async function loadDashboard() {
    try {
      const payload = await api("/api/v1/dashboard");
      setDashboard(payload);
    } catch (requestError) {
      setError(requestError.message);
    }
  }

  async function loadSettings() {
    try {
      const payload = await api("/api/v1/settings");
      setSettings(payload);
    } catch (requestError) {
      setError(requestError.message);
    }
  }

  async function submitBootstrap(event) {
    event.preventDefault();
    setError("");
    setNotice("");
    try {
      await api("/api/v1/bootstrap/init", {
        method: "POST",
        body: JSON.stringify({
          adminToken: bootstrapForm.adminToken,
          autoScanEnabled: bootstrapForm.autoScanEnabled,
          defaultScanPorts: parsePorts(bootstrapForm.defaultScanPorts),
          scanTargets: parseCIDRTargets(bootstrapForm.seedCIDRs, bootstrapForm.defaultScanPorts),
        }),
      });
      setAdminToken(bootstrapForm.adminToken);
      setInitialized(true);
      setNotice("Bootstrap completed.");
    } catch (requestError) {
      setError(requestError.message);
    }
  }

  async function saveManualService(event) {
    event.preventDefault();
    try {
      await api("/api/v1/services", {
        method: "POST",
        body: JSON.stringify(serviceForm),
      });
      setServiceForm(defaultService);
      setNotice("Manual service saved.");
      await loadDashboard();
    } catch (requestError) {
      setError(requestError.message);
    }
  }

  async function saveBookmark(event) {
    event.preventDefault();
    try {
      await api("/api/v1/bookmarks", {
        method: "POST",
        body: JSON.stringify(bookmarkForm),
      });
      setBookmarkForm(defaultBookmark);
      setNotice("Bookmark saved.");
      await loadDashboard();
    } catch (requestError) {
      setError(requestError.message);
    }
  }

  async function saveDockerEndpoint(event) {
    event.preventDefault();
    try {
      await api("/api/v1/discovery/docker-endpoints", {
        method: "POST",
        body: JSON.stringify({
          ...dockerForm,
          scanIntervalSeconds: Number(dockerForm.scanIntervalSeconds || 30),
        }),
      });
      setDockerForm(defaultDockerEndpoint);
      setNotice("Docker endpoint saved.");
      await loadSettings();
    } catch (requestError) {
      setError(requestError.message);
    }
  }

  async function saveScanTarget(event) {
    event.preventDefault();
    try {
      await api("/api/v1/discovery/scan-targets", {
        method: "POST",
        body: JSON.stringify({
          ...scanTargetForm,
          commonPorts: parsePorts(scanTargetForm.commonPorts),
          scanIntervalSeconds: Number(scanTargetForm.scanIntervalSeconds || 300),
        }),
      });
      setScanTargetForm(defaultScanTarget);
      setNotice("Scan target saved.");
      await loadSettings();
    } catch (requestError) {
      setError(requestError.message);
    }
  }

  async function runJob(path, label) {
    try {
      await api(path, { method: "POST" });
      setNotice(label);
      await loadDashboard();
      await loadSettings();
    } catch (requestError) {
      setError(requestError.message);
    }
  }

  if (loading) {
    return <Shell><EmptyState title="Loading control plane" body="Bootstrapping the dashboard state." /></Shell>;
  }

  if (!initialized) {
    return (
      <Shell>
        <section className="mx-auto max-w-3xl animate-floatIn rounded-[2rem] border border-white/10 bg-panel/80 p-8 shadow-halo backdrop-blur">
          <div className="mb-8 flex items-end justify-between gap-4">
            <div>
              <p className="text-sm uppercase tracking-[0.35em] text-accent">Homelabwatch</p>
              <h1 className="mt-2 font-display text-4xl font-semibold text-ink">Single-container homelab control plane</h1>
            </div>
            <div className="rounded-full border border-white/10 bg-white/5 px-4 py-2 text-xs text-muted">
              Bootstrap required
            </div>
          </div>
          <p className="max-w-2xl text-sm leading-7 text-muted">
            Initialize the embedded database, set the write token, and optionally seed scan targets. Docker socket discovery is added automatically when the container has access to <code>/var/run/docker.sock</code>.
          </p>
          <form className="mt-8 grid gap-4 md:grid-cols-2" onSubmit={submitBootstrap}>
            <Input
              label="Admin token"
              type="password"
              value={bootstrapForm.adminToken}
              onChange={(value) => setBootstrapForm((current) => ({ ...current, adminToken: value }))}
              placeholder="choose-a-long-random-token"
            />
            <Input
              label="Default ports"
              value={bootstrapForm.defaultScanPorts}
              onChange={(value) => setBootstrapForm((current) => ({ ...current, defaultScanPorts: value }))}
              placeholder="22,80,443,8080,8443"
            />
            <TextArea
              label="Optional seed CIDRs"
              value={bootstrapForm.seedCIDRs}
              onChange={(value) => setBootstrapForm((current) => ({ ...current, seedCIDRs: value }))}
              placeholder="192.168.1.0/24"
            />
            <label className="rounded-3xl border border-white/10 bg-white/5 p-4 text-sm text-ink">
              <span className="block text-xs uppercase tracking-[0.24em] text-muted">Discovery policy</span>
              <span className="mt-2 flex items-center gap-3">
                <input
                  checked={bootstrapForm.autoScanEnabled}
                  className="h-4 w-4 accent-accent"
                  onChange={(event) => setBootstrapForm((current) => ({ ...current, autoScanEnabled: event.target.checked }))}
                  type="checkbox"
                />
                Enable automatic LAN scans after bootstrap
              </span>
            </label>
            <div className="md:col-span-2 flex items-center justify-between gap-4 rounded-3xl border border-white/10 bg-base/70 p-4">
              <div className="text-sm text-muted">
                The token is required for every write endpoint, including manual services, bookmarks, and discovery settings.
              </div>
              <button className="rounded-full bg-accent px-5 py-3 text-sm font-semibold text-base transition hover:brightness-110" type="submit">
                Initialize
              </button>
            </div>
          </form>
          <Alerts error={error} notice={notice} />
        </section>
      </Shell>
    );
  }

  return (
    <Shell>
      <section className="grid gap-6">
        <header className="animate-floatIn rounded-[2rem] border border-white/10 bg-panel/80 p-6 shadow-halo backdrop-blur">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <p className="text-sm uppercase tracking-[0.35em] text-accent">Homelabwatch</p>
              <h1 className="mt-2 font-display text-4xl font-semibold text-ink">Discover, monitor, and reach everything in the lab.</h1>
              <p className="mt-3 max-w-3xl text-sm leading-7 text-muted">
                The dashboard tracks devices by MAC identity, discovers Docker workloads and LAN services, and streams health changes over a single embedded control plane.
              </p>
            </div>
            <div className="grid gap-3 sm:grid-cols-[minmax(0,18rem)_auto_auto_auto]">
              <Input
                compact
                label="Admin token"
                type="password"
                value={adminToken}
                onChange={setAdminToken}
                placeholder="required for writes"
              />
              <ActionButton onClick={() => runJob("/api/v1/discovery/run", "Discovery run started.")}>Run discovery</ActionButton>
              <ActionButton onClick={() => runJob("/api/v1/monitoring/run", "Health checks started.")}>Run checks</ActionButton>
              <ActionButton onClick={() => Promise.all([loadDashboard(), loadSettings()])}>Refresh</ActionButton>
            </div>
          </div>
          <Alerts error={error} notice={notice} />
          <div className="mt-6 grid gap-4 md:grid-cols-3 xl:grid-cols-6">
            {metrics.map((metric) => (
              <MetricCard key={metric.label} {...metric} />
            ))}
          </div>
        </header>

        <div className="grid gap-6 xl:grid-cols-[1.4fr_0.9fr]">
          <Section title="Services" subtitle="Manual links, Docker workloads, and LAN discoveries in one table.">
            <div className="grid gap-3">
              {services.length === 0 && <EmptyState title="No services yet" body="Run discovery or add a manual endpoint." compact />}
              {services.map((service) => (
                <article key={service.id} className="rounded-3xl border border-white/10 bg-base/70 p-4">
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <h3 className="font-display text-lg font-semibold text-ink">{service.name}</h3>
                        <StatusBadge status={service.status} />
                        <span className="rounded-full border border-white/10 px-3 py-1 text-[11px] uppercase tracking-[0.22em] text-muted">
                          {service.source}
                        </span>
                      </div>
                      <p className="mt-2 text-sm text-muted">{service.url}</p>
                      <p className="mt-1 text-xs uppercase tracking-[0.2em] text-muted/80">
                        {service.deviceName || service.host}:{service.port}
                      </p>
                    </div>
                    <a className="rounded-full border border-accent/50 px-4 py-2 text-sm text-accent transition hover:bg-accent hover:text-base" href={service.url} rel="noreferrer" target="_blank">
                      Open
                    </a>
                  </div>
                  {service.checks?.length > 0 && (
                    <div className="mt-4 flex flex-wrap gap-2 text-xs text-muted">
                      {service.checks.map((check) => (
                        <span key={check.id} className="rounded-full border border-white/10 px-3 py-1">
                          {check.type} {check.lastResult?.status ? `• ${check.lastResult.status}` : "• pending"}
                        </span>
                      ))}
                    </div>
                  )}
                </article>
              ))}
            </div>
            <form className="mt-5 grid gap-3 rounded-3xl border border-dashed border-accent/40 bg-white/5 p-4" onSubmit={saveManualService}>
              <h3 className="font-display text-lg font-semibold text-ink">Add manual service</h3>
              <Input label="Name" value={serviceForm.name} onChange={(value) => setServiceForm((current) => ({ ...current, name: value }))} placeholder="Plex" />
              <Input label="URL" value={serviceForm.url} onChange={(value) => setServiceForm((current) => ({ ...current, url: value }))} placeholder="http://192.168.1.20:32400" />
              <button className="rounded-full bg-accent px-4 py-3 text-sm font-semibold text-base" type="submit">Save service</button>
            </form>
          </Section>

          <Section title="Discovery" subtitle="Seed Docker engines and scan targets without leaving the dashboard.">
            <div className="grid gap-4">
              <CardList title="Docker endpoints" items={dockerEndpoints} renderItem={(item) => (
                <div key={item.id} className="rounded-3xl border border-white/10 bg-base/70 p-4">
                  <p className="font-semibold text-ink">{item.name}</p>
                  <p className="mt-1 text-sm text-muted">{item.address}</p>
                  <p className="mt-2 text-xs uppercase tracking-[0.2em] text-muted/80">
                    {item.enabled ? "enabled" : "disabled"} • every {item.scanIntervalSeconds}s
                  </p>
                </div>
              )} />
              <form className="grid gap-3 rounded-3xl border border-dashed border-accent/40 bg-white/5 p-4" onSubmit={saveDockerEndpoint}>
                <h3 className="font-display text-lg font-semibold text-ink">Add Docker endpoint</h3>
                <Input label="Name" value={dockerForm.name} onChange={(value) => setDockerForm((current) => ({ ...current, name: value }))} />
                <Input label="Kind" value={dockerForm.kind} onChange={(value) => setDockerForm((current) => ({ ...current, kind: value }))} />
                <Input label="Address" value={dockerForm.address} onChange={(value) => setDockerForm((current) => ({ ...current, address: value }))} />
                <Input label="Interval seconds" value={String(dockerForm.scanIntervalSeconds)} onChange={(value) => setDockerForm((current) => ({ ...current, scanIntervalSeconds: value }))} />
                <button className="rounded-full bg-accent px-4 py-3 text-sm font-semibold text-base" type="submit">Save endpoint</button>
              </form>
              <CardList title="Scan targets" items={scanTargets} renderItem={(item) => (
                <div key={item.id} className="rounded-3xl border border-white/10 bg-base/70 p-4">
                  <div className="flex items-center justify-between gap-4">
                    <p className="font-semibold text-ink">{item.name}</p>
                    <span className="text-xs uppercase tracking-[0.2em] text-muted">{item.autoDetected ? "auto" : "manual"}</span>
                  </div>
                  <p className="mt-1 text-sm text-muted">{item.cidr}</p>
                  <p className="mt-2 text-xs uppercase tracking-[0.2em] text-muted/80">
                    ports {item.commonPorts.join(", ")} • every {item.scanIntervalSeconds}s
                  </p>
                </div>
              )} />
              <form className="grid gap-3 rounded-3xl border border-dashed border-accent/40 bg-white/5 p-4" onSubmit={saveScanTarget}>
                <h3 className="font-display text-lg font-semibold text-ink">Add scan target</h3>
                <Input label="Name" value={scanTargetForm.name} onChange={(value) => setScanTargetForm((current) => ({ ...current, name: value }))} />
                <Input label="CIDR" value={scanTargetForm.cidr} onChange={(value) => setScanTargetForm((current) => ({ ...current, cidr: value }))} />
                <Input label="Common ports" value={scanTargetForm.commonPorts} onChange={(value) => setScanTargetForm((current) => ({ ...current, commonPorts: value }))} />
                <button className="rounded-full bg-accent px-4 py-3 text-sm font-semibold text-base" type="submit">Save target</button>
              </form>
            </div>
          </Section>
        </div>

        <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr_0.8fr]">
          <Section title="Devices" subtitle="Tracked by MAC or fallback fingerprint to survive IP churn.">
            <div className="grid gap-3">
              {devices.length === 0 && <EmptyState title="No devices yet" body="LAN scans will populate this list." compact />}
              {devices.map((device) => (
                <div key={device.id} className="rounded-3xl border border-white/10 bg-base/70 p-4">
                  <div className="flex items-center justify-between gap-4">
                    <div>
                      <h3 className="font-display text-lg font-semibold text-ink">{device.displayName || device.hostname || device.identityKey}</h3>
                      <p className="mt-1 text-sm text-muted">{device.primaryMac || device.identityKey}</p>
                    </div>
                    <span className="rounded-full border border-white/10 px-3 py-1 text-[11px] uppercase tracking-[0.22em] text-muted">
                      {device.identityConfidence}
                    </span>
                  </div>
                  <div className="mt-3 grid gap-2 text-sm text-muted">
                    <span>IPs: {device.addresses?.map((item) => item.ipAddress).join(", ") || "n/a"}</span>
                    <span>Ports: {device.ports?.map((item) => `${item.port}/${item.protocol}`).join(", ") || "n/a"}</span>
                    <span>Last seen: {formatDate(device.lastSeenAt)}</span>
                  </div>
                </div>
              ))}
            </div>
          </Section>

          <Section title="Bookmarks" subtitle="User-curated links that live beside auto-discovered services.">
            <div className="grid gap-3">
              {bookmarks.length === 0 && <EmptyState title="No bookmarks yet" body="Add external dashboards or docs here." compact />}
              {bookmarks.map((bookmark) => (
                <a key={bookmark.id} className="rounded-3xl border border-white/10 bg-base/70 p-4 transition hover:border-accent/50" href={bookmark.url} rel="noreferrer" target="_blank">
                  <div className="flex items-center justify-between gap-4">
                    <div>
                      <h3 className="font-display text-lg font-semibold text-ink">{bookmark.name}</h3>
                      <p className="mt-1 text-sm text-muted">{bookmark.url}</p>
                    </div>
                    <span className="text-xs uppercase tracking-[0.2em] text-accent">Open</span>
                  </div>
                </a>
              ))}
            </div>
            <form className="mt-5 grid gap-3 rounded-3xl border border-dashed border-accent/40 bg-white/5 p-4" onSubmit={saveBookmark}>
              <h3 className="font-display text-lg font-semibold text-ink">Add bookmark</h3>
              <Input label="Name" value={bookmarkForm.name} onChange={(value) => setBookmarkForm((current) => ({ ...current, name: value }))} />
              <Input label="URL" value={bookmarkForm.url} onChange={(value) => setBookmarkForm((current) => ({ ...current, url: value }))} />
              <TextArea label="Description" value={bookmarkForm.description} onChange={(value) => setBookmarkForm((current) => ({ ...current, description: value }))} />
              <button className="rounded-full bg-accent px-4 py-3 text-sm font-semibold text-base" type="submit">Save bookmark</button>
            </form>
          </Section>

          <Section title="Workers" subtitle="Recent scheduler outcomes and health events.">
            <div className="grid gap-3">
              {jobState.length === 0 && <EmptyState title="No worker runs yet" body="Background jobs will report here after bootstrap." compact />}
              {jobState.map((job) => (
                <div key={job.jobName} className="rounded-3xl border border-white/10 bg-base/70 p-4">
                  <div className="flex items-center justify-between gap-4">
                    <h3 className="font-semibold text-ink">{job.jobName}</h3>
                    <span className={`text-xs uppercase tracking-[0.2em] ${job.lastError ? "text-danger" : "text-ok"}`}>
                      {job.lastError ? "error" : "ok"}
                    </span>
                  </div>
                  <p className="mt-2 text-sm text-muted">Last run: {formatDate(job.lastRunAt)}</p>
                  {job.lastError ? <p className="mt-1 text-sm text-danger">{job.lastError}</p> : null}
                </div>
              ))}
              <CardList title="Recent events" items={recentEvents} renderItem={(item) => (
                <div key={item.id} className="rounded-3xl border border-white/10 bg-base/70 p-4">
                  <div className="flex items-center justify-between gap-4">
                    <span className="font-semibold text-ink">{item.eventType}</span>
                    <StatusBadge status={item.status} subtle />
                  </div>
                  <p className="mt-2 text-sm text-muted">{item.message}</p>
                  <p className="mt-2 text-xs uppercase tracking-[0.2em] text-muted/80">{formatDate(item.createdAt)}</p>
                </div>
              )} />
            </div>
          </Section>
        </div>
      </section>
    </Shell>
  );
}

function Shell({ children }) {
  return (
    <main className="min-h-screen bg-base px-4 py-8 text-ink sm:px-6 lg:px-10">
      <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(242,196,90,0.14),_transparent_35%),radial-gradient(circle_at_bottom_right,_rgba(82,212,155,0.12),_transparent_32%),linear-gradient(180deg,_rgba(255,255,255,0.03),_transparent_45%)]" />
      <div className="pointer-events-none fixed inset-0 opacity-30 [background-image:linear-gradient(rgba(255,255,255,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.03)_1px,transparent_1px)] [background-size:72px_72px]" />
      <div className="relative mx-auto max-w-7xl">{children}</div>
    </main>
  );
}

function Section({ title, subtitle, children }) {
  return (
    <section className="animate-floatIn rounded-[2rem] border border-white/10 bg-panel/80 p-5 shadow-halo backdrop-blur">
      <div className="mb-5">
        <h2 className="font-display text-2xl font-semibold text-ink">{title}</h2>
        <p className="mt-1 text-sm text-muted">{subtitle}</p>
      </div>
      {children}
    </section>
  );
}

function MetricCard({ label, value, tone }) {
  return (
    <div className="rounded-3xl border border-white/10 bg-base/70 p-4">
      <div className="text-xs uppercase tracking-[0.26em] text-muted">{label}</div>
      <div className={`mt-3 font-display text-4xl font-semibold ${tone}`}>{value}</div>
    </div>
  );
}

function StatusBadge({ status, subtle = false }) {
  const tones = {
    healthy: "border-ok/40 text-ok",
    degraded: "border-warn/40 text-warn",
    unhealthy: "border-danger/40 text-danger",
    unknown: "border-white/15 text-muted",
  };
  return (
    <span className={`rounded-full border px-3 py-1 text-[11px] uppercase tracking-[0.24em] ${subtle ? "bg-transparent" : "bg-white/5"} ${tones[status] || tones.unknown}`}>
      {status || "unknown"}
    </span>
  );
}

function Input({ label, value, onChange, placeholder, type = "text", compact = false }) {
  return (
    <label className={`block ${compact ? "" : "rounded-3xl border border-white/10 bg-white/5 p-4"}`}>
      <span className="block text-xs uppercase tracking-[0.24em] text-muted">{label}</span>
      <input
        className={`mt-2 w-full rounded-2xl border border-white/10 bg-base/80 px-4 py-3 text-sm text-ink outline-none ring-0 transition placeholder:text-muted/60 focus:border-accent/60 ${compact ? "" : ""}`}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        type={type}
        value={value}
      />
    </label>
  );
}

function TextArea({ label, value, onChange, placeholder }) {
  return (
    <label className="rounded-3xl border border-white/10 bg-white/5 p-4">
      <span className="block text-xs uppercase tracking-[0.24em] text-muted">{label}</span>
      <textarea
        className="mt-2 min-h-24 w-full rounded-2xl border border-white/10 bg-base/80 px-4 py-3 text-sm text-ink outline-none placeholder:text-muted/60 focus:border-accent/60"
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        value={value}
      />
    </label>
  );
}

function ActionButton({ children, onClick }) {
  return (
    <button className="rounded-full border border-accent/40 bg-base/70 px-4 py-3 text-sm font-semibold text-accent transition hover:bg-accent hover:text-base" onClick={onClick} type="button">
      {children}
    </button>
  );
}

function EmptyState({ title, body, compact = false }) {
  return (
    <div className={`rounded-3xl border border-dashed border-white/15 bg-base/50 ${compact ? "p-4" : "p-8"}`}>
      <h3 className="font-display text-lg font-semibold text-ink">{title}</h3>
      <p className="mt-2 text-sm text-muted">{body}</p>
    </div>
  );
}

function Alerts({ error, notice }) {
  if (!error && !notice) {
    return null;
  }
  return (
    <div className="mt-4 grid gap-2">
      {notice ? <div className="rounded-2xl border border-ok/30 bg-ok/10 px-4 py-3 text-sm text-ok">{notice}</div> : null}
      {error ? <div className="rounded-2xl border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger">{error}</div> : null}
    </div>
  );
}

function CardList({ title, items, renderItem }) {
  return (
    <div>
      <h3 className="mb-3 font-display text-lg font-semibold text-ink">{title}</h3>
      <div className="grid gap-3">
        {items.length === 0 ? <EmptyState title={`No ${title.toLowerCase()} yet`} body="This section will populate as configuration grows." compact /> : items.map(renderItem)}
      </div>
    </div>
  );
}

function parsePorts(raw) {
  return raw
    .split(",")
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isFinite(item) && item > 0);
}

function parseCIDRTargets(raw, portsRaw) {
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

function formatDate(value) {
  if (!value) {
    return "never";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}
