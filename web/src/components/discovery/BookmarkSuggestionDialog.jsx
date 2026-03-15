import { useEffect, useState } from "react";

import Button from "../ui/Button";
import Input from "../ui/Input";
import Modal from "../ui/Modal";

export default function BookmarkSuggestionDialog({
  folders = [],
  item = null,
  onClose,
  onSubmit,
  open = false,
}) {
  const [form, setForm] = useState({
    folderId: "",
    isFavorite: false,
    name: "",
    tags: "",
  });

  useEffect(() => {
    setForm({
      folderId: "",
      isFavorite: false,
      name: item?.name || "",
      tags: "",
    });
  }, [item]);

  async function handleSubmit(event) {
    event.preventDefault();
    if (!item) {
      return;
    }
    const successful = await onSubmit(item, {
      folderId: form.folderId,
      isFavorite: form.isFavorite,
      name: form.name,
      tags: form.tags
        .split(",")
        .map((value) => value.trim())
        .filter(Boolean),
    });
    if (successful) {
      onClose();
    }
  }

  return (
    <Modal
      description="Review the suggested bookmark details before promoting this discovery into the workspace."
      onClose={onClose}
      open={open}
      title="Create bookmark from discovery"
    >
      {item ? (
        <form className="grid gap-4" onSubmit={handleSubmit}>
          <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-600">
            <p className="font-medium text-slate-900">{item.deviceName || item.host}</p>
            <p className="mt-1">{item.url}</p>
          </div>

          <Input
            label="Display name"
            onChange={(value) => setForm((current) => ({ ...current, name: value }))}
            value={form.name}
          />

          <label className="grid gap-2 text-sm font-medium text-slate-700">
            Folder
            <select
              className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-accent focus:ring-4 focus:ring-accent/10"
              onChange={(event) =>
                setForm((current) => ({ ...current, folderId: event.target.value }))
              }
              value={form.folderId}
            >
              <option value="">Unfiled</option>
              {folders.map((folder) => (
                <option key={folder.id} value={folder.id}>
                  {folder.name}
                </option>
              ))}
            </select>
          </label>

          <Input
            label="Tags"
            onChange={(value) => setForm((current) => ({ ...current, tags: value }))}
            placeholder="monitoring, automation"
            value={form.tags}
          />

          <label className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
            <input
              checked={form.isFavorite}
              onChange={(event) =>
                setForm((current) => ({ ...current, isFavorite: event.target.checked }))
              }
              type="checkbox"
            />
            Pin this bookmark to favorites
          </label>

          <div className="flex justify-end gap-3">
            <Button onClick={onClose} type="button" variant="ghost">
              Cancel
            </Button>
            <Button type="submit">Create bookmark</Button>
          </div>
        </form>
      ) : null}
    </Modal>
  );
}
