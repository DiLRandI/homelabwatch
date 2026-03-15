import { useState } from "react";

import { defaultScanTargetForm, parsePorts } from "../../lib/forms";
import Input from "../ui/Input";

export default function ScanTargetForm({ onSubmit }) {
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
    <form
      className="grid gap-3 rounded-3xl border border-dashed border-accent/40 bg-white/5 p-4"
      onSubmit={handleSubmit}
    >
      <h3 className="font-display text-lg font-semibold text-ink">
        Add scan target
      </h3>
      <Input
        label="Name"
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
      <button
        className="rounded-full bg-accent px-4 py-3 text-sm font-semibold text-base"
        type="submit"
      >
        Save target
      </button>
    </form>
  );
}
