import { useState } from "react";

import { defaultServiceForm } from "../../lib/forms";
import Input from "../ui/Input";

export default function ManualServiceForm({ onSubmit }) {
  const [form, setForm] = useState(defaultServiceForm);

  async function handleSubmit(event) {
    event.preventDefault();
    const successful = await onSubmit(form);
    if (successful) {
      setForm(defaultServiceForm);
    }
  }

  return (
    <form
      className="mt-5 grid gap-3 rounded-3xl border border-dashed border-accent/40 bg-white/5 p-4"
      onSubmit={handleSubmit}
    >
      <h3 className="font-display text-lg font-semibold text-ink">
        Add manual service
      </h3>
      <Input
        label="Name"
        onChange={(value) => setForm((current) => ({ ...current, name: value }))}
        placeholder="Plex"
        value={form.name}
      />
      <Input
        label="URL"
        onChange={(value) => setForm((current) => ({ ...current, url: value }))}
        placeholder="http://192.168.1.20:32400"
        value={form.url}
      />
      <button
        className="rounded-full bg-accent px-4 py-3 text-sm font-semibold text-base"
        type="submit"
      >
        Save service
      </button>
    </form>
  );
}
