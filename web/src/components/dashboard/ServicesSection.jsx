import { useMemo, useState } from "react";

import { formatDate } from "../../lib/format";
import EndpointTester from "../health/EndpointTester";
import HealthSettingsModal from "../health/HealthSettingsModal";
import HealthStatusBadge from "../health/HealthStatusBadge";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import {
  ActivityIcon,
  ArrowUpRightIcon,
  DiscoveryIcon,
  EditIcon,
  PlusIcon,
  ServicesIcon,
} from "../ui/Icons";
import StatusBadge from "../ui/StatusBadge";

const sourceMeta = {
  docker: {
    icon: DiscoveryIcon,
    iconTone: "bg-sky-500/12 text-sky-300",
    tone: "info",
  },
  lan: {
    icon: ActivityIcon,
    iconTone: "bg-base text-ink-soft",
    tone: "neutral",
  },
  manual: {
    icon: ServicesIcon,
    iconTone: "bg-accent/10 text-accent-strong",
    tone: "accent",
  },
};

export default function ServicesSection({
  addLabel = "Add service",
  bookmarkedServiceIds = new Set(),
  canManage = true,
  description = "Tracked endpoints from Docker, manual entry, and network discovery.",
  emptyBody = "Run discovery or create a manual service to start monitoring the control plane.",
  emptyTitle = "No services connected yet",
  onAdd,
  onAddBookmark,
  onDeleteHealthCheck,
  onFetchHealthChecks,
  onSaveHealthCheck,
  onTestHealthCheck,
  sectionId = "services",
  services,
  title = "Services",
}) {
  const [selectedServiceId, setSelectedServiceId] = useState("");
  const selectedService = useMemo(
    () => services.find((item) => item.id === selectedServiceId) || null,
    [selectedServiceId, services],
  );

  return (
    <>
      <section id={sectionId}>
        <Card>
          <CardHeader
            action={
              <Button disabled={!canManage} leadingIcon={PlusIcon} onClick={onAdd}>
                {addLabel}
              </Button>
            }
            description={description}
            title={title}
          />
          <CardContent>
            {services.length === 0 ? (
              <EmptyState
                action={canManage ? onAdd : undefined}
                actionLabel={addLabel}
                body={emptyBody}
                title={emptyTitle}
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
                          <div className="flex flex-wrap items-center gap-3">
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
                          onClick={() =>
                            window.open(service.url, "_blank", "noopener,noreferrer")
                          }
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
                        {service.healthConfigMode ? (
                          <Badge tone={service.healthConfigMode === "custom" ? "accent" : "neutral"}>
                            {service.healthConfigMode}
                          </Badge>
                        ) : null}
                        {service.serviceDefinitionId ? (
                          <Badge tone="info">{service.serviceDefinitionId}</Badge>
                        ) : null}
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
                        {onAddBookmark ? (
                          <Button
                            disabled={!canManage || bookmarkedServiceIds.has(service.id)}
                            onClick={() => onAddBookmark(service)}
                            size="sm"
                            variant={
                              bookmarkedServiceIds.has(service.id) ? "subtle" : "ghost"
                            }
                          >
                            {bookmarkedServiceIds.has(service.id)
                              ? "Bookmarked"
                              : "Add bookmark"}
                          </Button>
                        ) : null}
                        <Button
                          disabled={!canManage}
                          leadingIcon={EditIcon}
                          onClick={() => setSelectedServiceId(service.id)}
                          size="sm"
                          variant="ghost"
                        >
                          Edit health
                        </Button>
                      </div>

                      <div className="mt-5 grid gap-2">
                        {service.checks?.length > 0 ? (
                          service.checks.map((check) => (
                            <div
                              className="rounded-2xl border border-white bg-white px-4 py-3"
                              key={check.id}
                            >
                              <div className="flex flex-wrap items-start justify-between gap-3">
                                <div>
                                  <p className="text-sm font-medium text-slate-900">
                                    {check.name || `${check.type} check`}
                                  </p>
                                  <p className="mt-1 text-xs uppercase tracking-[0.16em] text-slate-500">
                                    {check.type}
                                  </p>
                                </div>
                                <HealthStatusBadge
                                  result={check.lastResult}
                                  status={check.lastResult?.status || "unknown"}
                                  subtle
                                />
                              </div>
                            </div>
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

      <HealthSettingsModal
        canManage={canManage}
        onClose={() => setSelectedServiceId("")}
        onDeleteCheck={onDeleteHealthCheck}
        onFetchChecks={onFetchHealthChecks}
        onSaveCheck={onSaveHealthCheck}
        onTestCheck={onTestHealthCheck}
        open={Boolean(selectedService)}
        service={selectedService}
      />
    </>
  );
}
