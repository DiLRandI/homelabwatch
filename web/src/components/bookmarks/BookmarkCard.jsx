import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { ArrowUpRightIcon, EditIcon, PinIcon, TrashIcon } from "../ui/Icons";
import StatusBadge from "../ui/StatusBadge";

function iconFallback(name) {
  return (name || "?")
    .split(" ")
    .slice(0, 2)
    .map((item) => item[0] || "")
    .join("")
    .toUpperCase();
}

export default function BookmarkCard({
  bookmark,
  canManage = true,
  onDelete,
  onEdit,
  onOpen,
  onToggleFavorite,
}) {
  return (
    <article className="flex h-full flex-col rounded-[30px] border border-slate-200 bg-white p-5 shadow-card transition hover:-translate-y-0.5 hover:border-slate-300 hover:shadow-card-lg">
      <div className="flex items-start justify-between gap-4">
        <div className="flex min-w-0 items-center gap-3">
          {bookmark.icon ? (
            <img
              alt=""
              className="h-12 w-12 rounded-2xl border border-slate-200 bg-white object-contain p-2"
              src={bookmark.icon}
            />
          ) : (
            <span className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-accent/10 text-sm font-semibold text-accent-strong">
              {iconFallback(bookmark.name)}
            </span>
          )}
          <div className="min-w-0">
            <h3 className="truncate text-lg font-semibold tracking-tight text-slate-950">
              {bookmark.name}
            </h3>
            <p className="mt-1 truncate text-sm text-slate-500" title={bookmark.url}>
              {bookmark.deviceName || bookmark.folderName || "Independent link"}
            </p>
          </div>
        </div>
        <button
          disabled={!canManage}
          className={`rounded-2xl border px-3 py-2 text-sm font-medium transition ${
            bookmark.isFavorite
              ? "border-amber-200 bg-amber-50 text-amber-700"
              : "border-slate-200 bg-white text-slate-500 hover:border-slate-300 hover:text-slate-900"
          }`}
          onClick={() => onToggleFavorite(bookmark)}
          type="button"
        >
          <span className="flex items-center gap-2">
            <PinIcon className="h-4 w-4" />
            {bookmark.isFavorite ? "Pinned" : "Pin"}
          </span>
        </button>
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-2">
        <StatusBadge status={bookmark.healthStatus || "unknown"} subtle />
        <Badge>{bookmark.folderName || "Unfiled"}</Badge>
        <Badge>{bookmark.deviceName || bookmark.host || "No device"}</Badge>
      </div>

      <div className="mt-4 rounded-3xl border border-slate-100 bg-slate-50 p-4">
        <p className="truncate text-sm font-medium text-slate-900" title={bookmark.url}>
          {bookmark.url}
        </p>
        <p className="mt-2 text-sm leading-6 text-slate-500">
          {bookmark.description || "Launch straight into the service without remembering the port."}
        </p>
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        {(bookmark.tags || []).map((tag) => (
          <Badge key={`${bookmark.id}-${tag}`}>{tag}</Badge>
        ))}
      </div>

      <dl className="mt-5 grid gap-3 text-sm text-slate-600 sm:grid-cols-2">
        <div className="rounded-2xl border border-slate-100 bg-slate-50 px-4 py-3">
          <dt className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
            Last opened
          </dt>
          <dd className="mt-2 font-medium text-slate-900">
            {formatDate(bookmark.lastOpenedAt)}
          </dd>
        </div>
        <div className="rounded-2xl border border-slate-100 bg-slate-50 px-4 py-3">
          <dt className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
            Launches
          </dt>
          <dd className="mt-2 font-medium text-slate-900">{bookmark.clickCount || 0}</dd>
        </div>
      </dl>

      <div className="mt-5 flex flex-wrap items-center gap-2">
        <Button onClick={() => onOpen(bookmark)} trailingIcon={ArrowUpRightIcon}>
          Open
        </Button>
        <Button disabled={!canManage} onClick={() => onEdit(bookmark)} variant="secondary">
          <span className="flex items-center gap-2">
            <EditIcon className="h-4 w-4" />
            Edit
          </span>
        </Button>
        <Button disabled={!canManage} onClick={() => onDelete(bookmark)} variant="ghost">
          <span className="flex items-center gap-2">
            <TrashIcon className="h-4 w-4" />
            Remove
          </span>
        </Button>
      </div>
    </article>
  );
}
