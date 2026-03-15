import Button from "../ui/Button";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  EditIcon,
  FolderIcon,
  PlusIcon,
  TrashIcon,
} from "../ui/Icons";

function FolderNode({
  canManage,
  depth,
  folder,
  onDelete,
  onDropBookmark,
  onDropFolder,
  onEdit,
  onMoveDown,
  onMoveUp,
  onSelect,
  selectedFolderId,
}) {
  return (
    <div className="space-y-2">
      <div
        className={`flex items-center justify-between gap-3 rounded-2xl border px-3 py-2 transition ${
          selectedFolderId === folder.id
            ? "border-accent bg-accent/10"
            : "border-slate-200 bg-white hover:border-slate-300"
        }`}
        draggable={canManage}
        onDragOver={(event) => event.preventDefault()}
        onDragStart={(event) =>
          event.dataTransfer.setData("text/plain", `folder:${folder.id}`)
        }
        onDrop={(event) => {
          if (!canManage) {
            return;
          }
          event.preventDefault();
          const payload = event.dataTransfer.getData("text/plain");
          if (payload.startsWith("bookmark:")) {
            onDropBookmark(payload.slice("bookmark:".length), folder.id);
          }
          if (payload.startsWith("folder:")) {
            onDropFolder(payload.slice("folder:".length), folder.id);
          }
        }}
        style={{ marginLeft: `${depth * 12}px` }}
      >
        <button
          className="min-w-0 flex-1 text-left"
          onClick={() => onSelect(folder.id)}
          type="button"
        >
          <span className="flex items-center gap-2 text-sm font-medium text-slate-900">
            <FolderIcon className="h-4 w-4 text-accent-strong" />
            <span className="truncate">{folder.name}</span>
          </span>
          <span className="mt-1 block text-xs text-slate-500">
            {folder.bookmarkCount} bookmark{folder.bookmarkCount === 1 ? "" : "s"}
          </span>
        </button>
        <div className="flex items-center gap-1">
          <button
            className="rounded-xl p-2 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 disabled:opacity-50"
            disabled={!canManage}
            onClick={() => onMoveUp(folder)}
            type="button"
          >
            <ArrowUpIcon className="h-4 w-4" />
          </button>
          <button
            className="rounded-xl p-2 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 disabled:opacity-50"
            disabled={!canManage}
            onClick={() => onMoveDown(folder)}
            type="button"
          >
            <ArrowDownIcon className="h-4 w-4" />
          </button>
          <button
            className="rounded-xl p-2 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 disabled:opacity-50"
            disabled={!canManage}
            onClick={() => onEdit(folder)}
            type="button"
          >
            <EditIcon className="h-4 w-4" />
          </button>
          <button
            className="rounded-xl p-2 text-slate-400 transition hover:bg-rose-50 hover:text-rose-600 disabled:opacity-50"
            disabled={!canManage}
            onClick={() => onDelete(folder)}
            type="button"
          >
            <TrashIcon className="h-4 w-4" />
          </button>
        </div>
      </div>
      {folder.children?.map((child) => (
        <FolderNode
          canManage={canManage}
          depth={depth + 1}
          folder={child}
          key={child.id}
          onDelete={onDelete}
          onDropBookmark={onDropBookmark}
          onDropFolder={onDropFolder}
          onEdit={onEdit}
          onMoveDown={onMoveDown}
          onMoveUp={onMoveUp}
          onSelect={onSelect}
          selectedFolderId={selectedFolderId}
        />
      ))}
    </div>
  );
}

export default function FolderTree({
  canManage = true,
  folders = [],
  onAdd,
  onDelete,
  onDropBookmark,
  onDropFolder,
  onEdit,
  onMoveDown,
  onMoveUp,
  onSelect,
  selectedFolderId = "",
}) {
  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">
            Folders
          </p>
          <h3 className="mt-1 text-lg font-semibold tracking-tight text-slate-950">
            Organize navigation
          </h3>
        </div>
        <Button
          disabled={!canManage}
          leadingIcon={PlusIcon}
          onClick={onAdd}
          size="sm"
          variant="secondary"
        >
          New
        </Button>
      </div>

      <button
        className={`flex w-full items-center justify-between rounded-2xl border px-4 py-3 text-left transition ${
          selectedFolderId
            ? "border-slate-200 bg-white hover:border-slate-300"
            : "border-accent bg-accent/10"
        }`}
        onDragOver={(event) => event.preventDefault()}
        onDrop={(event) => {
          if (!canManage) {
            return;
          }
          event.preventDefault();
          const payload = event.dataTransfer.getData("text/plain");
          if (payload.startsWith("bookmark:")) {
            onDropBookmark(payload.slice("bookmark:".length), "");
          }
          if (payload.startsWith("folder:")) {
            onDropFolder(payload.slice("folder:".length), "");
          }
        }}
        onClick={() => onSelect("")}
        type="button"
      >
        <span>
          <span className="block text-sm font-medium text-slate-900">All bookmarks</span>
          <span className="mt-1 block text-xs text-slate-500">Across every folder</span>
        </span>
      </button>

      <div className="space-y-2">
        {folders.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
            Create folders for monitoring, media, infrastructure, or any workflow you reach for often.
          </div>
        ) : (
          folders.map((folder) => (
            <FolderNode
              canManage={canManage}
              depth={0}
              folder={folder}
              key={folder.id}
              onDelete={onDelete}
              onDropBookmark={onDropBookmark}
              onDropFolder={onDropFolder}
              onEdit={onEdit}
              onMoveDown={onMoveDown}
              onMoveUp={onMoveUp}
              onSelect={onSelect}
              selectedFolderId={selectedFolderId}
            />
          ))
        )}
      </div>
    </section>
  );
}
