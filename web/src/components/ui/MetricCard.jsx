export default function MetricCard({ label, value, tone }) {
  return (
    <div className="rounded-3xl border border-white/10 bg-base/70 p-4">
      <div className="text-xs uppercase tracking-[0.26em] text-muted">
        {label}
      </div>
      <div className={`mt-3 font-display text-4xl font-semibold ${tone}`}>
        {value}
      </div>
    </div>
  );
}
