import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import StatusBadge from "../ui/StatusBadge";
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
  items = [],
  onCreateBookmark,
  onIgnore,
  onRestore,
}) {
  return (
    <section id="discovered-services">
      <Card>
        <CardHeader
          description="Pending suggestions merged across Docker, LAN, and mDNS evidence before they become first-class bookmarks."
          title="Discovered services"
        />
        <CardContent>
          {items.length === 0 ? (
            <EmptyState
              body="Run discovery to collect bookmark suggestions from containers and network services."
              title="No pending suggestions"
            />
          ) : (
            <div className="grid gap-4 xl:grid-cols-2">
              {items.map((item) => (
                <article
                  className="flex h-full flex-col rounded-3xl border border-slate-200 bg-slate-50 p-5"
                  key={item.id}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0">
                      <div className="flex items-center gap-3">
                        <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-100 text-amber-700">
                          <DiscoveryIcon className="h-5 w-5" />
                        </span>
                        <div className="min-w-0">
                          <h3 className="truncate text-lg font-semibold tracking-tight text-slate-950">
                            {item.name}
                          </h3>
                          <p className="truncate text-sm text-slate-500" title={item.url}>
                            {item.url}
                          </p>
                        </div>
                      </div>
                    </div>
                    <StatusBadge status={item.status} />
                  </div>

                  <div className="mt-4 flex flex-wrap gap-2">
                    <Badge tone="warning">{item.confidenceScore}% confidence</Badge>
                    {item.serviceType ? <Badge>{item.serviceType}</Badge> : null}
                    <Badge>{item.deviceName || item.host || "Unlinked device"}</Badge>
                    {(item.sourceTypes || []).map((source) => (
                      <Badge key={source} tone={sourceTone(source)}>
                        {source}
                      </Badge>
                    ))}
                  </div>

                  <dl className="mt-5 grid gap-3 text-sm text-slate-600 sm:grid-cols-2">
                    <div className="rounded-2xl border border-white bg-white px-4 py-3">
                      <dt className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">
                        Endpoint
                      </dt>
                      <dd className="mt-2 truncate font-medium text-slate-900">
                        {item.host}:{item.port}
                      </dd>
                    </div>
                    <div className="rounded-2xl border border-white bg-white px-4 py-3">
                      <dt className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">
                        State
                      </dt>
                      <dd className="mt-2 font-medium text-slate-900">{item.state}</dd>
                    </div>
                  </dl>

                  <div className="mt-5 flex flex-wrap gap-2">
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
