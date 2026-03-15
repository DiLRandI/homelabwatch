import { formatDate } from "../../lib/format";
import EmptyState from "../ui/EmptyState";
import Section from "../ui/Section";

export default function DevicesSection({ devices }) {
  return (
    <Section
      title="Devices"
      subtitle="Tracked by MAC or fallback fingerprint to survive IP churn."
    >
      <div className="grid gap-3">
        {devices.length === 0 ? (
          <EmptyState
            title="No devices yet"
            body="LAN scans will populate this list."
            compact
          />
        ) : (
          devices.map((device) => (
            <div
              key={device.id}
              className="rounded-3xl border border-white/10 bg-base/70 p-4"
            >
              <div className="flex items-center justify-between gap-4">
                <div>
                  <h3 className="font-display text-lg font-semibold text-ink">
                    {device.displayName || device.hostname || device.identityKey}
                  </h3>
                  <p className="mt-1 text-sm text-muted">
                    {device.primaryMac || device.identityKey}
                  </p>
                </div>
                <span className="rounded-full border border-white/10 px-3 py-1 text-[11px] uppercase tracking-[0.22em] text-muted">
                  {device.identityConfidence}
                </span>
              </div>
              <div className="mt-3 grid gap-2 text-sm text-muted">
                <span>
                  IPs:{" "}
                  {device.addresses?.map((item) => item.ipAddress).join(", ") ||
                    "n/a"}
                </span>
                <span>
                  Ports:{" "}
                  {device.ports
                    ?.map((item) => `${item.port}/${item.protocol}`)
                    .join(", ") || "n/a"}
                </span>
                <span>Last seen: {formatDate(device.lastSeenAt)}</span>
              </div>
            </div>
          ))
        )}
      </div>
    </Section>
  );
}
