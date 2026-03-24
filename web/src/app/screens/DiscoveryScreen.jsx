import { useState } from "react";

import DiscoverySection from "../../components/dashboard/DiscoverySection";
import BookmarkSuggestionDialog from "../../components/discovery/BookmarkSuggestionDialog";
import DiscoveredServicesPanel from "../../components/discovery/DiscoveredServicesPanel";
import DockerEndpointForm from "../../components/forms/DockerEndpointForm";
import ScanTargetForm from "../../components/forms/ScanTargetForm";
import Modal from "../../components/ui/Modal";

function modalConfig(activeModal) {
  switch (activeModal) {
    case "endpoint":
      return {
        description:
          "Connect a local or remote Docker engine so containers show up as first-class inventory.",
        title: "Add Docker endpoint",
      };
    case "target":
      return {
        description:
          "Register a CIDR range and port profile for ongoing network discovery.",
        title: "Add scan target",
      };
    default:
      return { description: "", title: "" };
  }
}

export default function DiscoveryScreen({
  canManageUI,
  dashboard,
  folders = [],
  onIgnoreDiscoveredService,
  onRestoreDiscoveredService,
  onSaveBookmarkFromDiscoveredService,
  onSaveDiscoveryPolicy,
  onSaveDockerEndpoint,
  onSaveScanTarget,
  settings,
}) {
  const [activeModal, setActiveModal] = useState("");
  const [selectedDiscoveredService, setSelectedDiscoveredService] = useState(null);
  const pendingDiscoveredServices = (dashboard?.discoveredServices ?? []).filter(
    (item) => item.state === "pending" || item.state === "ignored",
  );
  const currentModal = modalConfig(activeModal);

  async function handleSaveDockerEndpoint(payload) {
    const successful = await onSaveDockerEndpoint(payload);
    if (successful) {
      setActiveModal("");
    }
    return successful;
  }

  async function handleSaveScanTarget(payload) {
    const successful = await onSaveScanTarget(payload);
    if (successful) {
      setActiveModal("");
    }
    return successful;
  }

  async function handleCreateBookmark(item, payload) {
    const successful = await onSaveBookmarkFromDiscoveredService(item.id, payload);
    if (successful) {
      setSelectedDiscoveredService(null);
    }
    return successful;
  }

  return (
    <>
      <DiscoverySection
        canManage={canManageUI}
        discoverySettings={settings?.discovery}
        dockerEndpoints={settings?.dockerEndpoints ?? []}
        onAddDockerEndpoint={() => setActiveModal("endpoint")}
        onAddScanTarget={() => setActiveModal("target")}
        onSaveSettings={onSaveDiscoveryPolicy}
        scanTargets={settings?.scanTargets ?? []}
      />
      <DiscoveredServicesPanel
        canManage={canManageUI}
        items={pendingDiscoveredServices}
        onCreateBookmark={(item) => setSelectedDiscoveredService(item)}
        onIgnore={(item) => void onIgnoreDiscoveredService(item.id)}
        onRestore={(item) => void onRestoreDiscoveredService(item.id)}
      />
      <Modal
        description={currentModal.description}
        onClose={() => setActiveModal("")}
        open={Boolean(activeModal)}
        title={currentModal.title}
      >
        {activeModal === "endpoint" ? (
          <DockerEndpointForm onSubmit={handleSaveDockerEndpoint} />
        ) : null}
        {activeModal === "target" ? (
          <ScanTargetForm onSubmit={handleSaveScanTarget} />
        ) : null}
      </Modal>
      <BookmarkSuggestionDialog
        folders={folders}
        item={selectedDiscoveredService}
        onClose={() => setSelectedDiscoveredService(null)}
        onSubmit={handleCreateBookmark}
        open={Boolean(selectedDiscoveredService)}
      />
    </>
  );
}
