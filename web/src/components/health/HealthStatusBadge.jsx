import { formatBytes, formatDate, formatLatency } from "../../lib/format";
import { cn } from "../../lib/cn";
import Badge from "../ui/Badge";
import StatusBadge from "../ui/StatusBadge";

export default function HealthStatusBadge({
  className,
  result,
  showCheckedAt = false,
  status,
  subtle = false,
}) {
  const resolvedStatus = result?.status || status || "unknown";
  const meta = [];

  if (result?.httpStatusCode) {
    meta.push(`HTTP ${result.httpStatusCode}`);
  }
  if (result?.latencyMs || result?.latencyMS) {
    meta.push(formatLatency(result.latencyMs ?? result.latencyMS));
  }
  if (result?.responseSizeBytes) {
    meta.push(formatBytes(result.responseSizeBytes));
  }
  if (showCheckedAt && result?.checkedAt) {
    meta.push(formatDate(result.checkedAt));
  }

  return (
    <div className={cn("flex flex-wrap items-center gap-2", className)}>
      <StatusBadge status={resolvedStatus} subtle={subtle} />
      {meta.map((item) => (
        <Badge key={item}>{item}</Badge>
      ))}
    </div>
  );
}
