import { useState } from "react";

import ContainersSection from "../../components/dashboard/ContainersSection";
import ServicesSection from "../../components/dashboard/ServicesSection";
import ManualServiceForm from "../../components/forms/ManualServiceForm";
import Modal from "../../components/ui/Modal";

export default function ServicesScreen({
  bookmarks = [],
  canManageUI,
  dashboard,
  onDeleteServiceHealthCheck,
  onFetchServiceHealthChecks,
  onSaveBookmarkFromService,
  onSaveManualService,
  onSaveServiceHealthCheck,
  onTestServiceCheck,
}) {
  const [creatingService, setCreatingService] = useState(false);
  const bookmarkedServiceIds = new Set(
    bookmarks.map((bookmark) => bookmark.serviceId).filter(Boolean),
  );

  async function handleSaveService(payload) {
    const successful = await onSaveManualService(payload);
    if (successful) {
      setCreatingService(false);
    }
    return successful;
  }

  return (
    <>
      <ServicesSection
        bookmarkedServiceIds={bookmarkedServiceIds}
        canManage={canManageUI}
        onAdd={() => setCreatingService(true)}
        onAddBookmark={(service) =>
          void onSaveBookmarkFromService({
            isFavorite: false,
            serviceId: service.id,
          })
        }
        onDeleteHealthCheck={onDeleteServiceHealthCheck}
        onFetchHealthChecks={onFetchServiceHealthChecks}
        onSaveHealthCheck={onSaveServiceHealthCheck}
        onTestHealthCheck={onTestServiceCheck}
        services={dashboard?.services ?? []}
      />
      <ContainersSection containers={dashboard?.containers ?? []} />
      <Modal
        description="Capture a stable URL that should be monitored alongside discovered infrastructure."
        onClose={() => setCreatingService(false)}
        open={creatingService}
        title="Add manual service"
      >
        <ManualServiceForm onSubmit={handleSaveService} />
      </Modal>
    </>
  );
}
