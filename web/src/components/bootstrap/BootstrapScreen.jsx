import { useState } from "react";

import Alerts from "../ui/Alerts";
import Input from "../ui/Input";
import TextArea from "../ui/TextArea";
import { defaultBootstrapForm, parseCIDRTargets, parsePorts } from "../../lib/forms";

export default function BootstrapScreen({ error, notice, onSubmit }) {
  const [form, setForm] = useState(defaultBootstrapForm);

  async function handleSubmit(event) {
    event.preventDefault();
    const successful = await onSubmit({
      adminToken: form.adminToken,
      autoScanEnabled: form.autoScanEnabled,
      defaultScanPorts: parsePorts(form.defaultScanPorts),
      scanTargets: parseCIDRTargets(form.seedCIDRs, form.defaultScanPorts),
    });

    if (successful) {
      setForm(defaultBootstrapForm);
    }
  }

  return (
    <section className="mx-auto max-w-3xl animate-floatIn rounded-4xl border border-white/10 bg-panel/80 p-8 shadow-halo backdrop-blur-sm">
      <div className="mb-8 flex items-end justify-between gap-4">
        <div>
          <p className="text-sm uppercase tracking-[0.35em] text-accent">
            Homelabwatch
          </p>
          <h1 className="mt-2 font-display text-4xl font-semibold text-ink">
            Single-container homelab control plane
          </h1>
        </div>
        <div className="rounded-full border border-white/10 bg-white/5 px-4 py-2 text-xs text-muted">
          Bootstrap required
        </div>
      </div>
      <p className="max-w-2xl text-sm leading-7 text-muted">
        Initialize the embedded database, set the write token, and optionally
        seed scan targets. Docker socket discovery is added automatically when
        the container has access to <code>/var/run/docker.sock</code>.
      </p>
      <form className="mt-8 grid gap-4 md:grid-cols-2" onSubmit={handleSubmit}>
        <Input
          label="Admin token"
          onChange={(value) =>
            setForm((current) => ({ ...current, adminToken: value }))
          }
          placeholder="choose-a-long-random-token"
          type="password"
          value={form.adminToken}
        />
        <Input
          label="Default ports"
          onChange={(value) =>
            setForm((current) => ({ ...current, defaultScanPorts: value }))
          }
          placeholder="22,80,443,8080,8443"
          value={form.defaultScanPorts}
        />
        <TextArea
          label="Optional seed CIDRs"
          onChange={(value) =>
            setForm((current) => ({ ...current, seedCIDRs: value }))
          }
          placeholder="192.168.1.0/24"
          value={form.seedCIDRs}
        />
        <label className="rounded-3xl border border-white/10 bg-white/5 p-4 text-sm text-ink">
          <span className="block text-xs uppercase tracking-[0.24em] text-muted">
            Discovery policy
          </span>
          <span className="mt-2 flex items-center gap-3">
            <input
              checked={form.autoScanEnabled}
              className="h-4 w-4 accent-accent"
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  autoScanEnabled: event.target.checked,
                }))
              }
              type="checkbox"
            />
            Enable automatic LAN scans after bootstrap
          </span>
        </label>
        <div className="md:col-span-2 flex items-center justify-between gap-4 rounded-3xl border border-white/10 bg-base/70 p-4">
          <div className="text-sm text-muted">
            The token is required for every write endpoint, including manual
            services, bookmarks, and discovery settings.
          </div>
          <button
            className="rounded-full bg-accent px-5 py-3 text-sm font-semibold text-base transition hover:brightness-110"
            type="submit"
          >
            Initialize
          </button>
        </div>
      </form>
      <Alerts error={error} notice={notice} />
    </section>
  );
}
