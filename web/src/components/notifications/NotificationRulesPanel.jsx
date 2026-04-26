import { useState } from "react";

import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import NotificationRuleForm from "./NotificationRuleForm";

export default function NotificationRulesPanel({ canManage, channels, onDelete, onSave, rules }) {
  const [editing, setEditing] = useState(null);
  return (
    <Card>
      <CardHeader title="Rules" description="Routes supported event types to one or more enabled channels." />
      <CardContent className="grid gap-5">
        {canManage ? (
          <NotificationRuleForm
            channels={channels}
            onCancel={() => setEditing(null)}
            onSubmit={async (payload) => {
              const ok = await onSave(payload);
              if (ok) setEditing(null);
            }}
            rule={editing}
          />
        ) : null}
        <div className="grid gap-3">
          {rules.map((rule) => (
            <div className="rounded-2xl border border-line bg-base p-4" key={rule.id}>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="font-semibold text-ink">{rule.name}</h3>
                    <Badge tone={rule.enabled ? "success" : "neutral"}>{rule.enabled ? "Enabled" : "Disabled"}</Badge>
                    <Badge>{rule.eventType}</Badge>
                  </div>
                  <p className="mt-1 text-sm text-muted">{(rule.channels || []).map((channel) => channel.name).join(", ") || "No channels"}</p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button disabled={!canManage} onClick={() => setEditing(rule)} size="sm" variant="ghost">Edit</Button>
                  <Button disabled={!canManage} onClick={() => void onDelete(rule.id)} size="sm" variant="ghost">Delete</Button>
                </div>
              </div>
            </div>
          ))}
          {rules.length === 0 ? <p className="text-sm text-muted">No notification rules configured.</p> : null}
        </div>
      </CardContent>
    </Card>
  );
}
