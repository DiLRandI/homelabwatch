import NotificationChannelsPanel from "../../components/notifications/NotificationChannelsPanel";
import NotificationDeliveryHistory from "../../components/notifications/NotificationDeliveryHistory";
import NotificationRulesPanel from "../../components/notifications/NotificationRulesPanel";

export default function NotificationsScreen({
  canManageUI,
  notifications,
  onDeleteChannel,
  onDeleteRule,
  onSaveChannel,
  onSaveRule,
  onTestChannel,
}) {
  return (
    <div className="grid gap-6">
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <NotificationChannelsPanel
          canManage={canManageUI}
          channels={notifications?.channels ?? []}
          onDelete={onDeleteChannel}
          onSave={onSaveChannel}
          onTest={onTestChannel}
        />
        <NotificationRulesPanel
          canManage={canManageUI}
          channels={notifications?.channels ?? []}
          onDelete={onDeleteRule}
          onSave={onSaveRule}
          rules={notifications?.rules ?? []}
        />
      </div>
      <NotificationDeliveryHistory deliveries={notifications?.deliveries ?? []} />
    </div>
  );
}
