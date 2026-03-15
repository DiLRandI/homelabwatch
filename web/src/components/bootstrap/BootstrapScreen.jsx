import { useState } from "react";

import { defaultBootstrapForm, parseCIDRTargets, parsePorts } from "../../lib/forms";
import Alerts from "../ui/Alerts";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent } from "../ui/Card";
import {
  ActivityIcon,
  DatabaseIcon,
  DiscoveryIcon,
  ShieldIcon,
} from "../ui/Icons";
import Input from "../ui/Input";
import TextArea from "../ui/TextArea";

const setupHighlights = [
  {
    body: "Store state in embedded SQLite and keep discovery history without extra services.",
    icon: DatabaseIcon,
    title: "Single-container persistence",
  },
  {
    body: "Seed Docker and LAN discovery immediately so the dashboard fills itself in after bootstrap.",
    icon: DiscoveryIcon,
    title: "Fast infrastructure inventory",
  },
  {
    body: "Use one admin token for write operations while read views stay easy to access.",
    icon: ShieldIcon,
    title: "Operator-friendly access control",
  },
];

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
    <div className="px-4 py-6 sm:px-6 lg:px-8">
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.1fr)_minmax(420px,0.9fr)]">
        <Card className="overflow-hidden border-transparent bg-[linear-gradient(145deg,#0f172a_0%,#0f172a_36%,#1e293b_100%)] text-white shadow-card-lg">
          <CardContent className="p-7 sm:p-9">
            <Badge className="border-white/10 bg-white/10 text-white" withDot>
              Bootstrap required
            </Badge>
            <h1 className="mt-5 max-w-2xl text-4xl font-semibold tracking-tight text-white sm:text-5xl">
              Stand up the control plane in one pass.
            </h1>
            <p className="mt-5 max-w-2xl text-sm leading-7 text-slate-300 sm:text-base">
              Configure the write token, define your initial discovery footprint,
              and let Homelabwatch populate the workspace with Docker and LAN
              infrastructure as soon as it boots.
            </p>

            <div className="mt-8 grid gap-4">
              {setupHighlights.map((item) => (
                <div
                  className="rounded-3xl border border-white/10 bg-white/5 p-5"
                  key={item.title}
                >
                  <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-white/10 text-white">
                    <item.icon className="h-5 w-5" />
                  </span>
                  <h2 className="mt-4 text-lg font-semibold text-white">
                    {item.title}
                  </h2>
                  <p className="mt-2 text-sm leading-6 text-slate-300">
                    {item.body}
                  </p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6 sm:p-8">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-sm font-semibold uppercase tracking-[0.24em] text-accent-strong">
                  Initial setup
                </p>
                <h2 className="mt-3 text-2xl font-semibold tracking-tight text-slate-950">
                  Configure bootstrap settings
                </h2>
                <p className="mt-2 text-sm leading-6 text-slate-500">
                  Docker socket discovery is added automatically when the
                  container can access <code>/var/run/docker.sock</code>.
                </p>
              </div>
              <span className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-accent/10 text-accent-strong">
                <ActivityIcon className="h-5 w-5" />
              </span>
            </div>

            <form className="mt-8 grid gap-4" onSubmit={handleSubmit}>
              <Input
                autoComplete="new-password"
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
              <label className="flex items-start gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4 text-sm text-slate-600">
                <input
                  checked={form.autoScanEnabled}
                  className="mt-1 h-4 w-4 rounded border-slate-300 text-accent focus:ring-accent"
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      autoScanEnabled: event.target.checked,
                    }))
                  }
                  type="checkbox"
                />
                <span>
                  <span className="block font-medium text-slate-900">
                    Enable automatic LAN scans
                  </span>
                  <span className="mt-1 block text-sm leading-6 text-slate-500">
                    Keep the network inventory fresh after bootstrap without
                    manual intervention.
                  </span>
                </span>
              </label>
              <div className="flex flex-col gap-3 border-t border-slate-200 pt-4 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm leading-6 text-slate-500">
                  The admin token is required for every write endpoint, including
                  manual services, bookmarks, and discovery settings.
                </p>
                <Button type="submit">Initialize workspace</Button>
              </div>
            </form>
            <div className="mt-5">
              <Alerts error={error} notice={notice} />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
