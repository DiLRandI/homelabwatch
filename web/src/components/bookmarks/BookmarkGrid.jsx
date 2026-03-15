import EmptyState from "../ui/EmptyState";
import BookmarkCard from "./BookmarkCard";

export default function BookmarkGrid({
  bookmarks = [],
  canManage = true,
  onDelete,
  onEdit,
  onOpen,
  onReorder,
  onToggleFavorite,
}) {
  if (bookmarks.length === 0) {
    return (
      <EmptyState
        body="Bookmarks will appear here once you add a service, save a manual link, or convert a discovered endpoint into a bookmark."
        title="No bookmarks match the current filters"
      />
    );
  }

  return (
    <div className="grid gap-4 xl:grid-cols-2 2xl:grid-cols-3">
      {bookmarks.map((bookmark) => (
        <div
          draggable={canManage}
          key={bookmark.id}
          onDragOver={(event) => event.preventDefault()}
          onDragStart={(event) =>
            event.dataTransfer.setData("text/plain", `bookmark:${bookmark.id}`)
          }
          onDrop={(event) => {
            event.preventDefault();
            const payload = event.dataTransfer.getData("text/plain");
            if (payload.startsWith("bookmark:")) {
              onReorder(payload.slice("bookmark:".length), bookmark.id);
            }
          }}
        >
          <BookmarkCard
            bookmark={bookmark}
            canManage={canManage}
            onDelete={onDelete}
            onEdit={onEdit}
            onOpen={onOpen}
            onToggleFavorite={onToggleFavorite}
          />
        </div>
      ))}
    </div>
  );
}
