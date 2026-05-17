import Button from "../ui/Button";
import { Card, CardContent } from "../ui/Card";
import {
  PlusIcon,
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

const priorityLabels = new Set(["discovered", "unhealthy", "degraded"]);

export default function DashboardHeader({
  canManageUI,
  discoveredCount = 0,
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
  const sortedMetrics = [...metrics].sort((left, right) => {
    const leftPriority = priorityLabels.has(String(left.label).toLowerCase());
    const rightPriority = priorityLabels.has(String(right.label).toLowerCase());

    if (leftPriority !== rightPriority) {
      return leftPriority ? -1 : 1;
    }

    return 0;
  });

  return (
    <section
      className="grid gap-4"
      id="overview"
    >
      <Card className="overflow-hidden">
        <CardContent className="p-4">
          <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
            <div className="grid min-w-0 flex-1 gap-3 md:grid-cols-3">
              {details.map((detail) => {
                const Icon = detailIcons[detail.label] || SparklesIcon;
                return (
                  <div
                    className="flex min-w-0 items-center gap-3 rounded-2xl border border-line bg-panel px-3 py-2.5"
                    key={detail.label}
                  >
                    <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-base text-muted">
                      <Icon className="h-4 w-4" />
                    </span>
                    <div className="min-w-0">
                      <p className="text-[0.68rem] font-semibold uppercase tracking-[0.18em] text-muted">
                        {detail.label}
                      </p>
                      <p className="mt-0.5 truncate text-sm font-medium text-ink">
                        {detail.value}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="flex shrink-0 flex-wrap gap-3">
              <Button
                disabled={!canManageUI}
                leadingIcon={PlusIcon}
                onClick={() => onOpenModal("service")}
              >
                Add service
              </Button>
              <Button
                leadingIcon={DiscoveryIcon}
                onClick={() => onOpenModal("discovery")}
                variant={discoveredCount > 0 ? "primary" : "secondary"}
              >
                {discoveredCount > 0
                  ? `Review ${discoveredCount} discovered`
                  : "Review discovery"}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        {sortedMetrics.map((metric) => {
          const isPriority =
            priorityLabels.has(String(metric.label).toLowerCase()) &&
            Number(metric.value) > 0;

          return (
            <Card
              className={isPriority ? "overflow-hidden border-warn/30" : "overflow-hidden"}
              key={metric.label}
            >
              <CardContent className="p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted">
                      {metric.label}
                    </p>
                    <p className="mt-2 text-3xl font-semibold tracking-tight text-ink">
                      {metric.value}
                    </p>
                  </div>
                  <span className={`inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl ${metric.iconTone}`}>
                    <metric.icon className="h-5 w-5" />
                  </span>
                </div>
                <p className="mt-3 line-clamp-2 text-sm leading-6 text-muted">
                  {metric.description}
                </p>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </section>
  );
}
