import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import StatusBadge from "../ui/StatusBadge";
import { cn } from "../../lib/cn";
import { BookmarkIcon, DiscoveryIcon, RefreshIcon } from "../ui/Icons";

function sourceTone(source) {
  switch (source) {
    case "docker":
      return "info";
    case "mdns":
      return "accent";
    default:
      return "neutral";
  }
}

export default function DiscoveredServicesPanel({
  canManage = true,
  compact = false,
  items = [],
  totalItems = items.length,
  onCreateBookmark,
  onIgnore,
  onRestore,
}) {
  return (
    <section id="discovered-services">
      <Card>
        <CardHeader
          action={
            items.length > 0 ? (
              <Badge tone="warning" withDot>
                {totalItems} pending
              </Badge>
            ) : null
          }
          className={compact ? "py-3" : undefined}
          description="Triage suggestions from Docker, LAN, and mDNS before creating bookmarks."
          title="Discovered services"
        />
        <CardContent className={compact ? "p-0" : undefined}>
          {items.length === 0 ? (
            <EmptyState
              body="Run discovery to collect bookmark suggestions from containers and network services."
              title="No pending suggestions"
            />
          ) : (
            <div className="divide-y divide-line">
              {items.map((item) => (
                <article
                  className={cn(
                    "grid gap-4 px-5 py-4 sm:px-6 xl:grid-cols-[minmax(260px,1.1fr)_minmax(180px,0.7fr)_minmax(180px,0.7fr)_auto] xl:items-center",
                    "transition hover:bg-panel",
                  )}
                  key={item.id}
                >
                  <div className="flex min-w-0 items-center gap-3">
                    <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-warn/12 text-warn-strong">
                      <DiscoveryIcon className="h-5 w-5" />
                    </span>
                    <div className="min-w-0">
                      <div className="flex min-w-0 flex-wrap items-center gap-2">
                        <h3 className="truncate text-base font-semibold tracking-tight text-ink">
                          {item.name}
                        </h3>
                        <StatusBadge status={item.status} />
                      </div>
                      <p className="mt-1 truncate text-sm text-muted" title={item.url}>
                        {item.url}
                      </p>
                    </div>
                  </div>

                  <div className="flex min-w-0 flex-wrap gap-2">
                    <Badge tone="warning">{item.confidenceScore}% confidence</Badge>
                    {item.serviceType ? <Badge>{item.serviceType}</Badge> : null}
                    {(item.sourceTypes || []).map((source) => (
                      <Badge key={source} tone={sourceTone(source)}>
                        {source}
                      </Badge>
                    ))}
                  </div>

                  <div className="grid min-w-0 gap-1 text-sm">
                    <p className="truncate font-medium text-ink">
                      {item.host}:{item.port}
                    </p>
                    <p className="truncate text-muted">
                      {item.deviceName || "Unlinked device"} · {item.state}
                    </p>
                  </div>

                  <div className="flex flex-wrap gap-2 xl:justify-end">
                    <Button
                      onClick={() => window.open(item.url, "_blank", "noopener,noreferrer")}
                      size="sm"
                      variant="secondary"
                    >
                      Open
                    </Button>
                    {item.state === "ignored" ? (
                      <Button
                        disabled={!canManage}
                        leadingIcon={RefreshIcon}
                        onClick={() => onRestore(item)}
                        size="sm"
                        variant="ghost"
                      >
                        Restore
                      </Button>
                    ) : (
                      <>
                        <Button
                          disabled={!canManage}
                          leadingIcon={BookmarkIcon}
                          onClick={() => onCreateBookmark(item)}
                          size="sm"
                          variant="ghost"
                        >
                          Create bookmark
                        </Button>
                        <Button
                          disabled={!canManage}
                          onClick={() => onIgnore(item)}
                          size="sm"
                          variant="ghost"
                        >
                          Ignore
                        </Button>
                      </>
                    )}
                  </div>
                </article>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
