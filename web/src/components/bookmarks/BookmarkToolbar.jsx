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
    <div className="rounded-[28px] border border-slate-200 bg-white p-4 shadow-card">
      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_auto] xl:items-end">
        <Input
          containerClassName="gap-3"
          inputClassName="pl-11"
          label="Search bookmarks"
          onChange={onSearchChange}
          placeholder="Search by name, tag, device, or service"
          value={search}
        />
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-[repeat(4,auto)]">
          <Button disabled={!canManage} leadingIcon={PlusIcon} onClick={onNewBookmark}>
            Add bookmark
          </Button>
          <Button
            onClick={() => setFavoritesOnly(!favoritesOnly)}
            variant={favoritesOnly ? "secondary" : "ghost"}
          >
            {favoritesOnly ? "Showing favorites" : "Favorites only"}
          </Button>
          <Button leadingIcon={DownloadIcon} onClick={onExport} variant="secondary">
            Export JSON
          </Button>
          <Button disabled={!canManage} leadingIcon={UploadIcon} onClick={onImport} variant="ghost">
            Import JSON
          </Button>
        </div>
      </div>
      <div className="pointer-events-none relative -mt-12 ml-4 hidden text-slate-400 md:block">
        <SearchIcon className="h-4 w-4" />
      </div>
    </div>
  );
}
