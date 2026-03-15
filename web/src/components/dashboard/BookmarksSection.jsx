import BookmarkForm from "../forms/BookmarkForm";
import EmptyState from "../ui/EmptyState";
import Section from "../ui/Section";

export default function BookmarksSection({ bookmarks, onSubmit }) {
  return (
    <Section
      title="Bookmarks"
      subtitle="User-curated links that live beside auto-discovered services."
    >
      <div className="grid gap-3">
        {bookmarks.length === 0 ? (
          <EmptyState
            title="No bookmarks yet"
            body="Add external dashboards or docs here."
            compact
          />
        ) : (
          bookmarks.map((bookmark) => (
            <a
              key={bookmark.id}
              className="rounded-3xl border border-white/10 bg-base/70 p-4 transition hover:border-accent/50"
              href={bookmark.url}
              rel="noreferrer"
              target="_blank"
            >
              <div className="flex items-center justify-between gap-4">
                <div>
                  <h3 className="font-display text-lg font-semibold text-ink">
                    {bookmark.name}
                  </h3>
                  <p className="mt-1 text-sm text-muted">{bookmark.url}</p>
                </div>
                <span className="text-xs uppercase tracking-[0.2em] text-accent">
                  Open
                </span>
              </div>
            </a>
          ))
        )}
      </div>
      <BookmarkForm onSubmit={onSubmit} />
    </Section>
  );
}
