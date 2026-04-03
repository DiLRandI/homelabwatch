import { useEffect, useMemo, useState } from "react";

import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import EmptyState from "../ui/EmptyState";
import Input from "../ui/Input";
import Modal from "../ui/Modal";
import { EditIcon, PlusIcon, ShieldIcon, TrashIcon } from "../ui/Icons";
import EndpointTester from "./EndpointTester";
import HealthStatusBadge from "./HealthStatusBadge";

const CHECK_LABELS = {
  http: "HTTP",
  ping: "Ping",
  tcp: "TCP",
};

const ADDRESS_SOURCE_OPTIONS = [
  { label: "Literal host", value: "literal_host" },
  { label: "Device primary IP", value: "device_primary_ip" },
  { label: "mDNS hostname", value: "mdns_hostname" },
];

const MATCH_MODE_LABELS = {
  auto: "auto",
  custom: "custom",
};

function nativeSelectClass(disabled = false) {
  return [
    "w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-hidden transition",
    "focus:border-accent focus-visible:ring-4 focus-visible:ring-accent/15",
    disabled ? "cursor-not-allowed opacity-60" : "",
  ]
    .filter(Boolean)
    .join(" ");
}

function makeDraftKey() {
  return `draft_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
}

function firstDefined(...values) {
  for (const value of values) {
    if (value !== undefined && value !== null && value !== "") {
      return value;
    }
  }
  return "";
}

function defaultCheckType(service) {
  const scheme = (service?.healthScheme || service?.scheme || "").toLowerCase();
  if (scheme === "http" || scheme === "https") {
    return "http";
  }
  return "tcp";
}

function buildDefaultCheck(service, type = defaultCheckType(service), sortOrder = 0) {
  const protocol =
    type === "http"
      ? firstDefined(service?.healthScheme, service?.scheme, "http")
      : "";

  return {
    addressSource: firstDefined(
      service?.healthAddressSource,
      service?.addressSource,
      "literal_host",
    ),
    configSource: "user",
    draftKey: makeDraftKey(),
    enabled: true,
    expectedStatusMax: type === "http" ? 399 : 0,
    expectedStatusMin: type === "http" ? 200 : 0,
    host: firstDefined(
      service?.healthHost,
      service?.healthHostValue,
      service?.host,
      service?.hostValue,
    ),
    hostValue: firstDefined(
      service?.healthHostValue,
      service?.healthHost,
      service?.hostValue,
      service?.host,
    ),
    id: "",
    intervalSeconds: 60,
    lastResult: null,
    method: "GET",
    name: `${service?.name || "Service"} ${CHECK_LABELS[type] || "check"}`,
    path: type === "http" ? service?.healthPath || service?.path || "" : "",
    port: Number(service?.healthPort || service?.port || 0),
    protocol,
    serviceDefinitionId: "",
    sortOrder,
    target: "",
    timeoutSeconds: 10,
    type,
  };
}

function hydrateCheck(check, service, index) {
  const fallback = buildDefaultCheck(service, check?.type || defaultCheckType(service), index);

  return {
    ...fallback,
    ...check,
    addressSource: firstDefined(check?.addressSource, fallback.addressSource),
    draftKey: check?.id || makeDraftKey(),
    host: firstDefined(check?.host, check?.hostValue, fallback.host),
    hostValue: firstDefined(check?.hostValue, check?.host, fallback.hostValue),
    method: firstDefined(check?.method, fallback.method),
    path: check?.type === "http" ? firstDefined(check?.path, fallback.path) : "",
    protocol:
      check?.type === "http"
        ? firstDefined(check?.protocol, fallback.protocol)
        : "",
  };
}

function normalizeChecks(items, service) {
  if (!Array.isArray(items) || items.length === 0) {
    return [buildDefaultCheck(service)];
  }
  return items.map((item, index) => hydrateCheck(item, service, index));
}

function checkKey(check) {
  return check?.id || check?.draftKey || "";
}

function sanitizeCheck(check, serviceId) {
  return {
    addressSource: check.addressSource || "literal_host",
    configSource: check.configSource || "user",
    enabled: Boolean(check.enabled),
    expectedStatusMax:
      check.type === "http" ? Number(check.expectedStatusMax || 0) : 0,
    expectedStatusMin:
      check.type === "http" ? Number(check.expectedStatusMin || 0) : 0,
    host: check.hostValue || check.host || "",
    hostValue: check.hostValue || check.host || "",
    id: check.id || undefined,
    intervalSeconds: Number(check.intervalSeconds || 60),
    method: check.type === "http" ? firstDefined(check.method, "GET") : "",
    name: (check.name || "").trim(),
    path: check.type === "http" ? check.path || "" : "",
    port: Number(check.port || 0),
    protocol: check.type === "http" ? firstDefined(check.protocol, "http") : "",
    serviceDefinitionId: check.serviceDefinitionId || "",
    serviceId,
    sortOrder: Number(check.sortOrder || 0),
    subjectId: serviceId,
    subjectType: "service",
    type: check.type,
  };
}

function Field({ children, label }) {
  return (
    <label className="grid gap-2">
      <span className="text-sm font-medium text-slate-700">{label}</span>
      {children}
    </label>
  );
}

function CheckRow({ active, check, onSelect }) {
  return (
    <button
      className={[
        "w-full rounded-2xl border px-4 py-4 text-left transition",
        active
          ? "border-accent/25 bg-accent/5 shadow-sm"
          : "border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50",
      ].join(" ")}
      onClick={onSelect}
      type="button"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-slate-950">
            {check.name || "Untitled check"}
          </p>
          <p className="mt-1 truncate text-xs uppercase tracking-[0.16em] text-slate-500">
            {CHECK_LABELS[check.type] || check.type}
          </p>
        </div>
        <HealthStatusBadge result={check.lastResult} status="unknown" subtle />
      </div>
      <div className="mt-3 flex flex-wrap gap-2">
        <Badge>{check.configSource || "user"}</Badge>
        {check.serviceDefinitionId ? <Badge tone="info">definition</Badge> : null}
        {check.enabled ? null : <Badge tone="warning">disabled</Badge>}
      </div>
    </button>
  );
}

export default function HealthSettingsModal({
  canManage = true,
  onClose,
  onDeleteCheck,
  onFetchChecks,
  onSaveCheck,
  onTestCheck,
  open,
  service,
}) {
  const [checks, setChecks] = useState([]);
  const [loadingChecks, setLoadingChecks] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [selectedKey, setSelectedKey] = useState("");
  const [testerResult, setTesterResult] = useState(null);

  useEffect(() => {
    if (!open || !service?.id) {
      return undefined;
    }

    let cancelled = false;

    async function loadChecks() {
      setLoadingChecks(true);
      setTesterResult(null);
      const items = await onFetchChecks(service.id);
      if (cancelled) {
        return;
      }
      const nextChecks = normalizeChecks(items, service);
      setChecks(nextChecks);
      setSelectedKey((current) => {
        if (current && nextChecks.some((item) => checkKey(item) === current)) {
          return current;
        }
        return checkKey(nextChecks[0]);
      });
      setLoadingChecks(false);
    }

    void loadChecks();

    return () => {
      cancelled = true;
    };
  }, [onFetchChecks, open, service]);

  useEffect(() => {
    if (!open) {
      setChecks([]);
      setSelectedKey("");
      setTesterResult(null);
    }
  }, [open]);

  const selectedCheck = useMemo(
    () => checks.find((item) => checkKey(item) === selectedKey) || null,
    [checks, selectedKey],
  );

  function patchSelectedCheck(patch) {
    if (!selectedCheck) {
      return;
    }
    setChecks((current) =>
      current.map((item) =>
        checkKey(item) === selectedKey ? { ...item, ...patch } : item,
      ),
    );
  }

  function createAndSelectCheck(type) {
    const draft = buildDefaultCheck(service, type, checks.length);
    setChecks((current) => [...current, draft]);
    setSelectedKey(checkKey(draft));
    setTesterResult(null);
  }

  async function reloadChecks(preferredKey = "") {
    if (!service?.id) {
      return;
    }
    setLoadingChecks(true);
    const items = await onFetchChecks(service.id);
    const nextChecks = normalizeChecks(items, service);
    setChecks(nextChecks);
    setSelectedKey((current) => {
      const targetKey = preferredKey || current;
      if (targetKey && nextChecks.some((item) => checkKey(item) === targetKey)) {
        return targetKey;
      }
      return checkKey(nextChecks[0]);
    });
    setLoadingChecks(false);
  }

  async function handleSave() {
    if (!selectedCheck || !service?.id) {
      return;
    }
    setSaving(true);
    const saved = await onSaveCheck(service.id, sanitizeCheck(selectedCheck, service.id));
    setSaving(false);
    if (!saved) {
      return;
    }
    await reloadChecks(saved.id);
  }

  async function handleDelete() {
    if (!selectedCheck) {
      return;
    }
    if (!selectedCheck.id) {
      const remaining = checks.filter((item) => checkKey(item) !== selectedKey);
      setChecks(remaining);
      setSelectedKey(checkKey(remaining[0]) || "");
      setTesterResult(null);
      return;
    }
    setSaving(true);
    const removed = await onDeleteCheck(selectedCheck.id);
    setSaving(false);
    if (!removed) {
      return;
    }
    setTesterResult(null);
    await reloadChecks();
  }

  async function handleTest() {
    if (!selectedCheck || !service?.id) {
      return;
    }
    setTesting(true);
    const result = await onTestCheck(service.id, {
      check: sanitizeCheck(selectedCheck, service.id),
      discoverPaths: selectedCheck.type === "http" && !selectedCheck.path,
    });
    setTesting(false);
    if (result) {
      setTesterResult(result);
    }
  }

  return (
    <Modal
      className="max-w-7xl"
      description={
        service
          ? `Tune one or more health probes for ${service.name}. Custom edits switch this service from ${MATCH_MODE_LABELS[service.healthConfigMode] || "auto"} definition defaults to user-managed checks.`
          : ""
      }
      onClose={onClose}
      open={open}
      title="Health check settings"
    >
      {!service ? null : (
        <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
          <aside className="grid gap-4">
            <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4">
              <p className="text-sm font-semibold text-slate-950">{service.name}</p>
              <p className="mt-1 text-sm text-slate-500">Open URL: {service.url}</p>
              {service.healthUrl && service.healthUrl !== service.url ? (
                <p className="mt-1 text-sm text-slate-500">
                  Health URL: {service.healthUrl}
                </p>
              ) : null}
              <div className="mt-3 flex flex-wrap gap-2">
                <Badge tone="info">{service.source}</Badge>
                <Badge>{service.serviceDefinitionId || "custom target"}</Badge>
                <Badge tone={service.healthConfigMode === "custom" ? "accent" : "neutral"}>
                  {MATCH_MODE_LABELS[service.healthConfigMode] || "auto"}
                </Badge>
              </div>
            </div>

            <div className="flex flex-wrap gap-2">
              <Button
                disabled={!canManage}
                leadingIcon={PlusIcon}
                onClick={() => createAndSelectCheck("http")}
                size="sm"
                variant="secondary"
              >
                HTTP
              </Button>
              <Button
                disabled={!canManage}
                leadingIcon={PlusIcon}
                onClick={() => createAndSelectCheck("tcp")}
                size="sm"
                variant="secondary"
              >
                TCP
              </Button>
              <Button
                disabled={!canManage}
                leadingIcon={PlusIcon}
                onClick={() => createAndSelectCheck("ping")}
                size="sm"
                variant="secondary"
              >
                Ping
              </Button>
            </div>

            <div className="grid gap-3">
              {checks.map((check) => (
                <CheckRow
                  active={checkKey(check) === selectedKey}
                  check={check}
                  key={checkKey(check)}
                  onSelect={() => {
                    setSelectedKey(checkKey(check));
                    setTesterResult(null);
                  }}
                />
              ))}
            </div>
          </aside>

          <div className="grid gap-6">
            {loadingChecks ? (
              <EmptyState
                body="Loading the latest saved checks for this service."
                title="Loading health checks"
              />
            ) : !selectedCheck ? (
              <EmptyState
                body="Create or select a check to edit protocol, endpoint, and scheduling details."
                title="No check selected"
              />
            ) : (
              <>
                <div className="rounded-3xl border border-slate-200 bg-white">
                  <div className="flex flex-wrap items-start justify-between gap-4 border-b border-slate-200 px-5 py-5">
                    <div>
                      <div className="flex items-center gap-3">
                        <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-accent/10 text-accent-strong">
                          <EditIcon className="h-5 w-5" />
                        </span>
                        <div>
                          <h3 className="text-lg font-semibold tracking-tight text-slate-950">
                            {selectedCheck.name || "Untitled check"}
                          </h3>
                          <p className="mt-1 text-sm text-slate-500">
                            {CHECK_LABELS[selectedCheck.type] || selectedCheck.type} probe
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      <Button
                        disabled={!canManage || saving}
                        onClick={handleDelete}
                        size="sm"
                        variant="ghost"
                      >
                        {selectedCheck.id ? "Delete check" : "Discard draft"}
                      </Button>
                      <Button
                        disabled={!canManage || testing}
                        leadingIcon={ShieldIcon}
                        onClick={handleTest}
                        size="sm"
                        variant="secondary"
                      >
                        {testing ? "Testing..." : "Test endpoint"}
                      </Button>
                      <Button
                        disabled={!canManage || saving}
                        onClick={handleSave}
                        size="sm"
                      >
                        {saving ? "Saving..." : "Save"}
                      </Button>
                    </div>
                  </div>

                  <div className="grid gap-6 px-5 py-5">
                    <div className="grid gap-4 lg:grid-cols-2">
                      <Input
                        label="Check name"
                        onChange={(value) => patchSelectedCheck({ name: value })}
                        value={selectedCheck.name}
                      />
                      <Field label="Check type">
                        <select
                          className={nativeSelectClass(!canManage)}
                          disabled={!canManage}
                          onChange={(event) => {
                            const type = event.target.value;
                            const template = buildDefaultCheck(service, type, selectedCheck.sortOrder);
                            patchSelectedCheck({
                              expectedStatusMax: template.expectedStatusMax,
                              expectedStatusMin: template.expectedStatusMin,
                              method: template.method,
                              path: template.path,
                              protocol: template.protocol,
                              type,
                            });
                            setTesterResult(null);
                          }}
                          value={selectedCheck.type}
                        >
                          <option value="http">HTTP</option>
                          <option value="tcp">TCP</option>
                          <option value="ping">Ping</option>
                        </select>
                      </Field>
                      <Field label="Address source">
                        <select
                          className={nativeSelectClass(!canManage)}
                          disabled={!canManage}
                          onChange={(event) =>
                            patchSelectedCheck({ addressSource: event.target.value })
                          }
                          value={selectedCheck.addressSource || "literal_host"}
                        >
                          {ADDRESS_SOURCE_OPTIONS.map((option) => (
                            <option key={option.value} value={option.value}>
                              {option.label}
                            </option>
                          ))}
                        </select>
                      </Field>
                      <label className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-700">
                        <input
                          checked={Boolean(selectedCheck.enabled)}
                          disabled={!canManage}
                          onChange={(event) =>
                            patchSelectedCheck({ enabled: event.target.checked })
                          }
                          type="checkbox"
                        />
                        Enable this check
                      </label>
                    </div>

                    <div className="grid gap-4 lg:grid-cols-2">
                      {selectedCheck.type === "http" ? (
                        <Field label="Protocol">
                          <select
                            className={nativeSelectClass(!canManage)}
                            disabled={!canManage}
                            onChange={(event) =>
                              patchSelectedCheck({ protocol: event.target.value })
                            }
                            value={selectedCheck.protocol || "http"}
                          >
                            <option value="http">http</option>
                            <option value="https">https</option>
                          </select>
                        </Field>
                      ) : null}

                      <Input
                        label="Host"
                        onChange={(value) =>
                          patchSelectedCheck({ host: value, hostValue: value })
                        }
                        placeholder="192.168.1.252"
                        value={selectedCheck.hostValue || selectedCheck.host}
                      />

                      {selectedCheck.type !== "ping" ? (
                        <Input
                          label="Port"
                          min="0"
                          onChange={(value) =>
                            patchSelectedCheck({ port: Number(value || 0) })
                          }
                          type="number"
                          value={selectedCheck.port}
                        />
                      ) : null}

                      {selectedCheck.type === "http" ? (
                        <Input
                          label="Path"
                          onChange={(value) => patchSelectedCheck({ path: value })}
                          placeholder="/api/health"
                          value={selectedCheck.path}
                        />
                      ) : null}

                      {selectedCheck.type === "http" ? (
                        <Field label="Method">
                          <select
                            className={nativeSelectClass(!canManage)}
                            disabled={!canManage}
                            onChange={(event) =>
                              patchSelectedCheck({ method: event.target.value })
                            }
                            value={selectedCheck.method || "GET"}
                          >
                            <option value="GET">GET</option>
                            <option value="HEAD">HEAD</option>
                            <option value="POST">POST</option>
                          </select>
                        </Field>
                      ) : null}

                      {selectedCheck.type === "http" ? (
                        <div className="grid gap-4 sm:grid-cols-2">
                          <Input
                            label="Expected status min"
                            min="0"
                            onChange={(value) =>
                              patchSelectedCheck({
                                expectedStatusMin: Number(value || 0),
                              })
                            }
                            type="number"
                            value={selectedCheck.expectedStatusMin}
                          />
                          <Input
                            label="Expected status max"
                            min="0"
                            onChange={(value) =>
                              patchSelectedCheck({
                                expectedStatusMax: Number(value || 0),
                              })
                            }
                            type="number"
                            value={selectedCheck.expectedStatusMax}
                          />
                        </div>
                      ) : null}

                      <Input
                        label="Interval (seconds)"
                        min="5"
                        onChange={(value) =>
                          patchSelectedCheck({ intervalSeconds: Number(value || 0) })
                        }
                        type="number"
                        value={selectedCheck.intervalSeconds}
                      />

                      <Input
                        label="Timeout (seconds)"
                        min="1"
                        onChange={(value) =>
                          patchSelectedCheck({ timeoutSeconds: Number(value || 0) })
                        }
                        type="number"
                        value={selectedCheck.timeoutSeconds}
                      />
                    </div>

                    <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                        Latest result
                      </p>
                      {selectedCheck.lastResult ? (
                        <div className="mt-3 grid gap-3">
                          <HealthStatusBadge
                            result={selectedCheck.lastResult}
                            showCheckedAt
                          />
                          <div className="grid gap-3 lg:grid-cols-2">
                            <div>
                              <p className="text-sm font-medium text-slate-900">
                                Resolved target
                              </p>
                              <p className="mt-1 break-words text-sm text-slate-500">
                                {selectedCheck.lastResult.resolvedTarget ||
                                  selectedCheck.target ||
                                  "Not available"}
                              </p>
                            </div>
                            <div>
                              <p className="text-sm font-medium text-slate-900">
                                Message
                              </p>
                              <p className="mt-1 text-sm text-slate-500">
                                {selectedCheck.lastResult.message || "No message recorded"}
                              </p>
                            </div>
                          </div>
                        </div>
                      ) : (
                        <p className="mt-3 text-sm text-slate-500">
                          This check has not completed a run yet.
                        </p>
                      )}
                    </div>
                  </div>
                </div>

                <EndpointTester loading={testing} result={testerResult} />
              </>
            )}
          </div>
        </div>
      )}
    </Modal>
  );
}
