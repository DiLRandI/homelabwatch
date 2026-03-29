import { useEffect } from "react";
import { createPortal } from "react-dom";

import { cn } from "../../lib/cn";
import { CloseIcon } from "./Icons";

export default function Modal({
  children,
  className,
  description,
  onClose,
  open,
  title,
}) {
  useEffect(() => {
    if (!open) {
      return undefined;
    }

    function handleKeyDown(event) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [onClose, open]);

  if (!open) {
    return null;
  }

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-overlay p-4 backdrop-blur-sm sm:items-center">
      <div
        aria-hidden="true"
        className="absolute inset-0"
        onClick={onClose}
      />
      <div
        aria-labelledby="modal-title"
        aria-modal="true"
        className={cn(
          "relative z-10 w-full max-w-2xl overflow-hidden rounded-[28px] border border-line bg-panel-strong shadow-card-lg",
          className,
        )}
        role="dialog"
      >
        <div className="flex items-start justify-between gap-4 border-b border-line px-5 py-5 sm:px-6">
          <div>
            <h2
              className="text-lg font-semibold tracking-tight text-ink"
              id="modal-title"
            >
              {title}
            </h2>
            {description ? (
              <p className="mt-1 text-sm leading-6 text-muted">
                {description}
              </p>
            ) : null}
          </div>
          <button
            aria-label="Close dialog"
            className="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-line text-muted transition hover:bg-base hover:text-ink focus-visible:outline-hidden focus-visible:ring-4 focus-visible:ring-line"
            onClick={onClose}
            type="button"
          >
            <CloseIcon className="h-4 w-4" />
          </button>
        </div>
        <div className="px-5 py-5 sm:px-6">{children}</div>
      </div>
    </div>,
    document.body,
  );
}
