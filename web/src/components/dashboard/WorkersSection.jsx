import { formatDate } from "../../lib/format";
import CardList from "../ui/CardList";
import EmptyState from "../ui/EmptyState";
import Section from "../ui/Section";
import StatusBadge from "../ui/StatusBadge";

export default function WorkersSection({ jobState, recentEvents }) {
  return (
    <Section
      title="Workers"
      subtitle="Recent scheduler outcomes and health events."
    >
      <div className="grid gap-3">
        {jobState.length === 0 ? (
          <EmptyState
            title="No worker runs yet"
            body="Background jobs will report here after bootstrap."
            compact
          />
        ) : (
          jobState.map((job) => (
            <div
              key={job.jobName}
              className="rounded-3xl border border-white/10 bg-base/70 p-4"
            >
              <div className="flex items-center justify-between gap-4">
                <h3 className="font-semibold text-ink">{job.jobName}</h3>
                <span
                  className={`text-xs uppercase tracking-[0.2em] ${job.lastError ? "text-danger" : "text-ok"}`}
                >
                  {job.lastError ? "error" : "ok"}
                </span>
              </div>
              <p className="mt-2 text-sm text-muted">
                Last run: {formatDate(job.lastRunAt)}
              </p>
              {job.lastError ? (
                <p className="mt-1 text-sm text-danger">{job.lastError}</p>
              ) : null}
            </div>
          ))
        )}

        <CardList
          items={recentEvents}
          renderItem={(item) => (
            <div
              key={item.id}
              className="rounded-3xl border border-white/10 bg-base/70 p-4"
            >
              <div className="flex items-center justify-between gap-4">
                <span className="font-semibold text-ink">{item.eventType}</span>
                <StatusBadge status={item.status} subtle />
              </div>
              <p className="mt-2 text-sm text-muted">{item.message}</p>
              <p className="mt-2 text-xs uppercase tracking-[0.2em] text-muted/80">
                {formatDate(item.createdAt)}
              </p>
            </div>
          )}
          title="Recent events"
        />
      </div>
    </Section>
  );
}
