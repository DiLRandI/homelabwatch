import { useEffect, useState } from "react";

import { formatDate } from "../../lib/format";
import Button from "../ui/Button";
import { Card } from "../ui/Card";
import Input from "../ui/Input";
import TextArea from "../ui/TextArea";
import Badge from "../ui/Badge";

const EMPTY = { kind: "info", title: "", message: "", startsAt: "", endsAt: "" };

function toLocalInput(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toISOString().slice(0, 16);
}

function fromLocalInput(value) {
  return value ? new Date(value).toISOString() : undefined;
}

export default function StatusPageAnnouncementsPanel({ canManage, onDelete, onSave, page }) {
  const [draft, setDraft] = useState(EMPTY);

  useEffect(() => setDraft(EMPTY), [page?.id]);

  if (!page) return null;

  async function submit(event) {
    event.preventDefault();
    await onSave(page.id, {
      ...draft,
      startsAt: fromLocalInput(draft.startsAt),
      endsAt: fromLocalInput(draft.endsAt),
    });
    setDraft(EMPTY);
  }

  return (
    <Card className="p-5">
      <h2 className="text-lg font-semibold text-ink">Announcements</h2>
      <div className="mt-4 grid gap-3">
        {(page.announcements || []).map((item) => (
          <div className="rounded-2xl border border-line bg-panel-strong p-4" key={item.id}>
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <Badge>{item.kind}</Badge>
                <h3 className="mt-2 font-medium text-ink">{item.title}</h3>
                <p className="mt-1 text-sm text-muted">{item.message}</p>
                <p className="mt-2 text-xs text-muted">{formatDate(item.startsAt)} to {formatDate(item.endsAt)}</p>
              </div>
              <div className="flex gap-2">
                <Button disabled={!canManage} onClick={() => setDraft({ ...item, startsAt: toLocalInput(item.startsAt), endsAt: toLocalInput(item.endsAt) })} size="sm" variant="secondary">Edit</Button>
                <Button disabled={!canManage} onClick={() => onDelete(page.id, item.id)} size="sm" variant="ghost">Delete</Button>
              </div>
            </div>
          </div>
        ))}
      </div>
      <form className="mt-5 grid gap-4" onSubmit={submit}>
        <div className="grid gap-4 md:grid-cols-3">
          <label className="grid gap-2">
            <span className="text-sm font-medium text-ink-soft">Kind</span>
            <select className="rounded-2xl border border-line bg-panel-strong px-4 py-3 text-sm text-ink" disabled={!canManage} onChange={(event) => setDraft({ ...draft, kind: event.target.value })} value={draft.kind}>
              <option value="info">Info</option>
              <option value="maintenance">Maintenance</option>
              <option value="incident">Incident</option>
              <option value="resolved">Resolved</option>
            </select>
          </label>
          <Input disabled={!canManage} label="Starts" onChange={(startsAt) => setDraft({ ...draft, startsAt })} type="datetime-local" value={draft.startsAt || ""} />
          <Input disabled={!canManage} label="Ends" onChange={(endsAt) => setDraft({ ...draft, endsAt })} type="datetime-local" value={draft.endsAt || ""} />
        </div>
        <Input disabled={!canManage} label="Title" onChange={(title) => setDraft({ ...draft, title })} value={draft.title || ""} />
        <TextArea label="Message" onChange={(message) => setDraft({ ...draft, message })} value={draft.message || ""} />
        <Button disabled={!canManage} type="submit">Save announcement</Button>
      </form>
    </Card>
  );
}
