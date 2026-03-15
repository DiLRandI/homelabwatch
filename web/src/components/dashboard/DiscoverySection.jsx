import { useEffect, useState } from "react";

import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import { DiscoveryIcon, PlusIcon } from "../ui/Icons";
import StatusBadge from "../ui/StatusBadge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/Table";

function cadenceLabel(value) {
  return `${value || 0}s`;
}

function DiscoverySettingsCard({
  canManage,
  discoverySettings,
  onSaveSettings,
}) {
  const [form, setForm] = useState({
    autoBookmarkMinConfidence: 90,
    autoBookmarkSources: ["docker", "lan", "mdns"],
    bookmarkPolicy: "manual",
  });

  useEffect(() => {
    setForm({
      autoBookmarkMinConfidence:
        discoverySettings?.autoBookmarkMinConfidence || 90,
      autoBookmarkSources:
        discoverySettings?.autoBookmarkSources?.length > 0
          ? discoverySettings.autoBookmarkSources
          : ["docker", "lan", "mdns"],
      bookmarkPolicy: discoverySettings?.bookmarkPolicy || "manual",
    });
  }, [discoverySettings]);

  async function handleSubmit(event) {
    event.preventDefault();
    await onSaveSettings(form);
  }

  function toggleSource(source) {
    setForm((current) => ({
      ...current,
      autoBookmarkSources: current.autoBookmarkSources.includes(source)
        ? current.autoBookmarkSources.filter((item) => item !== source)
        : [...current.autoBookmarkSources, source],
    }));
  }

  return (
    <Card>
      <CardHeader
        description="Control whether high-confidence discoveries stay in review or become bookmarks automatically."
        title="Discovery policy"
      />
      <CardContent>
        <form className="grid gap-4" onSubmit={handleSubmit}>
          <label className="grid gap-2 text-sm font-medium text-slate-700">
            Bookmark policy
            <select
              className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-accent focus:ring-4 focus:ring-accent/10"
              disabled={!canManage}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  bookmarkPolicy: event.target.value,
                }))
              }
              value={form.bookmarkPolicy}
            >
              <option value="manual">Manual review</option>
              <option value="auto_high_confidence">Auto-create high confidence</option>
            </select>
          </label>

          <label className="grid gap-2 text-sm font-medium text-slate-700">
            Auto-bookmark confidence threshold
            <input
              className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-accent focus:ring-4 focus:ring-accent/10"
              disabled={!canManage}
              min="50"
              max="100"
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  autoBookmarkMinConfidence: Number(event.target.value || 90),
                }))
              }
              type="number"
              value={form.autoBookmarkMinConfidence}
            />
          </label>

          <div className="grid gap-3">
            {["docker", "lan", "mdns"].map((source) => (
              <label
                className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700"
                key={source}
              >
                <input
                  checked={form.autoBookmarkSources.includes(source)}
                  disabled={!canManage}
                  onChange={() => toggleSource(source)}
                  type="checkbox"
                />
                Allow auto-bookmarking from {source}
              </label>
            ))}
          </div>

          <div className="flex justify-end">
            <Button disabled={!canManage} type="submit">
              Save policy
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

export default function DiscoverySection({
  canManage = true,
  discoverySettings,
  dockerEndpoints,
  onSaveSettings,
  onAddDockerEndpoint,
  onAddScanTarget,
  scanTargets,
}) {
  return (
    <section className="grid gap-6 xl:grid-cols-3" id="discovery">
      <Card>
        <CardHeader
          action={
            <Button
              disabled={!canManage}
              leadingIcon={PlusIcon}
              onClick={onAddDockerEndpoint}
              variant="secondary"
            >
              Add endpoint
            </Button>
          }
          description="Connected Docker engines scanned for service metadata and health."
          title="Docker endpoints"
        />
        <CardContent className="p-0">
          {dockerEndpoints.length === 0 ? (
            <div className="px-5 py-5 sm:px-6">
              <EmptyState
                action={canManage ? onAddDockerEndpoint : undefined}
                actionLabel="Add endpoint"
                body="Attach a local or remote Docker engine to discover running workloads."
                title="No Docker endpoints configured"
              />
            </div>
          ) : (
            <Table>
              <TableHead>
                <tr>
                  <TableHeader>Endpoint</TableHeader>
                  <TableHeader>State</TableHeader>
                  <TableHeader>Cadence</TableHeader>
                  <TableHeader>Last success</TableHeader>
                </tr>
              </TableHead>
              <TableBody>
                {dockerEndpoints.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell className="min-w-[240px]">
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-slate-100 text-slate-600">
                            <DiscoveryIcon className="h-4 w-4" />
                          </span>
                          <div className="min-w-0">
                            <p className="truncate font-medium text-slate-900">{item.name}</p>
                            <p className="truncate text-sm text-slate-500" title={item.address}>
                              {item.address}
                            </p>
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-2">
                        <StatusBadge status={item.enabled ? "healthy" : "unknown"} />
                        <Badge tone="info">{item.kind}</Badge>
                      </div>
                    </TableCell>
                    <TableCell>{cadenceLabel(item.scanIntervalSeconds)}</TableCell>
                    <TableCell>{formatDate(item.lastSuccessAt)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader
          action={
            <Button
              disabled={!canManage}
              leadingIcon={PlusIcon}
              onClick={onAddScanTarget}
              variant="secondary"
            >
              Add target
            </Button>
          }
          description="CIDR ranges scanned for responsive devices and services."
          title="Scan targets"
        />
        <CardContent className="p-0">
          {scanTargets.length === 0 ? (
            <div className="px-5 py-5 sm:px-6">
              <EmptyState
                action={canManage ? onAddScanTarget : undefined}
                actionLabel="Add target"
                body="Seed the dashboard with one or more subnets to discover devices beyond Docker."
                title="No scan targets configured"
              />
            </div>
          ) : (
            <Table>
              <TableHead>
                <tr>
                  <TableHeader>Target</TableHeader>
                  <TableHeader>Mode</TableHeader>
                  <TableHeader>Ports</TableHeader>
                  <TableHeader>Cadence</TableHeader>
                </tr>
              </TableHead>
              <TableBody>
                {scanTargets.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell className="min-w-[220px]">
                      <p className="font-medium text-slate-900">{item.name}</p>
                      <p className="mt-1 text-sm text-slate-500">{item.cidr}</p>
                    </TableCell>
                    <TableCell>
                      <Badge tone={item.autoDetected ? "neutral" : "accent"}>
                        {item.autoDetected ? "auto-detected" : "manual"}
                      </Badge>
                    </TableCell>
                    <TableCell className="max-w-[240px]">
                      <p className="truncate" title={item.commonPorts.join(", ")}>
                        {item.commonPorts.join(", ")}
                      </p>
                    </TableCell>
                    <TableCell>{cadenceLabel(item.scanIntervalSeconds)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <DiscoverySettingsCard
        canManage={canManage}
        discoverySettings={discoverySettings}
        onSaveSettings={onSaveSettings}
      />
    </section>
  );
}
