import { useState } from "react";

import { defaultScanTargetForm, parsePorts } from "../../lib/forms";
import Button from "../ui/Button";
import Input from "../ui/Input";

export default function ScanTargetForm({ onSubmit, submitLabel = "Save target" }) {
  const [form, setForm] = useState(defaultScanTargetForm);

  async function handleSubmit(event) {
    event.preventDefault();
    const successful = await onSubmit({
      ...form,
      commonPorts: parsePorts(form.commonPorts),
      scanIntervalSeconds: Number(form.scanIntervalSeconds || 300),
    });
    if (successful) {
      setForm(defaultScanTargetForm);
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
        label="CIDR"
        onChange={(value) => setForm((current) => ({ ...current, cidr: value }))}
        value={form.cidr}
      />
      <Input
        label="Common ports"
        onChange={(value) =>
          setForm((current) => ({ ...current, commonPorts: value }))
        }
        value={form.commonPorts}
      />
      <div className="flex justify-end">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
