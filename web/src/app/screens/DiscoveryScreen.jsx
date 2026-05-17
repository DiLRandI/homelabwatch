import { useState } from "react";

import DiscoverySection from "../../components/dashboard/DiscoverySection";
import BookmarkSuggestionDialog from "../../components/discovery/BookmarkSuggestionDialog";
import DiscoveredServicesPanel from "../../components/discovery/DiscoveredServicesPanel";
import TopologySourcesPanel from "../../components/discovery/TopologySourcesPanel";
import DockerEndpointForm from "../../components/forms/DockerEndpointForm";
import ScanTargetForm from "../../components/forms/ScanTargetForm";
import TopologySourceForm from "../../components/forms/TopologySourceForm";
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
    case "topologySource":
      return {
        description:
          "Add an SNMP source for observed LLDP and switch-port topology.",
        title: "Topology source",
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
  onSaveTopologySource,
  onAutoDiscoverTopologySources,
  onDeleteTopologySource,
  onRunTopologyDiscovery,
  settings,
  topologySources = [],
}) {
  const [activeModal, setActiveModal] = useState("");
  const [selectedDiscoveredService, setSelectedDiscoveredService] = useState(null);
  const [selectedTopologySource, setSelectedTopologySource] = useState(null);
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

  async function handleSaveTopologySource(payload) {
    const successful = await onSaveTopologySource(payload);
    if (successful) {
      setActiveModal("");
      setSelectedTopologySource(null);
    }
    return successful;
  }

  function handleEditTopologySource(item) {
    setSelectedTopologySource(item);
    setActiveModal("topologySource");
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
      <TopologySourcesPanel
        canManage={canManageUI}
        items={topologySources}
        onAutoDiscover={onAutoDiscoverTopologySources}
        onAdd={() => {
          setSelectedTopologySource(null);
          setActiveModal("topologySource");
        }}
        onDelete={onDeleteTopologySource}
        onEdit={handleEditTopologySource}
        onRun={onRunTopologyDiscovery}
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
        onClose={() => {
          setActiveModal("");
          setSelectedTopologySource(null);
        }}
        open={Boolean(activeModal)}
        title={currentModal.title}
      >
        {activeModal === "endpoint" ? (
          <DockerEndpointForm onSubmit={handleSaveDockerEndpoint} />
        ) : null}
        {activeModal === "target" ? (
          <ScanTargetForm onSubmit={handleSaveScanTarget} />
        ) : null}
        {activeModal === "topologySource" ? (
          <TopologySourceForm
            item={selectedTopologySource}
            onSubmit={handleSaveTopologySource}
            submitLabel={selectedTopologySource ? "Save source" : "Add source"}
          />
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
