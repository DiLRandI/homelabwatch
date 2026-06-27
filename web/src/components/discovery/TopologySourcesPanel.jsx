import { useState } from "react";

import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import { NetworkIcon, PlusIcon, RefreshIcon, SparklesIcon } from "../ui/Icons";
import Input from "../ui/Input";
import StatusBadge from "../ui/StatusBadge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/Table";

function credentialLabel(item) {
  const flags = [];
  if (item.hasCommunity) flags.push("community");
  if (item.hasAuthPassphrase) flags.push("auth");
  if (item.hasPrivacyPassphrase) flags.push("privacy");
  return flags.length ? flags.join(", ") : "none";
}

export default function TopologySourcesPanel({
  canManage,
  items = [],
  onAdd,
  onAutoDiscover,
  onDelete,
  onEdit,
  onRun,
}) {
  const [community, setCommunity] = useState("");
  const [runningAutoDiscovery, setRunningAutoDiscovery] = useState(false);

  async function handleAutoDiscover(event) {
    event?.preventDefault();
    if (!onAutoDiscover || runningAutoDiscovery) {
      return;
    }
    setRunningAutoDiscovery(true);
    try {
      await onAutoDiscover({ community: community.trim() });
    } finally {
      setRunningAutoDiscovery(false);
    }
  }

  return (
    <Card>
      <CardHeader
        action={
          <form
            className="flex flex-wrap items-end gap-2"
            onSubmit={handleAutoDiscover}
          >
            <Input
              autoComplete="off"
              compact
              containerClassName="w-44"
              label="SNMP key"
              onChange={setCommunity}
              placeholder="public"
              value={community}
            />
            <Button
              disabled={!canManage || runningAutoDiscovery}
              leadingIcon={SparklesIcon}
              type="submit"
            >
              Auto-discover
            </Button>
            <Button
              disabled={!canManage}
              leadingIcon={RefreshIcon}
              onClick={onRun}
              variant="secondary"
            >
              Run now
            </Button>
            <Button
              disabled={!canManage}
              leadingIcon={PlusIcon}
              onClick={onAdd}
              variant="secondary"
            >
              Advanced
            </Button>
          </form>
        }
        description="HomelabWatch probes likely routers and switches from scan targets, then uses any SNMP sources it finds for LLDP and switch-port topology."
        title="Topology sources"
      />
      <CardContent className="p-0">
        {items.length === 0 ? (
          <div className="px-5 py-5 sm:px-6">
            <EmptyState
              action={canManage ? handleAutoDiscover : undefined}
              actionLabel="Auto-discover sources"
              body="Run automatic probing first. Use Advanced only when a device needs custom SNMP settings."
              title="No topology sources found yet"
            />
          </div>
        ) : (
          <Table>
            <TableHead>
              <tr>
                <TableHeader>Source</TableHeader>
                <TableHeader>Status</TableHeader>
                <TableHeader>SNMP</TableHeader>
                <TableHeader>Credentials</TableHeader>
                <TableHeader>Last success</TableHeader>
                <TableHeader>Last error</TableHeader>
                <TableHeader></TableHeader>
              </tr>
            </TableHead>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="min-w-[240px]">
                    <div className="flex items-center gap-2">
                      <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-base text-muted">
                        <NetworkIcon className="h-4 w-4" />
                      </span>
                      <div className="min-w-0">
                        <p className="truncate font-medium text-ink">{item.name}</p>
                        <p className="truncate text-sm text-muted">
                          {item.address}:{item.port || 161}
                        </p>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-2">
                      <StatusBadge status={item.enabled ? "healthy" : "unknown"} />
                      {item.root ? <Badge tone="accent">root</Badge> : null}
                      <Badge tone="info">{item.role || "unknown"}</Badge>
                    </div>
                  </TableCell>
                  <TableCell>{item.snmpVersion || "v2c"}</TableCell>
                  <TableCell>{credentialLabel(item)}</TableCell>
                  <TableCell>{formatDate(item.lastSuccessAt)}</TableCell>
                  <TableCell className="max-w-[240px]">
                    <p className="truncate text-warn-strong" title={item.lastError || ""}>
                      {item.lastError || ""}
                    </p>
                  </TableCell>
                  <TableCell>
                    <div className="flex justify-end gap-2">
                      <Button
                        disabled={!canManage}
                        onClick={() => onEdit(item)}
                        size="sm"
                        variant="secondary"
                      >
                        Edit
                      </Button>
                      <Button
                        disabled={!canManage}
                        onClick={() => onDelete(item.id)}
                        size="sm"
                        variant="ghost"
                      >
                        Delete
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
