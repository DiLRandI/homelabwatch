import { useEffect, useState } from "react";

import Button from "../ui/Button";
import Input from "../ui/Input";

export default function FolderEditorForm({
  folders = [],
  initialFolder = null,
  onSubmit,
}) {
  const [name, setName] = useState("");
  const [parentId, setParentId] = useState("");

  useEffect(() => {
    setName(initialFolder?.name || "");
    setParentId(initialFolder?.parentId || "");
  }, [initialFolder]);

  async function handleSubmit(event) {
    event.preventDefault();
    await onSubmit({
      id: initialFolder?.id,
      name,
      parentId,
      position: initialFolder?.position || 0,
    });
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit}>
      <Input
        label="Folder name"
        onChange={setName}
        placeholder="Monitoring"
        value={name}
      />
      <label className="grid gap-2 text-sm font-medium text-slate-700">
        Parent folder
        <select
          className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-accent focus:ring-4 focus:ring-accent/10"
          onChange={(event) => setParentId(event.target.value)}
          value={parentId}
        >
          <option value="">Top level</option>
          {folders
            .filter((folder) => folder.id !== initialFolder?.id)
            .map((folder) => (
              <option key={folder.id} value={folder.id}>
                {folder.name}
              </option>
            ))}
        </select>
      </label>
      <div className="flex justify-end">
        <Button type="submit">{initialFolder ? "Save folder" : "Create folder"}</Button>
      </div>
    </form>
  );
}
