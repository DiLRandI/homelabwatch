import { cn } from "../../lib/cn";
import Badge from "../ui/Badge";
import { ActivityIcon, CloseIcon, ShieldIcon } from "../ui/Icons";

export default function Sidebar({
  metrics,
  navItems,
  onClose,
  open = false,
  sidebarMeta,
}) {
  return (
    <>
      <div
        className={cn(
          "fixed inset-0 z-30 bg-slate-950/35 backdrop-blur-sm transition lg:hidden",
          open ? "opacity-100" : "pointer-events-none opacity-0",
        )}
        onClick={onClose}
      />
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 flex w-[280px] max-w-[85vw] flex-col border-r border-slate-200 bg-white p-5 shadow-card-lg transition lg:sticky lg:top-0 lg:z-auto lg:h-screen lg:translate-x-0 lg:shadow-none",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-accent-strong">
              Homelabwatch
            </p>
            <h1 className="mt-2 text-xl font-semibold tracking-tight text-slate-950">
              {sidebarMeta?.applianceName || "Control plane"}
            </h1>
            <p className="mt-2 text-sm leading-6 text-slate-500">
              One workspace for services, infrastructure, and runtime health.
            </p>
          </div>
          <button
            aria-label="Close navigation"
            className="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 text-slate-500 transition hover:bg-slate-50 hover:text-slate-900 lg:hidden"
            onClick={onClose}
            type="button"
          >
            <CloseIcon className="h-4 w-4" />
          </button>
        </div>

        <div className="mt-8 rounded-3xl border border-slate-200 bg-slate-50 p-4">
          <div className="flex items-center gap-3">
            <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-accent/10 text-accent-strong">
              <ShieldIcon className="h-5 w-5" />
            </span>
            <div>
              <p className="text-sm font-medium text-slate-900">Workspace status</p>
              <p className="text-sm text-slate-500">Live service and network inventory</p>
            </div>
          </div>
          <dl className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
            {metrics.slice(0, 4).map((metric) => (
              <div
                className="rounded-2xl border border-white bg-white px-3 py-3"
                key={metric.label}
              >
                <dt className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                  {metric.label}
                </dt>
                <dd className="mt-2 text-2xl font-semibold tracking-tight text-slate-950">
                  {metric.value}
                </dd>
              </div>
            ))}
          </dl>
        </div>

        <nav aria-label="Dashboard sections" className="mt-8 flex-1 space-y-1">
          {navItems.map((item) => (
            <a
              className="group flex items-center justify-between rounded-2xl px-3 py-3 text-sm font-medium text-slate-600 transition hover:bg-slate-100 hover:text-slate-950"
              href={item.href}
              key={item.href}
              onClick={onClose}
            >
              <span className="flex items-center gap-3">
                <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-slate-100 text-slate-500 transition group-hover:bg-white group-hover:text-accent-strong">
                  <item.icon className="h-4 w-4" />
                </span>
                {item.label}
              </span>
              {item.count != null ? <Badge>{item.count}</Badge> : null}
            </a>
          ))}
        </nav>

        <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4">
          <div className="flex items-center gap-3">
            <span className="inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-white text-slate-500 shadow-sm">
              <ActivityIcon className="h-4 w-4" />
            </span>
            <div>
              <p className="text-sm font-medium text-slate-900">API access</p>
              <p className="text-sm text-slate-500">
                Trusted local UI plus external bearer tokens
              </p>
            </div>
          </div>
          <div className="mt-3 grid gap-2">
            <p className="rounded-2xl border border-slate-200 bg-white px-3 py-2 text-xs font-medium text-slate-500">
              {sidebarMeta?.trustedNetwork
                ? "This browser can perform local write actions"
                : "This browser is outside the trusted write boundary"}
            </p>
            <p className="rounded-2xl border border-slate-200 bg-white px-3 py-2 text-xs font-medium text-slate-500">
              {sidebarMeta?.apiTokenCount ?? 0} external API token
              {(sidebarMeta?.apiTokenCount ?? 0) === 1 ? "" : "s"} configured
            </p>
          </div>
        </div>
      </aside>
    </>
  );
}
