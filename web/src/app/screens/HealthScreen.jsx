import { useMemo, useState } from "react";

import ServicesSection from "../../components/dashboard/ServicesSection";
import ManualServiceForm from "../../components/forms/ManualServiceForm";
import Modal from "../../components/ui/Modal";

export default function HealthScreen({
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
  const monitoredServices = useMemo(() => {
    const statusRank = {
      unhealthy: 0,
      degraded: 1,
      unknown: 2,
      healthy: 3,
    };

    return (dashboard?.services ?? [])
      .filter(
        (service) =>
          (service.checks?.length ?? 0) > 0 ||
          service.status === "healthy" ||
          service.status === "degraded" ||
          service.status === "unhealthy",
      )
      .sort((left, right) => {
        const statusDelta =
          (statusRank[left.status] ?? 99) - (statusRank[right.status] ?? 99);
        if (statusDelta !== 0) {
          return statusDelta;
        }
        return left.name.localeCompare(right.name);
      });
  }, [dashboard?.services]);

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
        addLabel="Add monitored service"
        bookmarkedServiceIds={bookmarkedServiceIds}
        canManage={canManageUI}
        description="Managed checks and endpoint health across accepted services."
        emptyBody="Create a service or accept a discovery to start monitoring it."
        emptyTitle="No monitored services yet"
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
        sectionId="health"
        services={monitoredServices}
        title="Monitoring and health"
      />
      <Modal
        description="Capture the open URL and, optionally, a different health target for monitoring."
        onClose={() => setCreatingService(false)}
        open={creatingService}
        title="Add monitored service"
      >
        <ManualServiceForm onSubmit={handleSaveService} submitLabel="Create service" />
      </Modal>
    </>
  );
}
