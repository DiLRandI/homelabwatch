import { formatDate, formatLatency } from "../../lib/format";
import HealthStatusBadge from "../health/HealthStatusBadge";
import Badge from "../ui/Badge";
import EmptyState from "../ui/EmptyState";

const GROUPS = ["unhealthy", "degraded", "unknown", "healthy"];

export default function PublicStatusPageScreen({ embedded = false, error = "", loading = false, missing = false, page }) {
  if (loading) {
    return <EmptyState body="Loading current service health." title="Loading status" />;
  }
  if (missing || error || !page) {
    return <EmptyState body="This status page is missing or disabled." title="Status page unavailable" />;
  }

  const grouped = GROUPS.map((status) => ({
    status,
    services: (page.services || []).filter((service) => service.status === status),
  })).filter((group) => group.services.length > 0);

  return (
    <main className={embedded ? "rounded-2xl border border-line bg-base p-4" : "min-h-screen bg-base px-4 py-8 text-ink"}>
      <div className="mx-auto grid max-w-5xl gap-6">
        <header className="grid gap-3">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h1 className="text-3xl font-semibold text-ink">{page.title}</h1>
              {page.description ? <p className="mt-2 max-w-3xl text-muted">{page.description}</p> : null}
            </div>
            <HealthStatusBadge status={page.overallStatus} />
          </div>
          <p className="text-sm text-muted">Last updated {formatDate(page.lastUpdatedAt)}</p>
        </header>

        {page.announcements?.length ? (
          <section className="grid gap-3">
            {page.announcements.map((item, index) => (
              <article className="rounded-2xl border border-line bg-panel-strong p-4" key={`${item.title}-${index}`}>
                <Badge>{item.kind}</Badge>
                <h2 className="mt-2 font-semibold text-ink">{item.title}</h2>
                <p className="mt-1 text-sm text-muted">{item.message}</p>
              </article>
            ))}
          </section>
        ) : null}

        {(page.services || []).length === 0 ? (
          <EmptyState body="No services are published on this page yet." title="No published services" />
        ) : (
          <section className="grid gap-4">
            {grouped.map((group) => (
              <div className="grid gap-2" key={group.status}>
                <h2 className="text-sm font-semibold uppercase tracking-[0.18em] text-muted">{group.status}</h2>
                {group.services.map((service) => (
                  <div className="grid gap-2 rounded-2xl border border-line bg-panel-strong p-4 md:grid-cols-[minmax(0,1fr)_auto]" key={service.name}>
                    <div>
                      <div className="font-medium text-ink">{service.name}</div>
                      <div className="text-sm text-muted">Checked {formatDate(service.lastCheckedAt || service.latestCheck?.checkedAt)}</div>
                      {service.latestCheck?.message ? <div className="mt-1 text-sm text-muted">{service.latestCheck.message}</div> : null}
                    </div>
                    <div className="flex flex-wrap items-center gap-2 md:justify-end">
                      <HealthStatusBadge result={service.latestCheck} status={service.status} />
                      {service.latestCheck?.latencyMs ? <Badge>{formatLatency(service.latestCheck.latencyMs)}</Badge> : null}
                    </div>
                  </div>
                ))}
              </div>
            ))}
          </section>
        )}
      </div>
    </main>
  );
}
