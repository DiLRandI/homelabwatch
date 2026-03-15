import Badge from "./Badge";

const TONES = {
  healthy: "success",
  degraded: "warning",
  unhealthy: "danger",
  unknown: "neutral",
};

export default function StatusBadge({ status, subtle = false }) {
  return (
    <Badge tone={TONES[status] || TONES.unknown} withDot={!subtle}>
      {status || "unknown"}
    </Badge>
  );
}
