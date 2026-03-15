import { useState } from "react";

import { defaultAPITokenForm } from "../../lib/forms";
import Button from "../ui/Button";
import Input from "../ui/Input";

export default function APITokenForm({
  onSubmit,
  submitLabel = "Create token",
}) {
  const [form, setForm] = useState(defaultAPITokenForm);

  async function handleSubmit(event) {
    event.preventDefault();
    const successful = await onSubmit({
      name: form.name.trim(),
      scope: form.scope,
    });
    if (successful) {
      setForm(defaultAPITokenForm);
    }
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit}>
      <Input
        autoComplete="off"
        label="Token name"
        onChange={(value) => setForm((current) => ({ ...current, name: value }))}
        placeholder="Automation token"
        value={form.name}
      />

      <label className="grid gap-2">
        <span className="block text-sm font-medium text-slate-700">Scope</span>
        <select
          className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-950 shadow-sm outline-hidden transition focus:border-accent focus-visible:ring-4 focus-visible:ring-accent/15"
          onChange={(event) =>
            setForm((current) => ({ ...current, scope: event.target.value }))
          }
          value={form.scope}
        >
          <option value="write">Write</option>
          <option value="read">Read</option>
        </select>
      </label>

      <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4 text-sm leading-6 text-slate-600">
        Write tokens can run discovery, add services, and change settings. Read
        tokens are better for custom dashboards, alerts, or exports.
      </div>

      <div className="flex justify-end">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
