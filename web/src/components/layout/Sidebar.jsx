import { cn } from "../../lib/cn";
import Badge from "../ui/Badge";
import { CloseIcon } from "../ui/Icons";

export default function Sidebar({
  activeHref,
  navItems,
  onClose,
  onNavigate,
  open = false,
  sidebarMeta,
}) {
  return (
    <>
      <div
        className={cn(
          "fixed inset-0 z-30 bg-overlay backdrop-blur-sm transition lg:hidden",
          open ? "opacity-100" : "pointer-events-none opacity-0",
        )}
        onClick={onClose}
      />
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 flex w-[280px] max-w-[85vw] flex-col overflow-y-auto border-r border-line bg-panel-strong p-5 shadow-card-lg transition lg:sticky lg:top-0 lg:z-auto lg:h-screen lg:translate-x-0 lg:overscroll-contain lg:shadow-none",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-accent-strong">
              HomelabWatch
            </p>
            <h1 className="mt-2 text-xl font-semibold tracking-tight text-ink">
              {sidebarMeta?.applianceName || "Control plane"}
            </h1>
            <p className="mt-2 text-sm leading-6 text-muted">
              One workspace for services, infrastructure, and runtime health.
            </p>
          </div>
          <button
            aria-label="Close navigation"
            className="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-line text-muted transition hover:bg-base hover:text-ink lg:hidden"
            onClick={onClose}
            type="button"
          >
            <CloseIcon className="h-4 w-4" />
          </button>
        </div>

        <nav aria-label="Dashboard sections" className="mt-8 flex-1 space-y-1">
          {navItems.map((item) => (
            <a
              aria-current={item.href === activeHref ? "page" : undefined}
              className={cn(
                "group flex items-center justify-between rounded-2xl px-3 py-3 text-sm font-medium transition",
                item.href === activeHref
                  ? "bg-base text-ink"
                  : "text-muted hover:bg-base hover:text-ink",
              )}
              href={item.href}
              key={item.href}
              onClick={(event) => {
                event.preventDefault();
                onNavigate?.(item.href);
                onClose();
              }}
            >
              <span className="flex items-center gap-3">
                <span
                  className={cn(
                    "inline-flex h-9 w-9 items-center justify-center rounded-xl transition",
                    item.href === activeHref
                      ? "bg-panel-strong text-accent-strong"
                      : "bg-base text-muted group-hover:bg-panel-strong group-hover:text-accent-strong",
                  )}
                >
                  <item.icon className="h-4 w-4" />
                </span>
                {item.label}
              </span>
              {item.count != null ? <Badge>{item.count}</Badge> : null}
            </a>
          ))}
        </nav>

      </aside>
    </>
  );
}
