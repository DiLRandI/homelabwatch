import { cn } from "../../lib/cn";
import { MenuIcon } from "../ui/Icons";

export default function Navbar({
  onOpenSidebar,
  statusItems = [],
  subtitle,
  title,
  toolbar,
}) {
  return (
    <header className="sticky top-0 z-20 border-b border-line bg-page/90 backdrop-blur-xl">
      <div className="px-4 py-4 pr-20 sm:px-6 sm:pr-24 lg:px-8">
        <div className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
          <div className="min-w-0">
            <div className="flex items-center gap-3">
              <button
                aria-label="Open navigation"
                className="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-line bg-panel-strong text-muted shadow-sm transition hover:bg-base hover:text-ink lg:hidden"
                onClick={onOpenSidebar}
                type="button"
              >
                <MenuIcon className="h-4 w-4" />
              </button>
              <div className="min-w-0">
                <h2 className="truncate text-2xl font-semibold tracking-tight text-ink">
                  {title}
                </h2>
                <p className="mt-1 text-sm leading-6 text-muted">{subtitle}</p>
              </div>
            </div>
            {statusItems.length > 0 ? (
              <div className="mt-4 flex flex-wrap items-center gap-2">
                {statusItems.map((item) => (
                  <span
                    className={cn(
                      "inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-medium",
                      item.className,
                    )}
                    key={item.label}
                  >
                    {item.icon ? <item.icon className="h-3.5 w-3.5" /> : null}
                    {item.label}
                  </span>
                ))}
              </div>
            ) : null}
          </div>
          {toolbar ? <div className="xl:min-w-[520px]">{toolbar}</div> : null}
        </div>
      </div>
    </header>
  );
}
