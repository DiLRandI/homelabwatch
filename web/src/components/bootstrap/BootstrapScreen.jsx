import { useMemo, useState } from "react";

import { defaultSetupForm, parseCIDRTargets, parsePorts } from "../../lib/forms";
import Alerts from "../ui/Alerts";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent } from "../ui/Card";
import {
  ActivityIcon,
  DatabaseIcon,
  DiscoveryIcon,
  ShieldIcon,
  SparklesIcon,
} from "../ui/Icons";
import Input from "../ui/Input";
import TextArea from "../ui/TextArea";

const steps = [
  {
    description: "Name the appliance and confirm where operator writes are allowed.",
    id: "identity",
    title: "Appliance setup",
  },
  {
    description: "Define discovery defaults for Docker and LAN inventory collection.",
    id: "discovery",
    title: "Discovery defaults",
  },
  {
    description: "Review the plan and launch the first discovery cycle.",
    id: "launch",
    title: "Launch discovery",
  },
];

const setupHighlights = [
  {
    body: "Run the UI locally or on your LAN without layering user-account management onto a single-tenant appliance.",
    icon: ShieldIcon,
    title: "Trusted console model",
  },
  {
    body: "Seed Docker and network discovery once, then let background workers keep the inventory current.",
    icon: DiscoveryIcon,
    title: "Operational by default",
  },
  {
    body: "Generate external automation tokens later from settings instead of pasting secrets into the dashboard.",
    icon: DatabaseIcon,
    title: "Cleaner API access",
  },
];

function StepPill({ active, complete, index, step }) {
  return (
    <div className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
      <span
        className={`inline-flex h-9 w-9 items-center justify-center rounded-xl text-sm font-semibold ${
          active
            ? "bg-accent text-white"
            : complete
              ? "bg-ok/10 text-ok-strong"
              : "bg-slate-100 text-slate-500"
        }`}
      >
        {index + 1}
      </span>
      <div className="min-w-0">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
          Step {index + 1}
        </p>
        <p className="mt-1 text-sm font-medium text-slate-900">{step.title}</p>
      </div>
    </div>
  );
}

export default function BootstrapScreen({
  error,
  notice,
  onSubmit,
  trustedNetwork,
}) {
  const [form, setForm] = useState(defaultSetupForm);
  const [stepIndex, setStepIndex] = useState(0);
  const ports = useMemo(
    () => parsePorts(form.defaultScanPorts).join(", "),
    [form.defaultScanPorts],
  );
  const targets = useMemo(
    () => parseCIDRTargets(form.seedCIDRs, form.defaultScanPorts),
    [form.defaultScanPorts, form.seedCIDRs],
  );

  async function handleSubmit(event) {
    event.preventDefault();
    const successful = await onSubmit({
      applianceName: form.applianceName,
      autoScanEnabled: form.autoScanEnabled,
      defaultScanPorts: parsePorts(form.defaultScanPorts),
      dockerEndpoints: [],
      scanTargets: targets,
      runDiscovery: form.runDiscovery,
    });

    if (successful) {
      setForm(defaultSetupForm);
      setStepIndex(0);
    }
  }

  const currentStep = steps[stepIndex];

  return (
    <div className="px-4 py-6 sm:px-6 lg:px-8">
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.12fr)_minmax(420px,0.88fr)]">
        <Card className="overflow-hidden border-transparent bg-[linear-gradient(145deg,#0f172a_0%,#13233f_38%,#1d4ed8_100%)] text-white shadow-card-lg">
          <CardContent className="p-7 sm:p-9">
            <Badge className="border-white/10 bg-white/10 text-white" withDot>
              First-run setup
            </Badge>
            <h1 className="mt-5 max-w-2xl text-4xl font-semibold tracking-tight text-white sm:text-5xl">
              Bring the lab online with a guided control-plane setup.
            </h1>
            <p className="mt-5 max-w-2xl text-sm leading-7 text-slate-200 sm:text-base">
              HomelabWatch is tuned for a trusted local console. Configure the
              appliance once, seed discovery defaults, and move straight into
              operations without juggling bootstrap secrets in the browser.
            </p>

            <div className="mt-8 grid gap-4">
              {setupHighlights.map((item) => (
                <div
                  className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
                  key={item.title}
                >
                  <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-white/10 text-white">
                    <item.icon className="h-5 w-5" />
                  </span>
                  <h2 className="mt-4 text-lg font-semibold text-white">
                    {item.title}
                  </h2>
                  <p className="mt-2 text-sm leading-6 text-slate-200">
                    {item.body}
                  </p>
                </div>
              ))}
            </div>

            <div className="mt-8 rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm">
              <div className="flex items-center gap-3">
                <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-white/10 text-white">
                  <SparklesIcon className="h-5 w-5" />
                </span>
                <div>
                  <p className="text-sm font-medium text-white">
                    {trustedNetwork ? "Trusted network detected" : "Read-only network"}
                  </p>
                  <p className="mt-1 text-sm text-slate-200">
                    {trustedNetwork
                      ? "Setup actions and later dashboard writes are enabled from this network."
                      : "You can review the UI here, but setup must be completed from a trusted local or LAN client."}
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6 sm:p-8">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-sm font-semibold uppercase tracking-[0.24em] text-accent-strong">
                  Setup wizard
                </p>
                <h2 className="mt-3 text-2xl font-semibold tracking-tight text-slate-950">
                  {currentStep.title}
                </h2>
                <p className="mt-2 text-sm leading-6 text-slate-500">
                  {currentStep.description}
                </p>
              </div>
              <span className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-accent/10 text-accent-strong">
                <ActivityIcon className="h-5 w-5" />
              </span>
            </div>

            <div className="mt-6 grid gap-3">
              {steps.map((step, index) => (
                <StepPill
                  active={stepIndex === index}
                  complete={stepIndex > index}
                  index={index}
                  key={step.id}
                  step={step}
                />
              ))}
            </div>

            <form className="mt-8 grid gap-5" onSubmit={handleSubmit}>
              {stepIndex === 0 ? (
                <>
                  <Input
                    autoComplete="organization"
                    label="Appliance name"
                    onChange={(value) =>
                      setForm((current) => ({ ...current, applianceName: value }))
                    }
                    placeholder="Rack Alpha"
                    value={form.applianceName}
                  />
                  <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4 text-sm leading-6 text-slate-600">
                    Browser reads remain open, but write operations are limited
                    to trusted local networks and same-origin requests. External
                    API tokens are created later from settings for automation.
                  </div>
                </>
              ) : null}

              {stepIndex === 1 ? (
                <>
                  <Input
                    label="Default scan ports"
                    onChange={(value) =>
                      setForm((current) => ({
                        ...current,
                        defaultScanPorts: value,
                      }))
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
                        Keep LAN discovery scheduled after setup
                      </span>
                      <span className="mt-1 block text-sm leading-6 text-slate-500">
                        Docker socket discovery is seeded automatically when the
                        container can access <code>/var/run/docker.sock</code>.
                      </span>
                    </span>
                  </label>
                </>
              ) : null}

              {stepIndex === 2 ? (
                <div className="grid gap-4">
                  <div className="rounded-3xl border border-slate-200 bg-slate-50 p-5">
                    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                      Appliance
                    </p>
                    <p className="mt-2 text-lg font-semibold text-slate-950">
                      {form.applianceName || "HomelabWatch"}
                    </p>
                  </div>
                  <div className="grid gap-4 sm:grid-cols-2">
                    <div className="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
                      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                        Scan profile
                      </p>
                      <p className="mt-2 text-sm font-medium text-slate-900">
                        Ports: {ports || "22, 80, 443"}
                      </p>
                      <p className="mt-2 text-sm text-slate-500">
                        {targets.length > 0
                          ? `${targets.length} seeded network target${targets.length > 1 ? "s" : ""}`
                          : "No manual CIDRs provided; suggested targets will be derived automatically."}
                      </p>
                    </div>
                    <div className="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
                      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                        Launch mode
                      </p>
                      <label className="mt-3 flex items-start gap-3 text-sm text-slate-600">
                        <input
                          checked={form.runDiscovery}
                          className="mt-1 h-4 w-4 rounded border-slate-300 text-accent focus:ring-accent"
                          onChange={(event) =>
                            setForm((current) => ({
                              ...current,
                              runDiscovery: event.target.checked,
                            }))
                          }
                          type="checkbox"
                        />
                        <span>
                          <span className="block font-medium text-slate-900">
                            Run discovery immediately after setup
                          </span>
                          <span className="mt-1 block text-sm leading-6 text-slate-500">
                            Start with a populated dashboard instead of an empty
                            shell.
                          </span>
                        </span>
                      </label>
                    </div>
                  </div>
                </div>
              ) : null}

              <Alerts error={error} notice={notice} />

              <div className="flex flex-col gap-3 border-t border-slate-200 pt-5 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm leading-6 text-slate-500">
                  Step {stepIndex + 1} of {steps.length}. You can adjust
                  discovery settings later from the dashboard.
                </p>
                <div className="flex flex-wrap gap-3">
                  {stepIndex > 0 ? (
                    <Button onClick={() => setStepIndex((value) => value - 1)} variant="ghost">
                      Back
                    </Button>
                  ) : null}
                  {stepIndex < steps.length - 1 ? (
                    <Button
                      disabled={!trustedNetwork && stepIndex === 0}
                      onClick={() => setStepIndex((value) => value + 1)}
                      type="button"
                    >
                      Continue
                    </Button>
                  ) : (
                    <Button disabled={!trustedNetwork} type="submit">
                      Initialize workspace
                    </Button>
                  )}
                </div>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
