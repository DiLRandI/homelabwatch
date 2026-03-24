import { useEffect, useState } from "react";

import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import { DatabaseIcon, EditIcon, PlusIcon, RefreshIcon, TrashIcon } from "../ui/Icons";
import Input from "../ui/Input";
import Modal from "../ui/Modal";
import TextArea from "../ui/TextArea";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/Table";

function buildDefaultDefinition() {
  return {
    checkTemplates: [
      {
        addressSource: "literal_host",
        configSource: "definition",
        enabled: true,
        expectedStatusMax: 399,
        expectedStatusMin: 200,
        intervalSeconds: 60,
        method: "GET",
        name: "HTTP health",
        path: "/",
        port: 0,
        protocol: "http",
        sortOrder: 0,
        timeoutSeconds: 10,
        type: "http",
      },
    ],
    enabled: true,
    icon: "",
    id: "",
    key: "",
    matchers: [
      {
        extra: "",
        operator: "exact",
        sortOrder: 0,
        type: "port",
        value: "",
        weight: 80,
      },
    ],
    name: "",
    priority: 100,
  };
}

function hydrateDefinition(definition) {
  if (!definition) {
    return buildDefaultDefinition();
  }
  return {
    ...buildDefaultDefinition(),
    ...definition,
    checkTemplates:
      Array.isArray(definition.checkTemplates) && definition.checkTemplates.length > 0
        ? definition.checkTemplates
        : buildDefaultDefinition().checkTemplates,
    matchers:
      Array.isArray(definition.matchers) && definition.matchers.length > 0
        ? definition.matchers
        : buildDefaultDefinition().matchers,
  };
}

function prettyJSON(value) {
  return JSON.stringify(value, null, 2);
}

function parseJSONArray(raw, label) {
  let value;
  try {
    value = JSON.parse(raw);
  } catch {
    throw new Error(`${label} must be valid JSON.`);
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be a JSON array.`);
  }
  return value;
}

function matcherSummary(definition) {
  if (!Array.isArray(definition?.matchers) || definition.matchers.length === 0) {
    return "No matchers";
  }
  const matcher = definition.matchers[0];
  if (!matcher) {
    return "No matchers";
  }
  const extra = matcher.extra ? `${matcher.extra}:` : "";
  return `${matcher.type} ${extra}${matcher.value}`;
}

function checkSummary(definition) {
  if (!Array.isArray(definition?.checkTemplates) || definition.checkTemplates.length === 0) {
    return "No checks";
  }
  const check = definition.checkTemplates[0];
  if (!check) {
    return "No checks";
  }
  const path = check.path || "/";
  const suffix = check.port ? `:${check.port}` : "";
  return `${String(check.type || "http").toUpperCase()} ${check.protocol || "http"}${suffix}${path}`;
}

function DefinitionEditorModal({
  canManage,
  definition,
  onClose,
  onSave,
  open,
}) {
  const [draft, setDraft] = useState(buildDefaultDefinition());
  const [matchersJSON, setMatchersJSON] = useState(prettyJSON(buildDefaultDefinition().matchers));
  const [checksJSON, setChecksJSON] = useState(prettyJSON(buildDefaultDefinition().checkTemplates));
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!open) {
      return;
    }
    const next = hydrateDefinition(definition);
    setDraft(next);
    setMatchersJSON(prettyJSON(next.matchers));
    setChecksJSON(prettyJSON(next.checkTemplates));
    setError("");
    setSaving(false);
  }, [definition, open]);

  const readOnly = Boolean(definition?.builtIn) || !canManage;

  async function handleSave() {
    try {
      setSaving(true);
      setError("");
      const payload = {
        checkTemplates: parseJSONArray(checksJSON, "Check templates"),
        enabled: Boolean(draft.enabled),
        icon: draft.icon || "",
        id: draft.id || undefined,
        key: draft.key || "",
        matchers: parseJSONArray(matchersJSON, "Matchers"),
        name: (draft.name || "").trim(),
        priority: Number(draft.priority || 0),
      };
      const saved = await onSave(payload);
      if (saved) {
        onClose();
      }
    } catch (saveError) {
      setError(saveError.message || "Unable to save service definition.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal
      className="max-w-5xl"
      description={
        definition?.builtIn
          ? "Built-in definitions are read-only. Inspect the matchers and default health templates, then use reapply to push them onto auto-managed services."
          : "Custom definitions let you fingerprint services and set default health endpoints without touching each service card."
      }
      onClose={onClose}
      open={open}
      title={definition?.id ? definition.name || "Service definition" : "New service definition"}
    >
      <div className="grid gap-6">
        <div className="grid gap-4 lg:grid-cols-2">
          <Input
            disabled={readOnly}
            label="Name"
            onChange={(value) => setDraft((current) => ({ ...current, name: value }))}
            placeholder="Grafana"
            value={draft.name}
          />
          <Input
            disabled={readOnly}
            label="Key"
            onChange={(value) => setDraft((current) => ({ ...current, key: value }))}
            placeholder="grafana"
            value={draft.key}
          />
          <Input
            disabled={readOnly}
            label="Icon"
            onChange={(value) => setDraft((current) => ({ ...current, icon: value }))}
            placeholder="grafana"
            value={draft.icon}
          />
          <Input
            disabled={readOnly}
            label="Priority"
            min="0"
            onChange={(value) =>
              setDraft((current) => ({ ...current, priority: Number(value || 0) }))
            }
            type="number"
            value={draft.priority}
          />
          <label className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-700">
            <input
              checked={Boolean(draft.enabled)}
              disabled={readOnly}
              onChange={(event) =>
                setDraft((current) => ({ ...current, enabled: event.target.checked }))
              }
              type="checkbox"
            />
            Enable this definition
          </label>
          <div className="rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-600">
            <p className="font-medium text-slate-900">Editing guidance</p>
            <p className="mt-2 leading-6">
              Matchers and check templates use the API JSON shape. Example matcher:
              <code className="mx-1 rounded bg-slate-100 px-1.5 py-0.5 text-xs">
                {"{\"type\":\"port\",\"value\":\"3000\",\"weight\":80}"}
              </code>
            </p>
          </div>
        </div>

        <div className="grid gap-4 lg:grid-cols-2">
          <TextArea
            label="Matchers JSON"
            onChange={setMatchersJSON}
            rows={14}
            value={matchersJSON}
          />
          <TextArea
            label="Check templates JSON"
            onChange={setChecksJSON}
            rows={14}
            value={checksJSON}
          />
        </div>

        {error ? (
          <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
            {error}
          </div>
        ) : null}

        <div className="flex flex-wrap justify-end gap-3">
          <Button onClick={onClose} variant="ghost">
            Close
          </Button>
          {readOnly ? null : (
            <Button disabled={saving} onClick={handleSave}>
              {saving ? "Saving..." : "Save definition"}
            </Button>
          )}
        </div>
      </div>
    </Modal>
  );
}

export default function ServiceDefinitionsSection({
  canManage = true,
  definitions = [],
  onDeleteDefinition,
  onReapplyDefinition,
  onSaveDefinition,
}) {
  const [selectedDefinition, setSelectedDefinition] = useState(null);

  async function handleDelete(definition) {
    if (!definition || definition.builtIn) {
      return;
    }
    if (!window.confirm(`Delete the custom service definition "${definition.name}"?`)) {
      return;
    }
    await onDeleteDefinition(definition.id);
  }

  return (
    <>
      <section id="service-definitions">
        <Card>
          <CardHeader
            action={
              <Button
                disabled={!canManage}
                leadingIcon={PlusIcon}
                onClick={() => setSelectedDefinition(buildDefaultDefinition())}
              >
                Add definition
              </Button>
            }
            description="Control how services are fingerprinted and which health checks auto-attach when a match is found."
            title="Service definitions"
          />
          <CardContent className="grid gap-5">
            {definitions.length === 0 ? (
              <EmptyState
                action={canManage ? () => setSelectedDefinition(buildDefaultDefinition()) : undefined}
                actionLabel="Create definition"
                body="Start with a custom definition to map a service signature to one or more default health checks."
                title="No service definitions loaded"
              />
            ) : (
              <Table>
                <TableHead>
                  <tr>
                    <TableHeader>Definition</TableHeader>
                    <TableHeader>Matchers</TableHeader>
                    <TableHeader>Default checks</TableHeader>
                    <TableHeader>Updated</TableHeader>
                    <TableHeader className="text-right">Actions</TableHeader>
                  </tr>
                </TableHead>
                <TableBody>
                  {definitions.map((definition) => (
                    <TableRow key={definition.id}>
                      <TableCell className="min-w-[260px]">
                        <div>
                          <p className="font-medium text-slate-900">{definition.name}</p>
                          <div className="mt-2 flex flex-wrap gap-2">
                            <Badge tone={definition.builtIn ? "info" : "accent"}>
                              {definition.builtIn ? "built-in" : "custom"}
                            </Badge>
                            <Badge tone={definition.enabled ? "neutral" : "warning"}>
                              {definition.enabled ? "enabled" : "disabled"}
                            </Badge>
                            <Badge>priority {definition.priority}</Badge>
                            <Badge>{definition.key}</Badge>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <p className="font-medium text-slate-900">{matcherSummary(definition)}</p>
                        <p className="mt-1 text-xs text-slate-500">
                          {definition.matchers?.length || 0} matcher(s)
                        </p>
                      </TableCell>
                      <TableCell>
                        <p className="font-medium text-slate-900">{checkSummary(definition)}</p>
                        <p className="mt-1 text-xs text-slate-500">
                          {definition.checkTemplates?.length || 0} template(s)
                        </p>
                      </TableCell>
                      <TableCell>{formatDate(definition.updatedAt)}</TableCell>
                      <TableCell className="text-right">
                        <div className="flex flex-wrap justify-end gap-2">
                          <Button
                            leadingIcon={EditIcon}
                            onClick={() => setSelectedDefinition(hydrateDefinition(definition))}
                            size="sm"
                            variant="ghost"
                          >
                            {definition.builtIn ? "Inspect" : "Edit"}
                          </Button>
                          <Button
                            disabled={!canManage}
                            leadingIcon={RefreshIcon}
                            onClick={() => void onReapplyDefinition(definition.id)}
                            size="sm"
                            variant="secondary"
                          >
                            Reapply
                          </Button>
                          {definition.builtIn ? null : (
                            <Button
                              disabled={!canManage}
                              leadingIcon={TrashIcon}
                              onClick={() => void handleDelete(definition)}
                              size="sm"
                              variant="ghost"
                            >
                              Delete
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}

            <div className="rounded-3xl border border-slate-200 bg-slate-50 p-5">
              <div className="flex items-start gap-3">
                <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-accent/10 text-accent-strong">
                  <DatabaseIcon className="h-5 w-5" />
                </span>
                <div>
                  <p className="text-sm font-medium text-slate-900">Custom registry workflow</p>
                  <p className="mt-2 text-sm leading-6 text-slate-600">
                    Use higher priority values to let custom rules outrank built-ins. Reapply updates
                    auto-managed services without touching services that users already customized in
                    the health modal.
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      <DefinitionEditorModal
        canManage={canManage}
        definition={selectedDefinition}
        onClose={() => setSelectedDefinition(null)}
        onSave={onSaveDefinition}
        open={Boolean(selectedDefinition)}
      />
    </>
  );
}
