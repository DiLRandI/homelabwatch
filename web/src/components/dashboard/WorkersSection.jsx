import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import StatusBadge from "../ui/StatusBadge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/Table";

export default function WorkersSection({ jobState, recentEvents }) {
  return (
    <section className="grid gap-6 xl:grid-cols-[minmax(0,0.92fr)_minmax(0,1.08fr)]" id="activity">
      <Card>
        <CardHeader
          description="Scheduler runs, cadence, and failure context for long-running background jobs."
          title="Workers"
        />
        <CardContent className="p-0">
          {jobState.length === 0 ? (
            <div className="px-5 py-5 sm:px-6">
              <EmptyState
                body="Background jobs will report here after bootstrap completes and the scheduler starts running."
                title="No worker runs yet"
              />
            </div>
          ) : (
            <Table>
              <TableHead>
                <tr>
                  <TableHeader>Job</TableHeader>
                  <TableHeader>Status</TableHeader>
                  <TableHeader>Last run</TableHeader>
                </tr>
              </TableHead>
              <TableBody>
                {jobState.map((job) => (
                  <TableRow key={job.jobName}>
                    <TableCell className="min-w-[220px]">
                      <p className="font-medium text-slate-900">{job.jobName}</p>
                      {job.lastError ? (
                        <p className="mt-1 text-sm text-danger-strong">{job.lastError}</p>
                      ) : (
                        <p className="mt-1 text-sm text-slate-500">
                          Completed without reported errors
                        </p>
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge tone={job.lastError ? "danger" : "success"}>
                        {job.lastError ? "error" : "ok"}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatDate(job.lastRunAt)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader
          description="Recent service and infrastructure events streamed from the live control plane."
          title="Recent events"
        />
        <CardContent>
          {recentEvents.length === 0 ? (
            <EmptyState
              body="Health changes, discoveries, and state transitions will appear here as the system starts observing your estate."
              title="No recent events yet"
            />
          ) : (
            <div className="grid gap-3">
              {recentEvents.map((item) => (
                <article
                  className="rounded-3xl border border-slate-200 bg-slate-50 p-4"
                  key={item.id}
                >
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p className="font-medium text-slate-900">{item.eventType}</p>
                      <p className="mt-1 text-sm leading-6 text-slate-500">
                        {item.message}
                      </p>
                    </div>
                    <StatusBadge status={item.status} subtle />
                  </div>
                  <p className="mt-4 text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">
                    {formatDate(item.createdAt)}
                  </p>
                </article>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
