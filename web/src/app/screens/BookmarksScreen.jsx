import BookmarksHome from "../../components/bookmarks/BookmarksHome";

export default function BookmarksScreen({
  bookmarks,
  canManageUI,
  dashboard,
  folders,
  onDeleteBookmark,
  onDeleteFolder,
  onExportBookmarks,
  onImportBookmarks,
  onReorderBookmarks,
  onReorderFolders,
  onSaveBookmark,
  onSaveFolder,
  onUploadBookmarkIcon,
  services,
  tags,
}) {
  return (
    <BookmarksHome
      bookmarks={bookmarks}
      canManage={canManageUI}
      devices={dashboard?.devices ?? []}
      folders={folders}
      onDeleteBookmark={onDeleteBookmark}
      onDeleteFolder={onDeleteFolder}
      onExportBookmarks={onExportBookmarks}
      onImportBookmarks={onImportBookmarks}
      onReorderBookmarks={onReorderBookmarks}
      onReorderFolders={onReorderFolders}
      onSaveBookmark={onSaveBookmark}
      onSaveFolder={onSaveFolder}
      onUploadBookmarkIcon={onUploadBookmarkIcon}
      services={services}
      tags={tags}
    />
  );
}
