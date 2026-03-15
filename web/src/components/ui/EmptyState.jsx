export default function EmptyState({ title, body, compact = false }) {
  return (
    <div
      className={`rounded-3xl border border-dashed border-white/15 bg-base/50 ${compact ? "p-4" : "p-8"}`}
    >
      <h3 className="font-display text-lg font-semibold text-ink">{title}</h3>
      <p className="mt-2 text-sm text-muted">{body}</p>
    </div>
  );
}
