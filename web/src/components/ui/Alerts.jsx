import { cn } from "../../lib/cn";

export default function Alerts({ error, notice }) {
  if (!error && !notice) {
    return null;
  }

  return (
    <div className="grid gap-3">
      {notice ? (
        <div
          className={cn(
            "rounded-2xl border px-4 py-3 text-sm font-medium",
            "border-ok/15 bg-ok/10 text-ok-strong",
          )}
        >
          {notice}
        </div>
      ) : null}
      {error ? (
        <div
          className={cn(
            "rounded-2xl border px-4 py-3 text-sm font-medium",
            "border-danger/15 bg-danger/10 text-danger-strong",
          )}
        >
          {error}
        </div>
      ) : null}
    </div>
  );
}
