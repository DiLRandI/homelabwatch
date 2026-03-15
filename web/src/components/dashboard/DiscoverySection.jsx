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

export default function DiscoverySection({
  dockerEndpoints,
  onAddDockerEndpoint,
  onAddScanTarget,
  scanTargets,
}) {
  return (
    <section className="grid gap-6 xl:grid-cols-2" id="discovery">
      <Card>
        <CardHeader
          action={
            <Button leadingIcon={PlusIcon} onClick={onAddDockerEndpoint} variant="secondary">
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
                action={onAddDockerEndpoint}
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
            <Button leadingIcon={PlusIcon} onClick={onAddScanTarget} variant="secondary">
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
                action={onAddScanTarget}
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
    </section>
  );
}
