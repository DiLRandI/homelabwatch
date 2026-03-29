import Button from "../ui/Button";
import { ArrowUpRightIcon, PinIcon } from "../ui/Icons";

export default function FavoritesStrip({ bookmarks = [], onOpen }) {
  if (bookmarks.length === 0) {
    return null;
  }

  return (
    <section className="surface-warm rounded-[30px] border border-amber-400/25 p-5 shadow-card">
      <div className="flex items-center gap-3">
        <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-500/12 text-amber-300">
          <PinIcon className="h-5 w-5" />
        </span>
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-amber-300">
            Favorites
          </p>
          <h3 className="mt-1 text-lg font-semibold tracking-tight text-ink">
            Most important services
          </h3>
        </div>
      </div>

      <div className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {bookmarks.map((bookmark) => (
          <button
            className="rounded-2xl border border-line bg-panel-strong px-4 py-4 text-left shadow-sm transition hover:-translate-y-0.5 hover:border-amber-400/35"
            key={bookmark.id}
            onClick={() => onOpen(bookmark)}
            type="button"
          >
            <span className="block text-sm font-semibold text-ink">{bookmark.name}</span>
            <span className="mt-1 block truncate text-sm text-muted">
              {bookmark.deviceName || bookmark.url}
            </span>
            <span className="mt-4 inline-flex items-center gap-2 text-sm font-medium text-accent-strong">
              Open now
              <ArrowUpRightIcon className="h-4 w-4" />
            </span>
          </button>
        ))}
      </div>
    </section>
  );
}
