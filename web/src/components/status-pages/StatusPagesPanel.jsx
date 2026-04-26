import { formatDate } from "../../lib/format";
import HealthStatusBadge from "../health/HealthStatusBadge";
import Button from "../ui/Button";
import { Card } from "../ui/Card";
import Badge from "../ui/Badge";
import { PlusIcon } from "../ui/Icons";

export default function StatusPagesPanel({ items = [], onCreate, onSelect, selectedId }) {
  return (
    <Card className="p-5">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-ink">Pages</h2>
          <p className="text-sm text-muted">Public, read-only status views.</p>
        </div>
        <Button leadingIcon={PlusIcon} onClick={onCreate} size="sm">Create</Button>
      </div>
      <div className="mt-5 grid gap-3">
        {items.length === 0 ? (
          <p className="rounded-2xl border border-dashed border-line p-4 text-sm text-muted">No status pages yet.</p>
        ) : items.map((item) => (
          <button
            className={`grid gap-2 rounded-2xl border p-4 text-left transition ${
              item.id === selectedId ? "border-accent bg-accent/5" : "border-line bg-panel-strong hover:border-line-strong"
            }`}
            key={item.id}
            onClick={() => onSelect(item.id)}
            type="button"
          >
            <div className="flex items-start justify-between gap-3">
              <div>
                <div className="font-medium text-ink">{item.title}</div>
                <div className="text-sm text-muted">/status/{item.slug}</div>
              </div>
              <Badge>{item.enabled ? "Enabled" : "Disabled"}</Badge>
            </div>
            <HealthStatusBadge status={item.overallStatus} subtle />
            <div className="text-xs text-muted">
              {item.serviceCount} services, {item.announcementCount} announcements - updated {formatDate(item.updatedAt)}
            </div>
          </button>
        ))}
      </div>
    </Card>
  );
}
