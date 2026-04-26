import { useEffect, useState } from "react";

import Button from "../ui/Button";
import Input from "../ui/Input";

const EVENT_TYPES = [
  "service_health_changed",
  "check_failed",
  "check_recovered",
  "discovered_service_created",
  "device_created",
  "worker_failed",
];

export default function NotificationRuleForm({ channels, onCancel, onSubmit, rule }) {
  const [form, setForm] = useState(defaultForm(rule));
  useEffect(() => setForm(defaultForm(rule)), [rule]);

  function toggleChannel(id) {
    setForm((current) => ({
      ...current,
      channelIds: current.channelIds.includes(id)
        ? current.channelIds.filter((item) => item !== id)
        : [...current.channelIds, id],
    }));
  }

  function submit(event) {
    event.preventDefault();
    const filters = {};
    if (form.eventType === "service_health_changed" && form.statuses.length > 0) {
      filters.statuses = form.statuses;
    }
    if (form.eventType === "worker_failed") {
      filters.minConsecutiveFailures = Number(form.minConsecutiveFailures || 3);
    }
    onSubmit({ id: rule?.id, name: form.name, eventType: form.eventType, enabled: form.enabled, channelIds: form.channelIds, filters });
  }

  return (
    <form className="grid gap-4" onSubmit={submit}>
      <div className="grid gap-4 md:grid-cols-2">
        <Input label="Name" onChange={(name) => setForm({ ...form, name })} value={form.name} />
        <label className="grid gap-2 text-sm font-medium text-ink-soft">
          Event type
          <select className="rounded-2xl border border-line bg-panel-strong px-4 py-3 text-sm text-ink" onChange={(event) => setForm({ ...form, eventType: event.target.value })} value={form.eventType}>
            {EVENT_TYPES.map((type) => <option key={type} value={type}>{type}</option>)}
          </select>
        </label>
      </div>
      {form.eventType === "service_health_changed" ? (
        <div className="flex flex-wrap gap-3">
          {["healthy", "degraded", "unhealthy", "unknown"].map((status) => (
            <label className="inline-flex items-center gap-2 text-sm text-ink-soft" key={status}>
              <input checked={form.statuses.includes(status)} onChange={() => setForm((current) => ({ ...current, statuses: current.statuses.includes(status) ? current.statuses.filter((item) => item !== status) : [...current.statuses, status] }))} type="checkbox" />
              {status}
            </label>
          ))}
        </div>
      ) : null}
      {form.eventType === "worker_failed" ? (
        <Input label="Minimum consecutive failures" onChange={(value) => setForm({ ...form, minConsecutiveFailures: value })} type="number" value={String(form.minConsecutiveFailures)} />
      ) : null}
      <div className="grid gap-2">
        <span className="text-sm font-medium text-ink-soft">Channels</span>
        <div className="flex flex-wrap gap-3">
          {channels.map((channel) => (
            <label className="inline-flex items-center gap-2 text-sm text-ink-soft" key={channel.id}>
              <input checked={form.channelIds.includes(channel.id)} onChange={() => toggleChannel(channel.id)} type="checkbox" />
              {channel.name}
            </label>
          ))}
        </div>
      </div>
      <label className="inline-flex items-center gap-2 text-sm text-ink-soft">
        <input checked={form.enabled} onChange={(event) => setForm({ ...form, enabled: event.target.checked })} type="checkbox" />
        Enabled
      </label>
      <div className="flex gap-2">
        <Button disabled={channels.length === 0} type="submit">{rule?.id ? "Save rule" : "Create rule"}</Button>
        {onCancel ? <Button onClick={onCancel} type="button" variant="ghost">Cancel</Button> : null}
      </div>
    </form>
  );
}

function defaultForm(rule) {
  return {
    channelIds: rule?.channelIds || [],
    enabled: rule?.enabled ?? true,
    eventType: rule?.eventType || "service_health_changed",
    minConsecutiveFailures: rule?.filters?.minConsecutiveFailures ?? 3,
    name: rule?.name || "",
    statuses: rule?.filters?.statuses || [],
  };
}
