import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent } from "../ui/Card";
import {
  BookmarkIcon,
  PlusIcon,
  TokenIcon,
  ShieldIcon,
  DatabaseIcon,
  DiscoveryIcon,
  SparklesIcon,
} from "../ui/Icons";

const detailIcons = {
  Appliance: DatabaseIcon,
  Access: ShieldIcon,
  "Discovery footprint": DiscoveryIcon,
};

export default function DashboardHeader({
  canManageUI,
  metrics,
  onOpenModal,
  settings,
}) {
  const details = [
    {
      label: "Appliance",
      value: settings?.appSettings?.applianceName || "HomelabWatch",
    },
    {
      label: "Discovery footprint",
      value: `${settings?.scanTargets?.length ?? 0} scan targets, ${settings?.dockerEndpoints?.length ?? 0} Docker endpoints`,
    },
    {
      label: "Access",
      value: canManageUI
        ? "Trusted LAN writes enabled"
        : "Read-only from this network",
    },
  ];

  return (
    <section
      className="grid gap-6 xl:grid-cols-[minmax(0,1.4fr)_minmax(360px,1fr)]"
      id="overview"
    >
      <Card className="overflow-hidden border-transparent bg-[linear-gradient(135deg,#eff6ff_0%,#ffffff_44%,#f8fafc_100%)]">
        <CardContent className="p-6 sm:p-8">
          <Badge tone="accent" withDot>
            Operations overview
          </Badge>
          <div className="mt-5 max-w-3xl">
            <h1 className="text-3xl font-semibold tracking-tight text-slate-950 sm:text-4xl">
              Operate the lab from one clean control plane.
            </h1>
            <p className="mt-4 text-sm leading-7 text-slate-600 sm:text-base">
              HomelabWatch keeps discovery, runtime health, containers, and
              automation access in one operator console so you can see what
              changed and act without credential juggling in the browser.
            </p>
          </div>

          <div className="mt-6 grid gap-4 sm:grid-cols-3">
            {details.map((detail) => {
              const Icon = detailIcons[detail.label] || SparklesIcon;
              return (
                <div
                  className="rounded-3xl border border-white bg-white/90 p-4 shadow-card"
                  key={detail.label}
                >
                  <span className="inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-slate-100 text-slate-500">
                    <Icon className="h-4 w-4" />
                  </span>
                  <p className="mt-4 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    {detail.label}
                  </p>
                  <p className="mt-2 text-sm font-medium text-slate-900">
                    {detail.value}
                  </p>
                </div>
              );
            })}
          </div>

          <div className="mt-6 flex flex-wrap gap-3">
            <Button
              disabled={!canManageUI}
              leadingIcon={PlusIcon}
              onClick={() => onOpenModal("service")}
            >
              Add service
            </Button>
            <Button
              disabled={!canManageUI}
              leadingIcon={BookmarkIcon}
              onClick={() => onOpenModal("bookmark")}
              variant="secondary"
            >
              Add bookmark
            </Button>
            <Button
              disabled={!canManageUI}
              leadingIcon={TokenIcon}
              onClick={() => onOpenModal("apiToken")}
              variant="subtle"
            >
              Create API token
            </Button>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 sm:grid-cols-2">
        {metrics.map((metric) => (
          <Card className="overflow-hidden" key={metric.label}>
            <CardContent className="p-5">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    {metric.label}
                  </p>
                  <p className="mt-3 text-4xl font-semibold tracking-tight text-slate-950">
                    {metric.value}
                  </p>
                </div>
                <span className={`inline-flex h-11 w-11 items-center justify-center rounded-2xl ${metric.iconTone}`}>
                  <metric.icon className="h-5 w-5" />
                </span>
              </div>
              <p className="mt-4 text-sm text-slate-500">{metric.description}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
