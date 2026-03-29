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
      className={`rounded-3xl border border-dashed border-line bg-base ${compact ? "p-5" : "p-8"}`}
    >
      <h3 className="font-display text-lg font-semibold text-ink">{title}</h3>
      <p className="mt-2 max-w-xl text-sm leading-6 text-muted">{body}</p>
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
