import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/Table";

export default function DevicesSection({
  devices,
  discoveryCounts = {},
  serviceCounts = {},
}) {
  return (
    <section id="devices">
      <Card>
        <CardHeader
          description="Tracked by MAC identity when available, with network metadata preserved across IP churn."
          title="Devices"
        />
        <CardContent className="p-0">
          {devices.length === 0 ? (
            <div className="px-5 py-5 sm:px-6">
              <EmptyState
                body="Once discovery runs, responsive network devices will appear here with addresses, ports, and confidence."
                title="No devices discovered yet"
              />
            </div>
          ) : (
            <Table>
              <TableHead>
                <tr>
                  <TableHeader>Device</TableHeader>
                  <TableHeader>Addresses</TableHeader>
                  <TableHeader>Ports</TableHeader>
                  <TableHeader>Confidence</TableHeader>
                  <TableHeader>Last seen</TableHeader>
                </tr>
              </TableHead>
              <TableBody>
                {devices.map((device) => (
                  <TableRow key={device.id}>
                    <TableCell className="min-w-[220px]">
                      <p className="font-medium text-slate-900">
                        {device.displayName || device.hostname || device.identityKey}
                      </p>
                      <p className="mt-1 text-sm text-slate-500">
                        {device.primaryMac || device.identityKey}
                      </p>
                      <div className="mt-2 flex flex-wrap gap-2">
                        <Badge tone="info">
                          {serviceCounts[device.id] || 0} accepted services
                        </Badge>
                        <Badge tone="neutral">
                          {discoveryCounts[device.id] || 0} discoveries
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell className="max-w-[260px]">
                      <p
                        className="truncate"
                        title={
                          device.addresses?.map((item) => item.ipAddress).join(", ") || "n/a"
                        }
                      >
                        {device.addresses?.map((item) => item.ipAddress).join(", ") || "n/a"}
                      </p>
                    </TableCell>
                    <TableCell className="max-w-[220px]">
                      <p
                        className="truncate"
                        title={
                          device.ports
                            ?.map((item) => `${item.port}/${item.protocol}`)
                            .join(", ") || "n/a"
                        }
                      >
                        {device.ports?.map((item) => `${item.port}/${item.protocol}`).join(", ") ||
                          "n/a"}
                      </p>
                    </TableCell>
                    <TableCell>
                      <Badge tone={device.identityConfidence === "high" ? "success" : "neutral"}>
                        {device.identityConfidence}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatDate(device.lastSeenAt)}</TableCell>
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
