import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import {
  ActivityIcon,
  ArrowUpRightIcon,
  DiscoveryIcon,
  PlusIcon,
  ServicesIcon,
} from "../ui/Icons";
import StatusBadge from "../ui/StatusBadge";

const sourceMeta = {
  docker: {
    icon: DiscoveryIcon,
    iconTone: "bg-sky-50 text-sky-700",
    tone: "info",
  },
  lan: {
    icon: ActivityIcon,
    iconTone: "bg-slate-100 text-slate-700",
    tone: "neutral",
  },
  manual: {
    icon: ServicesIcon,
    iconTone: "bg-accent/10 text-accent-strong",
    tone: "accent",
  },
};

export default function ServicesSection({ onAdd, services }) {
  return (
    <section id="services">
      <Card>
        <CardHeader
          action={
            <Button leadingIcon={PlusIcon} onClick={onAdd}>
              Add service
            </Button>
          }
          description="Tracked endpoints from Docker, manual entry, and network discovery."
          title="Services"
        />
        <CardContent>
          {services.length === 0 ? (
            <EmptyState
              action={onAdd}
              actionLabel="Add your first service"
              body="Run discovery or create a manual service to start monitoring the control plane."
              title="No services connected yet"
            />
          ) : (
            <div className="grid gap-4 xl:grid-cols-2">
              {services.map((service) => {
                const meta = sourceMeta[service.source] || sourceMeta.manual;
                const Icon = meta.icon;

                return (
                  <article
                    className="flex h-full flex-col rounded-3xl border border-slate-200 bg-slate-50 p-5 transition hover:border-slate-300 hover:shadow-card"
                    key={service.id}
                  >
                    <div className="flex items-start justify-between gap-4">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className={`inline-flex h-11 w-11 items-center justify-center rounded-2xl ${meta.iconTone}`}>
                            <Icon className="h-5 w-5" />
                          </span>
                          <div className="min-w-0">
                            <h3 className="truncate text-lg font-semibold tracking-tight text-slate-950">
                              {service.name}
                            </h3>
                            <p
                              className="mt-1 truncate text-sm text-slate-500"
                              title={service.url}
                            >
                              {service.url}
                            </p>
                          </div>
                        </div>
                      </div>
                      <Button
                        className="shrink-0"
                        onClick={() => window.open(service.url, "_blank", "noopener,noreferrer")}
                        size="sm"
                        trailingIcon={ArrowUpRightIcon}
                        variant="secondary"
                      >
                        Open
                      </Button>
                    </div>

                    <div className="mt-4 flex flex-wrap items-center gap-2">
                      <StatusBadge status={service.status} />
                      <Badge tone={meta.tone}>{service.source}</Badge>
                      <Badge>{service.deviceName || service.host || "Unassigned"}</Badge>
                    </div>

                    <dl className="mt-5 grid gap-3 text-sm text-slate-600 sm:grid-cols-2">
                      <div className="rounded-2xl border border-white bg-white px-4 py-3">
                        <dt className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">
                          Endpoint
                        </dt>
                        <dd className="mt-2 truncate font-medium text-slate-900">
                          {service.host}:{service.port}
                        </dd>
                      </div>
                      <div className="rounded-2xl border border-white bg-white px-4 py-3">
                        <dt className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">
                          Last seen
                        </dt>
                        <dd className="mt-2 font-medium text-slate-900">
                          {formatDate(service.lastSeenAt)}
                        </dd>
                      </div>
                    </dl>

                    <div className="mt-5 flex flex-wrap gap-2">
                      {service.checks?.length > 0 ? (
                        service.checks.map((check) => (
                          <Badge key={check.id}>
                            {check.type}
                            {check.lastResult?.status ? ` • ${check.lastResult.status}` : " • pending"}
                          </Badge>
                        ))
                      ) : (
                        <Badge>No active checks</Badge>
                      )}
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
