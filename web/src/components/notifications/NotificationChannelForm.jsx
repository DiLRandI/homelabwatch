import { useEffect, useState } from "react";

import Button from "../ui/Button";
import Input from "../ui/Input";

export default function NotificationChannelForm({ channel, onCancel, onSubmit }) {
  const [form, setForm] = useState(defaultForm(channel));

  useEffect(() => setForm(defaultForm(channel)), [channel]);

  function updateConfig(key, value) {
    setForm((current) => ({ ...current, config: { ...current.config, [key]: value } }));
  }

  function submit(event) {
    event.preventDefault();
    onSubmit({
      id: channel?.id,
      name: form.name,
      type: form.type,
      enabled: form.enabled,
      config: form.config,
    });
  }

  return (
    <form className="grid gap-4" onSubmit={submit}>
      <div className="grid gap-4 md:grid-cols-2">
        <Input label="Name" onChange={(name) => setForm({ ...form, name })} value={form.name} />
        <label className="grid gap-2 text-sm font-medium text-ink-soft">
          Type
          <select
            className="rounded-2xl border border-line bg-panel-strong px-4 py-3 text-sm text-ink"
            onChange={(event) => setForm(defaultForm({ ...channel, type: event.target.value, name: form.name, enabled: form.enabled }))}
            value={form.type}
          >
            <option value="webhook">Webhook</option>
            <option value="ntfy">ntfy</option>
          </select>
        </label>
      </div>
      {form.type === "webhook" ? (
        <div className="grid gap-4 md:grid-cols-[1fr_160px]">
          <Input label="Webhook URL" onChange={(value) => updateConfig("url", value)} value={form.config.url || ""} />
          <Input label="Timeout seconds" onChange={(value) => updateConfig("timeoutSeconds", Number(value))} type="number" value={String(form.config.timeoutSeconds ?? 10)} />
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          <Input label="Server URL" onChange={(value) => updateConfig("serverUrl", value)} value={form.config.serverUrl || ""} />
          <Input label="Topic" onChange={(value) => updateConfig("topic", value)} value={form.config.topic || ""} />
          <Input label="Token" onChange={(value) => updateConfig("token", value)} type="password" value={form.config.token || ""} />
          <Input label="Priority" onChange={(value) => updateConfig("priority", value)} value={form.config.priority || "default"} />
        </div>
      )}
      <label className="inline-flex items-center gap-2 text-sm text-ink-soft">
        <input checked={form.enabled} onChange={(event) => setForm({ ...form, enabled: event.target.checked })} type="checkbox" />
        Enabled
      </label>
      <div className="flex gap-2">
        <Button type="submit">{channel?.id ? "Save channel" : "Create channel"}</Button>
        {onCancel ? <Button onClick={onCancel} type="button" variant="ghost">Cancel</Button> : null}
      </div>
    </form>
  );
}

function defaultForm(channel) {
  const type = channel?.type || "webhook";
  return {
    enabled: channel?.enabled ?? true,
    name: channel?.name || "",
    type,
    config:
      type === "ntfy"
        ? { priority: "default", ...(channel?.config || {}) }
        : { timeoutSeconds: 10, ...(channel?.config || {}) },
  };
}
