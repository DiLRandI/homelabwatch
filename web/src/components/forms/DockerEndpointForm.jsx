import { useState } from "react";

import { defaultDockerEndpointForm } from "../../lib/forms";
import Button from "../ui/Button";
import Input from "../ui/Input";

export default function DockerEndpointForm({
  onSubmit,
  submitLabel = "Save endpoint",
}) {
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
    <form className="grid gap-4" onSubmit={handleSubmit}>
      <Input
        label="Name"
        autoComplete="off"
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
      <div className="flex justify-end">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
