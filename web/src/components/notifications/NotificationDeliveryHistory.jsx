import Badge from "../ui/Badge";
import { Card, CardContent, CardHeader } from "../ui/Card";

export default function NotificationDeliveryHistory({ deliveries }) {
  return (
    <Card>
      <CardHeader title="Delivery History" description="Latest notification attempts, including tests and failed sends." />
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[720px] text-left text-sm">
            <thead className="text-xs uppercase tracking-[0.16em] text-muted">
              <tr>
                <th className="py-2 pr-4">Status</th>
                <th className="py-2 pr-4">Event</th>
                <th className="py-2 pr-4">Channel</th>
                <th className="py-2 pr-4">Rule</th>
                <th className="py-2 pr-4">Message</th>
                <th className="py-2">Attempted</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {deliveries.map((delivery) => (
                <tr key={delivery.id}>
                  <td className="py-3 pr-4"><Badge tone={delivery.status === "sent" ? "success" : delivery.status === "failed" ? "danger" : "neutral"}>{delivery.status}</Badge></td>
                  <td className="py-3 pr-4 text-ink-soft">{delivery.eventType}</td>
                  <td className="py-3 pr-4 text-ink-soft">{delivery.channelName || delivery.channelId || "Deleted channel"}</td>
                  <td className="py-3 pr-4 text-ink-soft">{delivery.ruleName || delivery.ruleId || "Test send"}</td>
                  <td className="py-3 pr-4 text-muted">{delivery.message}</td>
                  <td className="py-3 text-muted">{delivery.attemptedAt ? new Date(delivery.attemptedAt).toLocaleString() : ""}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {deliveries.length === 0 ? <p className="mt-4 text-sm text-muted">No delivery attempts recorded.</p> : null}
      </CardContent>
    </Card>
  );
}
