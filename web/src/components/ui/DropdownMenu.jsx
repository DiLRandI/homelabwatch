import { useEffect, useRef, useState } from "react";

import { cn } from "../../lib/cn";
import Button from "./Button";
import { ChevronDownIcon } from "./Icons";

export default function DropdownMenu({
  align = "right",
  items,
  label,
  leadingIcon,
  variant = "secondary",
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);

  useEffect(() => {
    if (!open) {
      return undefined;
    }

    function handlePointerDown(event) {
      if (!ref.current?.contains(event.target)) {
        setOpen(false);
      }
    }

    function handleKeyDown(event) {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  return (
    <div className="relative" ref={ref}>
      <Button
        aria-expanded={open}
        aria-haspopup="menu"
        leadingIcon={leadingIcon}
        onClick={() => setOpen((value) => !value)}
        trailingIcon={ChevronDownIcon}
        variant={variant}
      >
        {label}
      </Button>
      {open ? (
        <div
          className={cn(
            "absolute top-[calc(100%+0.75rem)] z-30 w-72 rounded-2xl border border-slate-200 bg-white p-2 shadow-card-lg",
            align === "left" ? "left-0" : "right-0",
          )}
          role="menu"
        >
          {items.map((item) => (
            <button
              className="flex w-full items-start gap-3 rounded-2xl px-3 py-3 text-left transition hover:bg-slate-50 focus-visible:outline-hidden focus-visible:ring-4 focus-visible:ring-slate-200"
              key={item.label}
              onClick={() => {
                setOpen(false);
                item.onSelect();
              }}
              role="menuitem"
              type="button"
            >
              {item.icon ? (
                <span className="mt-0.5 inline-flex h-9 w-9 items-center justify-center rounded-xl bg-slate-100 text-slate-600">
                  <item.icon className="h-4 w-4" />
                </span>
              ) : null}
              <span className="min-w-0">
                <span className="block text-sm font-medium text-slate-900">
                  {item.label}
                </span>
                {item.description ? (
                  <span className="mt-1 block text-sm leading-5 text-slate-500">
                    {item.description}
                  </span>
                ) : null}
              </span>
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
