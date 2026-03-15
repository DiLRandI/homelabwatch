import { formatBytes, formatDate, formatLatency } from "../../lib/format";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import { DiscoveryIcon, ShieldIcon, SparklesIcon } from "../ui/Icons";
import HealthStatusBadge from "./HealthStatusBadge";

function Metric({ label, value }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-white px-4 py-3">
      <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500">
        {label}
      </p>
      <p className="mt-2 break-words text-sm font-medium text-slate-900">{value}</p>
    </div>
  );
}

export default function EndpointTester({ loading = false, result }) {
  return (
    <Card className="border-slate-200 bg-slate-50">
      <CardHeader
        description="Run an on-demand probe before saving changes. Blank HTTP paths can trigger smart endpoint discovery."
        title="Endpoint tester"
      />
      <CardContent>
        {!result ? (
          <EmptyState
            body="Test the selected check to inspect latency, payload size, and the resolved endpoint before it is saved."
            title={loading ? "Testing endpoint" : "No test run yet"}
          />
        ) : (
          <div className="grid gap-4">
            <div className="flex flex-wrap items-start justify-between gap-3 rounded-2xl border border-white bg-white px-4 py-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                  Last probe
                </p>
                <HealthStatusBadge
                  className="mt-3"
                  result={{
                    checkedAt: result.checkedAt,
                    httpStatusCode: result.httpStatusCode,
                    latencyMs: result.latencyMs,
                    responseSizeBytes: result.responseSizeBytes,
                    status: result.status,
                  }}
                  showCheckedAt
                />
              </div>
              {result.message ? (
                <p className="max-w-xl text-sm leading-6 text-slate-600">{result.message}</p>
              ) : null}
            </div>

            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <Metric
                label="Resolved URL"
                value={result.resolvedUrl || result.check?.target || "Not resolved"}
              />
              <Metric
                label="Latency"
                value={formatLatency(result.latencyMs)}
              />
              <Metric
                label="Response size"
                value={formatBytes(result.responseSizeBytes)}
              />
              <Metric
                label="Matched definition"
                value={result.matchedServiceDefinition?.name || "None"}
              />
            </div>

            <div className="grid gap-3 lg:grid-cols-3">
              <div className="rounded-2xl border border-white bg-white px-4 py-4">
                <div className="flex items-center gap-3">
                  <span className="inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-ok/10 text-ok-strong">
                    <ShieldIcon className="h-4 w-4" />
                  </span>
                  <div>
                    <p className="text-sm font-medium text-slate-900">Status code</p>
                    <p className="text-sm text-slate-500">
                      {result.httpStatusCode || "n/a"}
                    </p>
                  </div>
                </div>
              </div>
              <div className="rounded-2xl border border-white bg-white px-4 py-4">
                <div className="flex items-center gap-3">
                  <span className="inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-sky-50 text-sky-700">
                    <DiscoveryIcon className="h-4 w-4" />
                  </span>
                  <div>
                    <p className="text-sm font-medium text-slate-900">Check type</p>
                    <p className="text-sm text-slate-500">
                      {result.check?.type || "Unknown"}
                    </p>
                  </div>
                </div>
              </div>
              <div className="rounded-2xl border border-white bg-white px-4 py-4">
                <div className="flex items-center gap-3">
                  <span className="inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-amber-100 text-amber-700">
                    <SparklesIcon className="h-4 w-4" />
                  </span>
                  <div>
                    <p className="text-sm font-medium text-slate-900">Observed at</p>
                    <p className="text-sm text-slate-500">{formatDate(result.checkedAt)}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
