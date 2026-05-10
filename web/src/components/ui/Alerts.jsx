import { useEffect, useState } from "react";

import { cn } from "../../lib/cn";
import { CloseIcon } from "./Icons";

export default function Alerts({ error, notice }) {
  const [dismissedNotice, setDismissedNotice] = useState(null);

  useEffect(() => {
    setDismissedNotice(null);
  }, [notice]);

  const visibleNotice = notice && dismissedNotice !== notice ? notice : null;

  if (!error && !visibleNotice) {
    return null;
  }

  return (
    <div className="grid gap-3">
      {visibleNotice ? (
        <div
          className={cn(
            "flex items-center justify-between gap-4 rounded-2xl border px-4 py-3 text-sm font-medium shadow-sm",
            "border-ok/15 bg-ok/10 text-ok-strong",
          )}
        >
          <span>{visibleNotice}</span>
          <button
            aria-label="Dismiss notice"
            className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-xl border border-ok/15 text-ok-strong transition hover:bg-ok/10"
            onClick={() => setDismissedNotice(visibleNotice)}
            type="button"
          >
            <CloseIcon className="h-4 w-4" />
          </button>
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
