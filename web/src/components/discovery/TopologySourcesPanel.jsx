import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import { NetworkIcon, PlusIcon, RefreshIcon } from "../ui/Icons";
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
  onDelete,
  onEdit,
  onRun,
}) {
  return (
    <Card>
      <CardHeader
        action={
          <div className="flex flex-wrap gap-2">
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
              Add source
            </Button>
          </div>
        }
        description="SNMP sources used for LLDP and switch MAC-table topology."
        title="Topology sources"
      />
      <CardContent className="p-0">
        {items.length === 0 ? (
          <div className="px-5 py-5 sm:px-6">
            <EmptyState
              action={canManage ? onAdd : undefined}
              actionLabel="Add source"
              body="Add managed switches, routers, APs, or hypervisors to discover observed network links."
              title="No topology sources configured"
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
                      <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-slate-100 text-slate-600">
                        <NetworkIcon className="h-4 w-4" />
                      </span>
                      <div className="min-w-0">
                        <p className="truncate font-medium text-slate-900">{item.name}</p>
                        <p className="truncate text-sm text-slate-500">
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
