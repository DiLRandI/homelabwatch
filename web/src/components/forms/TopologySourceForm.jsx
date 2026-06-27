import { useEffect, useState } from "react";

import { defaultTopologySourceForm } from "../../lib/forms";
import Button from "../ui/Button";
import Input from "../ui/Input";

const selectClass =
  "w-full rounded-2xl border border-line bg-panel-strong px-4 py-3 text-sm text-ink shadow-sm outline-hidden transition focus:border-accent focus-visible:ring-4 focus-visible:ring-accent/15";

export default function TopologySourceForm({
  item = null,
  onSubmit,
  submitLabel = "Save source",
}) {
  const [form, setForm] = useState(defaultTopologySourceForm);
  const editing = Boolean(item?.id);

  useEffect(() => {
    setForm({
      ...defaultTopologySourceForm,
      ...item,
      community: "",
      authPassphrase: "",
      privacyPassphrase: "",
    });
  }, [item]);

  async function handleSubmit(event) {
    event.preventDefault();
    const payload = {
      ...form,
      enabled: Boolean(form.enabled),
      port: Number(form.port || 161),
      pollIntervalSeconds: Number(form.pollIntervalSeconds || 300),
      retries: Number(form.retries || 1),
      root: Boolean(form.root),
      timeoutMs: Number(form.timeoutMs || 1500),
    };
    if (editing) {
      payload.id = item.id;
      if (!payload.community) delete payload.community;
      if (!payload.authPassphrase) delete payload.authPassphrase;
      if (!payload.privacyPassphrase) delete payload.privacyPassphrase;
    }
    const successful = await onSubmit(payload);
    if (successful && !editing) {
      setForm(defaultTopologySourceForm);
    }
  }

  function update(key, value) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit}>
      <div className="grid gap-4 sm:grid-cols-2">
        <Input
          autoComplete="off"
          label="Name"
          onChange={(value) => update("name", value)}
          value={form.name}
        />
        <Input
          autoComplete="off"
          label="Address"
          onChange={(value) => update("address", value)}
          value={form.address}
        />
        <Input
          label="Port"
          onChange={(value) => update("port", value)}
          type="number"
          value={String(form.port)}
        />
        <Input
          label="Poll interval"
          onChange={(value) => update("pollIntervalSeconds", value)}
          type="number"
          value={String(form.pollIntervalSeconds)}
        />
        <Input
          label="Timeout ms"
          onChange={(value) => update("timeoutMs", value)}
          type="number"
          value={String(form.timeoutMs)}
        />
        <Input
          label="Retries"
          onChange={(value) => update("retries", value)}
          type="number"
          value={String(form.retries)}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <label className="grid gap-2 text-sm font-medium text-ink-soft">
          SNMP version
          <select
            className={selectClass}
            onChange={(event) => update("snmpVersion", event.target.value)}
            value={form.snmpVersion}
          >
            <option value="v2c">v2c</option>
            <option value="v3">v3</option>
          </select>
        </label>
        <label className="grid gap-2 text-sm font-medium text-ink-soft">
          Role
          <select
            className={selectClass}
            onChange={(event) => update("role", event.target.value)}
            value={form.role}
          >
            <option value="router">Router</option>
            <option value="switch">Switch</option>
            <option value="ap">AP</option>
            <option value="unknown">Unknown</option>
          </select>
        </label>
      </div>

      {form.snmpVersion === "v2c" ? (
        <Input
          autoComplete="new-password"
          label="Community"
          onChange={(value) => update("community", value)}
          type="password"
          value={form.community}
        />
      ) : (
        <div className="grid gap-4">
          <Input
            autoComplete="off"
            label="Username"
            onChange={(value) => update("username", value)}
            value={form.username}
          />
          <div className="grid gap-4 sm:grid-cols-2">
            <label className="grid gap-2 text-sm font-medium text-ink-soft">
              Auth protocol
              <select
                className={selectClass}
                onChange={(event) => update("authProtocol", event.target.value)}
                value={form.authProtocol}
              >
                <option value="none">None</option>
                <option value="md5">MD5</option>
                <option value="sha">SHA</option>
                <option value="sha224">SHA-224</option>
                <option value="sha256">SHA-256</option>
                <option value="sha384">SHA-384</option>
                <option value="sha512">SHA-512</option>
              </select>
            </label>
            <Input
              autoComplete="new-password"
              label="Auth passphrase"
              onChange={(value) => update("authPassphrase", value)}
              type="password"
              value={form.authPassphrase}
            />
            <label className="grid gap-2 text-sm font-medium text-ink-soft">
              Privacy protocol
              <select
                className={selectClass}
                onChange={(event) => update("privacyProtocol", event.target.value)}
                value={form.privacyProtocol}
              >
                <option value="none">None</option>
                <option value="des">DES</option>
                <option value="aes">AES</option>
                <option value="aes192">AES-192</option>
                <option value="aes256">AES-256</option>
              </select>
            </label>
            <Input
              autoComplete="new-password"
              label="Privacy passphrase"
              onChange={(value) => update("privacyPassphrase", value)}
              type="password"
              value={form.privacyPassphrase}
            />
          </div>
        </div>
      )}

      <div className="flex flex-wrap gap-3">
        <label className="inline-flex items-center gap-2 text-sm text-ink-soft">
          <input
            checked={form.enabled}
            className="h-4 w-4 accent-accent"
            onChange={(event) => update("enabled", event.target.checked)}
            type="checkbox"
          />
          Enabled
        </label>
        <label className="inline-flex items-center gap-2 text-sm text-ink-soft">
          <input
            checked={form.root}
            className="h-4 w-4 accent-accent"
            onChange={(event) => update("root", event.target.checked)}
            type="checkbox"
          />
          Root
        </label>
      </div>

      <div className="flex justify-end">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
