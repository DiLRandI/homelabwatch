import { useEffect, useState } from "react";

import Button from "../ui/Button";
import { Card } from "../ui/Card";
import Input from "../ui/Input";
import TextArea from "../ui/TextArea";

const EMPTY_PAGE = { description: "", enabled: true, slug: "", title: "" };

export default function StatusPageEditor({ canManage, onDelete, onOpenPublic, onSave, page }) {
  const [draft, setDraft] = useState(EMPTY_PAGE);

  useEffect(() => {
    setDraft(page || EMPTY_PAGE);
  }, [page?.id]);

  if (!page) {
    return (
      <Card className="p-5">
        <p className="text-sm text-muted">Create or select a status page.</p>
      </Card>
    );
  }

  async function submit(event) {
    event.preventDefault();
    await onSave(draft);
  }

  return (
    <Card className="p-5">
      <form className="grid gap-4" onSubmit={submit}>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold text-ink">Details</h2>
          <label className="flex items-center gap-2 text-sm text-ink-soft">
            <input
              checked={Boolean(draft.enabled)}
              disabled={!canManage}
              onChange={(event) => setDraft({ ...draft, enabled: event.target.checked })}
              type="checkbox"
            />
            Enabled
          </label>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <Input disabled={!canManage} label="Title" onChange={(title) => setDraft({ ...draft, title })} value={draft.title || ""} />
          <Input disabled={!canManage} label="Slug" onChange={(slug) => setDraft({ ...draft, slug })} value={draft.slug || ""} />
        </div>
        <TextArea label="Description" onChange={(description) => setDraft({ ...draft, description })} value={draft.description || ""} />
        <div className="flex flex-wrap justify-between gap-3">
          <div className="flex gap-3">
            <Button disabled={!canManage} type="submit">Save</Button>
            <Button onClick={onOpenPublic} variant="secondary">Open public page</Button>
          </div>
          <Button disabled={!canManage} onClick={() => onDelete(page.id)} variant="ghost">Delete</Button>
        </div>
      </form>
    </Card>
  );
}
