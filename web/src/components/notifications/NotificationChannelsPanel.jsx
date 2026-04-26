import { useState } from "react";

import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import NotificationChannelForm from "./NotificationChannelForm";

export default function NotificationChannelsPanel({
  canManage,
  channels,
  onDelete,
  onSave,
  onTest,
}) {
  const [editing, setEditing] = useState(null);
  return (
    <Card>
      <CardHeader
        description="Webhook URLs are redacted in read responses. Blank or redacted secret fields keep the stored value."
        title="Channels"
      />
      <CardContent className="grid gap-5">
        {canManage ? (
          <NotificationChannelForm
            channel={editing}
            onCancel={() => setEditing(null)}
            onSubmit={async (payload) => {
              const ok = await onSave(payload);
              if (ok) setEditing(null);
            }}
          />
        ) : null}
        <div className="grid gap-3">
          {channels.map((channel) => (
            <div className="rounded-2xl border border-line bg-base p-4" key={channel.id}>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold text-ink">{channel.name}</h3>
                    <Badge tone={channel.enabled ? "success" : "neutral"}>{channel.enabled ? "Enabled" : "Disabled"}</Badge>
                    <Badge>{channel.type}</Badge>
                  </div>
                  <p className="mt-1 text-sm text-muted">{channel.type === "webhook" ? channel.config?.url : `${channel.config?.serverUrl || ""}/${channel.config?.topic || ""}`}</p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button disabled={!canManage} onClick={() => void onTest(channel.id)} size="sm" variant="secondary">Test</Button>
                  <Button disabled={!canManage} onClick={() => setEditing(channel)} size="sm" variant="ghost">Edit</Button>
                  <Button disabled={!canManage} onClick={() => void onDelete(channel.id)} size="sm" variant="ghost">Delete</Button>
                </div>
              </div>
            </div>
          ))}
          {channels.length === 0 ? <p className="text-sm text-muted">No notification channels configured.</p> : null}
        </div>
      </CardContent>
    </Card>
  );
}
