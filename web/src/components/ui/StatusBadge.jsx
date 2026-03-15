const TONES = {
  healthy: "border-ok/40 text-ok",
  degraded: "border-warn/40 text-warn",
  unhealthy: "border-danger/40 text-danger",
  unknown: "border-white/15 text-muted",
};

export default function StatusBadge({ status, subtle = false }) {
  return (
    <span
      className={`rounded-full border px-3 py-1 text-[11px] uppercase tracking-[0.24em] ${subtle ? "bg-transparent" : "bg-white/5"} ${TONES[status] || TONES.unknown}`}
    >
      {status || "unknown"}
    </span>
  );
}
