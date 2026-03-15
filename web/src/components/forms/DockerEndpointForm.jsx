import { useState } from "react";

import { defaultDockerEndpointForm } from "../../lib/forms";
import Input from "../ui/Input";

export default function DockerEndpointForm({ onSubmit }) {
  const [form, setForm] = useState(defaultDockerEndpointForm);

  async function handleSubmit(event) {
    event.preventDefault();
    const successful = await onSubmit({
      ...form,
      scanIntervalSeconds: Number(form.scanIntervalSeconds || 30),
    });
    if (successful) {
      setForm(defaultDockerEndpointForm);
    }
  }

  return (
    <form
      className="grid gap-3 rounded-3xl border border-dashed border-accent/40 bg-white/5 p-4"
      onSubmit={handleSubmit}
    >
      <h3 className="font-display text-lg font-semibold text-ink">
        Add Docker endpoint
      </h3>
      <Input
        label="Name"
        onChange={(value) => setForm((current) => ({ ...current, name: value }))}
        value={form.name}
      />
      <Input
        label="Kind"
        onChange={(value) => setForm((current) => ({ ...current, kind: value }))}
        value={form.kind}
      />
      <Input
        label="Address"
        onChange={(value) =>
          setForm((current) => ({ ...current, address: value }))
        }
        value={form.address}
      />
      <Input
        label="Interval seconds"
        onChange={(value) =>
          setForm((current) => ({ ...current, scanIntervalSeconds: value }))
        }
        value={String(form.scanIntervalSeconds)}
      />
      <button
        className="rounded-full bg-accent px-4 py-3 text-sm font-semibold text-base"
        type="submit"
      >
        Save endpoint
      </button>
    </form>
  );
}
