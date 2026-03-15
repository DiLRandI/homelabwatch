import { useState } from "react";

import { defaultBookmarkForm } from "../../lib/forms";
import Button from "../ui/Button";
import Input from "../ui/Input";
import TextArea from "../ui/TextArea";

export default function BookmarkForm({ onSubmit, submitLabel = "Save bookmark" }) {
  const [form, setForm] = useState(defaultBookmarkForm);

  async function handleSubmit(event) {
    event.preventDefault();
    const successful = await onSubmit(form);
    if (successful) {
      setForm(defaultBookmarkForm);
    }
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit}>
      <Input
        label="Name"
        autoComplete="off"
        onChange={(value) => setForm((current) => ({ ...current, name: value }))}
        value={form.name}
      />
      <Input
        label="URL"
        autoComplete="url"
        onChange={(value) => setForm((current) => ({ ...current, url: value }))}
        value={form.url}
      />
      <TextArea
        label="Description"
        onChange={(value) =>
          setForm((current) => ({ ...current, description: value }))
        }
        value={form.description}
      />
      <div className="flex justify-end">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
