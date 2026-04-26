import StatusPageAnnouncementsPanel from "../../components/status-pages/StatusPageAnnouncementsPanel";
import StatusPageEditor from "../../components/status-pages/StatusPageEditor";
import StatusPagePreview from "../../components/status-pages/StatusPagePreview";
import StatusPageServicePicker from "../../components/status-pages/StatusPageServicePicker";
import StatusPagesPanel from "../../components/status-pages/StatusPagesPanel";

export default function StatusPagesScreen({
  canManageUI,
  dashboard,
  onCreatePage,
  onDeleteAnnouncement,
  onDeletePage,
  onSaveAnnouncement,
  onSavePage,
  onSaveServices,
  onSelectPage,
  statusPages,
}) {
  const selected = statusPages?.selected || null;
  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(320px,0.38fr)_minmax(0,1fr)]">
      <StatusPagesPanel
        items={statusPages?.list || []}
        onCreate={onCreatePage}
        onSelect={onSelectPage}
        selectedId={selected?.id}
      />
      <div className="grid gap-6">
        <StatusPageEditor
          canManage={canManageUI}
          onDelete={onDeletePage}
          onOpenPublic={() => selected?.slug && window.open(`/status/${selected.slug}`, "_blank", "noopener,noreferrer")}
          onSave={onSavePage}
          page={selected}
        />
        <StatusPageServicePicker
          canManage={canManageUI}
          onSave={onSaveServices}
          page={selected}
          services={dashboard?.services || []}
        />
        <StatusPageAnnouncementsPanel
          canManage={canManageUI}
          onDelete={onDeleteAnnouncement}
          onSave={onSaveAnnouncement}
          page={selected}
        />
        <StatusPagePreview page={selected} />
      </div>
    </div>
  );
}
