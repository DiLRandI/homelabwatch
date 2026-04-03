import { useState } from "react";

import { defaultServiceForm } from "../../lib/forms";
import Button from "../ui/Button";
import Input from "../ui/Input";

export default function ManualServiceForm({ onSubmit, submitLabel = "Save service" }) {
  const [form, setForm] = useState(defaultServiceForm);

  async function handleSubmit(event) {
    event.preventDefault();
    const successful = await onSubmit(form);
    if (successful) {
      setForm(defaultServiceForm);
    }
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit}>
      <Input
        label="Name"
        autoComplete="off"
        onChange={(value) => setForm((current) => ({ ...current, name: value }))}
        placeholder="Plex"
        value={form.name}
      />
      <Input
        label="URL"
        autoComplete="url"
        onChange={(value) => setForm((current) => ({ ...current, url: value }))}
        placeholder="http://192.168.1.20:32400"
        value={form.url}
      />
      <Input
        label="Health URL"
        autoComplete="url"
        onChange={(value) =>
          setForm((current) => ({ ...current, healthUrl: value }))
        }
        placeholder="http://192.168.1.20:32400/identity"
        value={form.healthUrl}
      />
      <p className="-mt-1 text-sm leading-6 text-slate-500">
        Leave blank to monitor the open URL. If this differs, HomelabWatch will
        use it as the explicit health target.
      </p>
      <div className="flex justify-end">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
