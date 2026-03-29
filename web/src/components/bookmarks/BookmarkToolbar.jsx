import Button from "../ui/Button";
import Input from "../ui/Input";
import { DownloadIcon, PlusIcon, SearchIcon, UploadIcon } from "../ui/Icons";

export default function BookmarkToolbar({
  canManage = true,
  favoritesOnly,
  onExport,
  onImport,
  onNewBookmark,
  onSearchChange,
  search,
  setFavoritesOnly,
}) {
  return (
    <div className="rounded-[28px] border border-line bg-panel p-4 shadow-card">
      <div className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
        <div className="min-w-0 flex-1">
          <div className="relative">
            <Input
              containerClassName="gap-3"
              inputClassName="pl-11"
              label="Search bookmarks"
              onChange={onSearchChange}
              placeholder="Search by name, tag, device, or service"
              value={search}
            />
            <div className="pointer-events-none absolute bottom-3.5 left-4 text-copy-subtle">
              <SearchIcon className="h-4 w-4" />
            </div>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-3 xl:justify-end xl:pl-4">
          <Button
            className="shrink-0 whitespace-nowrap"
            disabled={!canManage}
            leadingIcon={PlusIcon}
            onClick={onNewBookmark}
          >
            Add bookmark
          </Button>
          <Button
            className="shrink-0 whitespace-nowrap"
            onClick={() => setFavoritesOnly(!favoritesOnly)}
            variant={favoritesOnly ? "secondary" : "subtle"}
          >
            {favoritesOnly ? "Showing favorites" : "Favorites only"}
          </Button>
          <Button
            className="shrink-0 whitespace-nowrap"
            leadingIcon={DownloadIcon}
            onClick={onExport}
            variant="secondary"
          >
            Export JSON
          </Button>
          <Button
            className="shrink-0 whitespace-nowrap"
            disabled={!canManage}
            leadingIcon={UploadIcon}
            onClick={onImport}
            variant="subtle"
          >
            Import JSON
          </Button>
        </div>
      </div>
    </div>
  );
}
