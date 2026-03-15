import Button from "./Button";

export default function EmptyState({
  action,
  actionLabel,
  body,
  compact = false,
  title,
}) {
  return (
    <div
      className={`rounded-3xl border border-dashed border-slate-200 bg-slate-50 ${compact ? "p-5" : "p-8"}`}
    >
      <h3 className="font-display text-lg font-semibold text-slate-950">{title}</h3>
      <p className="mt-2 max-w-xl text-sm leading-6 text-slate-500">{body}</p>
      {action && actionLabel ? (
        <div className="mt-4">
          <Button onClick={action} variant="secondary">
            {actionLabel}
          </Button>
        </div>
      ) : null}
    </div>
  );
}
