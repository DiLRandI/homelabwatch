import { useEffect, useState } from "react";

import { bookmarkOpenURL } from "../../lib/api";
import { useBookmarksWorkspace } from "../../hooks/useBookmarksWorkspace";
import Modal from "../ui/Modal";
import { FolderIcon } from "../ui/Icons";
import BookmarkEditorForm from "./BookmarkEditorForm";
import BookmarkGrid from "./BookmarkGrid";
import BookmarkToolbar from "./BookmarkToolbar";
import FavoritesStrip from "./FavoritesStrip";
import ImportExportDialog from "./ImportExportDialog";
import FolderEditorForm from "../folders/FolderEditorForm";
import FolderTree from "../folders/FolderTree";
import TagFilterBar from "../tags/TagFilterBar";

function bookmarkToInput(bookmark) {
  return {
    description: bookmark.description || "",
    deviceId: bookmark.deviceId || "",
    folderId: bookmark.folderId || "",
    iconMode: bookmark.iconMode || "auto",
    iconValue: bookmark.iconValue || "",
    id: bookmark.id,
    isFavorite: Boolean(bookmark.isFavorite),
    name: bookmark.manualName || bookmark.name || "",
    serviceId: bookmark.serviceId || "",
    tags: bookmark.tags || [],
    url: bookmark.manualUrl || bookmark.url || "",
    useDevicePrimaryAddress: Boolean(bookmark.useDevicePrimaryAddress),
  };
}

function reorderPayload(bookmarks, ordered, folderId) {
  return ordered.map((bookmark, index) => ({
    id: bookmark.id,
    favoritePosition: bookmark.isFavorite ? bookmark.favoritePosition ?? index : 0,
    folderId: folderId ?? (bookmark.folderId || ""),
    isFavorite: Boolean(bookmark.isFavorite),
    position: index,
  }));
}

function moveItem(items, draggedId, targetId) {
  const list = [...items];
  const draggedIndex = list.findIndex((item) => item.id === draggedId);
  const targetIndex = list.findIndex((item) => item.id === targetId);
  if (draggedIndex === -1 || targetIndex === -1) {
    return list;
  }
  const [dragged] = list.splice(draggedIndex, 1);
  list.splice(targetIndex, 0, dragged);
  return list;
}

export default function BookmarksHome({
  bookmarks = [],
  canManage = true,
  devices = [],
  folders = [],
  onDeleteBookmark,
  onDeleteFolder,
  onExportBookmarks,
  onImportBookmarks,
  onReorderBookmarks,
  onReorderFolders,
  onSaveBookmark,
  onSaveFolder,
  onUploadBookmarkIcon,
  openBookmarkComposerToken = 0,
  services = [],
  tags = [],
}) {
  const workspace = useBookmarksWorkspace({ bookmarks, folders, tags });
  const [bookmarkEditorOpen, setBookmarkEditorOpen] = useState(false);
  const [folderEditorOpen, setFolderEditorOpen] = useState(false);
  const [importExportOpen, setImportExportOpen] = useState(false);
  const [editingBookmark, setEditingBookmark] = useState(null);
  const [editingFolder, setEditingFolder] = useState(null);
  const [exportData, setExportData] = useState(null);

  useEffect(() => {
    if (!openBookmarkComposerToken) {
      return;
    }
    setEditingBookmark(null);
    setBookmarkEditorOpen(true);
  }, [openBookmarkComposerToken]);

  const currentFolderBookmarks = workspace.filteredBookmarks;

  async function handleOpenBookmark(bookmark) {
    window.open(bookmarkOpenURL(bookmark.id), "_blank", "noopener,noreferrer");
  }

  async function handleSaveBookmark(payload) {
    const successful = await onSaveBookmark(payload);
    if (successful) {
      setBookmarkEditorOpen(false);
      setEditingBookmark(null);
    }
    return successful;
  }

  async function handleSaveFolder(payload) {
    const successful = await onSaveFolder(payload);
    if (successful) {
      setFolderEditorOpen(false);
      setEditingFolder(null);
    }
    return successful;
  }

  async function handleDeleteBookmark(bookmark) {
    if (!window.confirm(`Remove bookmark "${bookmark.name}"?`)) {
      return;
    }
    await onDeleteBookmark(bookmark.id);
  }

  async function handleDeleteFolder(folder) {
    if (!window.confirm(`Delete folder "${folder.name}"? Child folders and bookmarks will be promoted.`)) {
      return;
    }
    await onDeleteFolder(folder.id);
  }

  async function handleToggleFavorite(bookmark) {
    const favoriteCount = bookmarks.filter((item) => item.isFavorite).length;
    await onSaveBookmark({
      ...bookmarkToInput(bookmark),
      favoritePosition: bookmark.isFavorite ? 0 : favoriteCount,
      isFavorite: !bookmark.isFavorite,
    });
  }

  async function handleReorderBookmarks(draggedId, targetId) {
    const targetBookmark = bookmarks.find((bookmark) => bookmark.id === targetId);
    if (!targetBookmark) {
      return;
    }
    const siblings = bookmarks.filter(
      (bookmark) => (bookmark.folderId || "") === (targetBookmark.folderId || ""),
    );
    const ordered = moveItem(siblings, draggedId, targetId);
    await onReorderBookmarks(reorderPayload(bookmarks, ordered, targetBookmark.folderId || ""));
  }

  async function handleDropBookmark(bookmarkId, folderId) {
    const siblings = bookmarks.filter((bookmark) => (bookmark.folderId || "") === folderId);
    const draggedBookmark = bookmarks.find((bookmark) => bookmark.id === bookmarkId);
    if (!draggedBookmark) {
      return;
    }
    const ordered = [...siblings.filter((bookmark) => bookmark.id !== bookmarkId), draggedBookmark];
    await onReorderBookmarks(reorderPayload(bookmarks, ordered, folderId));
  }

  async function handleMoveFolder(folder, direction) {
    const siblings = folders.filter(
      (item) => (item.parentId || "") === (folder.parentId || ""),
    );
    const currentIndex = siblings.findIndex((item) => item.id === folder.id);
    const nextIndex = currentIndex + direction;
    if (currentIndex === -1 || nextIndex < 0 || nextIndex >= siblings.length) {
      return;
    }
    const ordered = [...siblings];
    const [item] = ordered.splice(currentIndex, 1);
    ordered.splice(nextIndex, 0, item);
    await onReorderFolders(
      ordered.map((entry, index) => ({
        id: entry.id,
        parentId: entry.parentId || "",
        position: index,
      })),
    );
  }

  async function handleDropFolder(draggedId, targetParentId) {
    const draggedFolder = folders.find((folder) => folder.id === draggedId);
    if (!draggedFolder) {
      return;
    }
    const sourceParentId = draggedFolder.parentId || "";
    const destinationParentId = targetParentId || "";
    const sourceSiblings = folders
      .filter(
        (folder) =>
          (folder.parentId || "") === sourceParentId && folder.id !== draggedFolder.id,
      )
      .map((folder, index) => ({
        id: folder.id,
        parentId: sourceParentId,
        position: index,
      }));
    const destinationSiblings = folders
      .filter((folder) => (folder.parentId || "") === destinationParentId)
      .filter((folder) => folder.id !== draggedFolder.id);
    destinationSiblings.push(draggedFolder);
    const destinationPayload = destinationSiblings.map((folder, index) => ({
      id: folder.id,
      parentId: destinationParentId,
      position: index,
    }));
    await onReorderFolders(
      sourceParentId === destinationParentId
        ? destinationPayload
        : [...sourceSiblings, ...destinationPayload],
    );
  }

  return (
    <>
      <section className="grid gap-6" id="bookmarks">
        <div className="surface-hero rounded-[34px] border border-line p-6 shadow-card">
          <div className="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
            <div className="max-w-3xl">
              <p className="text-xs font-semibold uppercase tracking-[0.24em] text-accent-strong">
                Primary navigation
              </p>
              <h2 className="mt-3 text-3xl font-semibold tracking-tight text-ink">
                Launch the homelab from one curated workspace
              </h2>
              <p className="mt-3 text-sm leading-7 text-muted">
                Discovery turns services into bookmarks, folders keep them ordered, and health badges make the next click obvious.
              </p>
            </div>
            <div className="grid gap-3 sm:grid-cols-2 xl:min-w-[420px]">
              <div className="rounded-3xl border border-line bg-panel px-4 py-4 shadow-sm">
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted">
                  Total bookmarks
                </p>
                <p className="mt-2 text-3xl font-semibold tracking-tight text-ink">
                  {bookmarks.length}
                </p>
              </div>
              <div className="rounded-3xl border border-line bg-panel px-4 py-4 shadow-sm">
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted">
                  Favorites
                </p>
                <p className="mt-2 text-3xl font-semibold tracking-tight text-ink">
                  {bookmarks.filter((bookmark) => bookmark.isFavorite).length}
                </p>
              </div>
            </div>
          </div>
        </div>

        <BookmarkToolbar
          canManage={canManage}
          favoritesOnly={workspace.favoritesOnly}
          onExport={async () => {
            const payload = await onExportBookmarks?.();
            if (payload) {
              setExportData(payload);
            }
            setImportExportOpen(true);
          }}
          onImport={() => setImportExportOpen(true)}
          onNewBookmark={() => {
            setEditingBookmark(null);
            setBookmarkEditorOpen(true);
          }}
          onSearchChange={workspace.setSearch}
          search={workspace.search}
          setFavoritesOnly={workspace.setFavoritesOnly}
        />

        <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
          <aside className="space-y-5">
            <div className="rounded-[30px] border border-slate-200 bg-slate-50 p-5">
              <FolderTree
                canManage={canManage}
                folders={workspace.folderTree}
                onAdd={() => {
                  setEditingFolder(null);
                  setFolderEditorOpen(true);
                }}
                onDelete={(folder) => void handleDeleteFolder(folder)}
                onDropBookmark={(bookmarkId, folderId) =>
                  void handleDropBookmark(bookmarkId, folderId)
                }
                onDropFolder={(draggedId, folderId) =>
                  void handleDropFolder(draggedId, folderId)
                }
                onEdit={(folder) => {
                  setEditingFolder(folder);
                  setFolderEditorOpen(true);
                }}
                onMoveDown={(folder) => void handleMoveFolder(folder, 1)}
                onMoveUp={(folder) => void handleMoveFolder(folder, -1)}
                onSelect={workspace.setSelectedFolderId}
                selectedFolderId={workspace.selectedFolderId}
              />
            </div>

            <div className="rounded-[30px] border border-slate-200 bg-white p-5 shadow-card">
              <div className="flex items-center gap-3">
                <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-accent/10 text-accent-strong">
                  <FolderIcon className="h-5 w-5" />
                </span>
                <div>
                  <p className="text-sm font-semibold text-slate-950">
                    {workspace.activeFolder?.name || "All folders"}
                  </p>
                  <p className="text-sm text-slate-500">
                    {currentFolderBookmarks.length} bookmark{currentFolderBookmarks.length === 1 ? "" : "s"} match the current view
                  </p>
                </div>
              </div>
            </div>
          </aside>

          <div className="space-y-5">
            <div className="rounded-[30px] border border-slate-200 bg-white p-5 shadow-card">
              <div className="flex items-center justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">
                    Filters
                  </p>
                  <h3 className="mt-1 text-lg font-semibold tracking-tight text-slate-950">
                    Tags and favorites
                  </h3>
                </div>
              </div>
              <div className="mt-4">
                <TagFilterBar
                  onSelect={workspace.setSelectedTag}
                  selectedTag={workspace.selectedTag}
                  tags={workspace.tags}
                />
              </div>
            </div>

            <FavoritesStrip bookmarks={workspace.favorites} onOpen={handleOpenBookmark} />

            <div className="space-y-4">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">
                    Bookmark cards
                  </p>
                  <h3 className="mt-1 text-xl font-semibold tracking-tight text-slate-950">
                    {workspace.activeFolder?.name || "All services"}
                  </h3>
                </div>
              </div>

              <BookmarkGrid
                bookmarks={currentFolderBookmarks}
                canManage={canManage}
                onDelete={(bookmark) => void handleDeleteBookmark(bookmark)}
                onEdit={(bookmark) => {
                  setEditingBookmark(bookmark);
                  setBookmarkEditorOpen(true);
                }}
                onOpen={handleOpenBookmark}
                onReorder={(draggedId, targetId) =>
                  void handleReorderBookmarks(draggedId, targetId)
                }
                onToggleFavorite={(bookmark) => void handleToggleFavorite(bookmark)}
              />
            </div>
          </div>
        </div>
      </section>

      <Modal
        description="Create a manual bookmark, link an existing service, or attach monitoring without leaving the navigation view."
        onClose={() => {
          setBookmarkEditorOpen(false);
          setEditingBookmark(null);
        }}
        open={bookmarkEditorOpen}
        title={editingBookmark ? "Edit bookmark" : "Add bookmark"}
      >
        <BookmarkEditorForm
          bookmark={editingBookmark}
          devices={devices}
          folders={folders}
          onSubmit={(payload) => handleSaveBookmark(payload)}
          onUploadIcon={onUploadBookmarkIcon}
          services={services}
        />
      </Modal>

      <Modal
        description="Create nested folders to shape the navigation tree for your homelab."
        onClose={() => {
          setFolderEditorOpen(false);
          setEditingFolder(null);
        }}
        open={folderEditorOpen}
        title={editingFolder ? "Edit folder" : "Add folder"}
      >
        <FolderEditorForm
          folders={folders}
          initialFolder={editingFolder}
          onSubmit={(payload) => handleSaveFolder(payload)}
        />
      </Modal>

      <Modal
        description="Export the full bookmark workspace to JSON or import a backup from another HomelabWatch instance."
        onClose={() => setImportExportOpen(false)}
        open={importExportOpen}
        title="Import or export bookmarks"
      >
        <ImportExportDialog
          exportData={exportData}
          onExportRefresh={async () => {
            const payload = await onExportBookmarks?.();
            if (payload) {
              setExportData(payload);
            }
          }}
          onImport={onImportBookmarks}
        />
      </Modal>
    </>
  );
}
