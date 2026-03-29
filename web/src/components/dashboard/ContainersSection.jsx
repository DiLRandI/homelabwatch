import { formatDate } from "../../lib/format";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import { ArrowUpRightIcon, DiscoveryIcon } from "../ui/Icons";
import StatusBadge from "../ui/StatusBadge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/Table";

function detailValue(service, key, fallback = "n/a") {
  return service?.details?.[key] || fallback;
}

export default function ContainersSection({ containers }) {
  return (
    <section id="containers">
      <Card>
        <CardHeader
          description="Running workloads discovered from attached Docker endpoints."
          title="Containers"
        />
        <CardContent className="p-0">
          {containers.length === 0 ? (
            <div className="px-5 py-5 sm:px-6">
              <EmptyState
                body="Attach a Docker endpoint or run discovery to turn containers into first-class operational inventory."
                title="No running containers discovered"
              />
            </div>
          ) : (
            <Table>
              <TableHead>
                <tr>
                  <TableHeader>Container</TableHeader>
                  <TableHeader>Image</TableHeader>
                  <TableHeader>Endpoint</TableHeader>
                  <TableHeader>Status</TableHeader>
                  <TableHeader>Last seen</TableHeader>
                  <TableHeader className="text-right">Open</TableHeader>
                </tr>
              </TableHead>
              <TableBody>
                {containers.map((service) => (
                  <TableRow key={service.id}>
                    <TableCell className="min-w-[220px]">
                      <div className="flex items-center gap-3">
                        <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-sky-500/12 text-sky-300">
                          <DiscoveryIcon className="h-4 w-4" />
                        </span>
                        <div className="min-w-0">
                          <p className="truncate font-medium text-slate-900">
                            {detailValue(service, "containerName", service.name)}
                          </p>
                          <p className="truncate text-sm text-slate-500">
                            {service.host}:{service.port}
                          </p>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="max-w-[240px]">
                      <p className="truncate" title={detailValue(service, "image")}>
                        {detailValue(service, "image")}
                      </p>
                    </TableCell>
                    <TableCell className="max-w-[220px]">
                      <p className="truncate" title={detailValue(service, "endpoint")}>
                        {detailValue(service, "endpoint")}
                      </p>
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={service.status} />
                    </TableCell>
                    <TableCell>{formatDate(service.lastSeenAt)}</TableCell>
                    <TableCell className="text-right">
                      <Button
                        onClick={() =>
                          window.open(service.url, "_blank", "noopener,noreferrer")
                        }
                        size="sm"
                        trailingIcon={ArrowUpRightIcon}
                        variant="ghost"
                      >
                        Open
                      </Button>
                    </TableCell>
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
