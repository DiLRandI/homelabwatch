import ManualServiceForm from "../forms/ManualServiceForm";
import EmptyState from "../ui/EmptyState";
import Section from "../ui/Section";
import StatusBadge from "../ui/StatusBadge";

export default function ServicesSection({ services, onSubmit }) {
  return (
    <Section
      title="Services"
      subtitle="Manual links, Docker workloads, and LAN discoveries in one table."
    >
      <div className="grid gap-3">
        {services.length === 0 ? (
          <EmptyState
            title="No services yet"
            body="Run discovery or add a manual endpoint."
            compact
          />
        ) : (
          services.map((service) => (
            <article
              key={service.id}
              className="rounded-3xl border border-white/10 bg-base/70 p-4"
            >
              <div className="flex items-start justify-between gap-4">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="font-display text-lg font-semibold text-ink">
                      {service.name}
                    </h3>
                    <StatusBadge status={service.status} />
                    <span className="rounded-full border border-white/10 px-3 py-1 text-[11px] uppercase tracking-[0.22em] text-muted">
                      {service.source}
                    </span>
                  </div>
                  <p className="mt-2 text-sm text-muted">{service.url}</p>
                  <p className="mt-1 text-xs uppercase tracking-[0.2em] text-muted/80">
                    {service.deviceName || service.host}:{service.port}
                  </p>
                </div>
                <a
                  className="rounded-full border border-accent/50 px-4 py-2 text-sm text-accent transition hover:bg-accent hover:text-base"
                  href={service.url}
                  rel="noreferrer"
                  target="_blank"
                >
                  Open
                </a>
              </div>
              {service.checks?.length > 0 ? (
                <div className="mt-4 flex flex-wrap gap-2 text-xs text-muted">
                  {service.checks.map((check) => (
                    <span
                      key={check.id}
                      className="rounded-full border border-white/10 px-3 py-1"
                    >
                      {check.type}{" "}
                      {check.lastResult?.status
                        ? `• ${check.lastResult.status}`
                        : "• pending"}
                    </span>
                  ))}
                </div>
              ) : null}
            </article>
          ))
        )}
      </div>
      <ManualServiceForm onSubmit={onSubmit} />
    </Section>
  );
}
