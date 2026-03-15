import { useState } from "react";

import DashboardLayout from "../layout/DashboardLayout";
import Alerts from "../ui/Alerts";
import Button from "../ui/Button";
import DropdownMenu from "../ui/DropdownMenu";
import Modal from "../ui/Modal";
import {
  ActivityIcon,
  BookmarkIcon,
  ClockIcon,
  DevicesIcon,
  DiscoveryIcon,
  OverviewIcon,
  PlusIcon,
  RefreshIcon,
  ServicesIcon,
  ShieldIcon,
  SparklesIcon,
  TokenIcon,
} from "../ui/Icons";
import APITokenForm from "../forms/APITokenForm";
import DockerEndpointForm from "../forms/DockerEndpointForm";
import ManualServiceForm from "../forms/ManualServiceForm";
import ScanTargetForm from "../forms/ScanTargetForm";
import ApiAccessSection from "./ApiAccessSection";
import BookmarksHome from "../bookmarks/BookmarksHome";
import ContainersSection from "./ContainersSection";
import DashboardHeader from "./DashboardHeader";
import DevicesSection from "./DevicesSection";
import DiscoverySection from "./DiscoverySection";
import DiscoveredServicesPanel from "../discovery/DiscoveredServicesPanel";
import BookmarkSuggestionDialog from "../discovery/BookmarkSuggestionDialog";
import ServicesSection from "./ServicesSection";
import WorkersSection from "./WorkersSection";

const DEFAULT_SUMMARY = {
  bookmarks: 0,
  degradedServices: 0,
  devicesSeen: 0,
  healthyServices: 0,
  runningContainers: 0,
  totalServices: 0,
  unhealthyServices: 0,
  discoveredServices: 0,
};

function modalConfig(activeModal) {
  switch (activeModal) {
    case "service":
      return {
        description:
          "Capture a stable URL that should be monitored alongside discovered infrastructure.",
        title: "Add manual service",
      };
    case "bookmark":
      return {
        description:
          "Save a frequently used external dashboard or reference link for the team.",
        title: "Add bookmark",
      };
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
    case "apiToken":
      return {
        description:
          "Create a scoped bearer token for scripts, integrations, or external dashboards.",
        title: "Create external API token",
      };
    default:
      return { description: "", title: "" };
  }
}

export default function DashboardScreen({
  bookmarks,
  canManageUI,
  dashboard,
  error,
  folders,
  notice,
  onCreateAPIToken,
  onDeleteBookmark,
  onDeleteFolder,
  onExportBookmarks,
  onIgnoreDiscoveredService,
  onImportBookmarks,
  onRefresh,
  onReorderBookmarks,
  onReorderFolders,
  onRevokeAPIToken,
  onRunDiscovery,
  onRunMonitoring,
  onSaveBookmark,
  onSaveBookmarkFromDiscoveredService,
  onSaveBookmarkFromService,
  onSaveDiscoveryPolicy,
  onSaveDockerEndpoint,
  onSaveFolder,
  onSaveManualService,
  onSaveScanTarget,
  onRestoreDiscoveredService,
  settings,
  tags,
  onUploadBookmarkIcon,
}) {
  const [activeModal, setActiveModal] = useState("");
  const [createdToken, setCreatedToken] = useState(null);
  const [bookmarkComposerToken, setBookmarkComposerToken] = useState(0);
  const [selectedDiscoveredService, setSelectedDiscoveredService] = useState(null);
  const summary = dashboard?.summary ?? DEFAULT_SUMMARY;
  const pendingDiscoveredServices = (dashboard?.discoveredServices ?? []).filter(
    (item) => item.state === "pending" || item.state === "ignored",
  );
  const issuesCount = summary.degradedServices + summary.unhealthyServices;
  const serviceCounts = (dashboard?.services ?? []).reduce((items, service) => {
    if (!service.deviceId) {
      return items;
    }
    return {
      ...items,
      [service.deviceId]: (items[service.deviceId] || 0) + 1,
    };
  }, {});
  const discoveryCounts = (dashboard?.discoveredServices ?? []).reduce((items, service) => {
    if (!service.deviceId) {
      return items;
    }
    return {
      ...items,
      [service.deviceId]: (items[service.deviceId] || 0) + 1,
    };
  }, {});
  const metrics = [
    {
      description: "Tracked endpoints across all discovery sources.",
      icon: ServicesIcon,
      iconTone: "bg-accent/10 text-accent-strong",
      label: "Services",
      value: summary.totalServices,
    },
    {
      description: "Passing checks and responding to requests.",
      icon: ShieldIcon,
      iconTone: "bg-ok/10 text-ok-strong",
      label: "Healthy",
      value: summary.healthyServices,
    },
    {
      description: "Running workloads discovered from attached Docker engines.",
      icon: DiscoveryIcon,
      iconTone: "bg-sky-50 text-sky-700",
      label: "Containers",
      value: summary.runningContainers,
    },
    {
      description: "Devices known to the control plane inventory.",
      icon: DevicesIcon,
      iconTone: "bg-slate-100 text-slate-700",
      label: "Devices",
      value: summary.devicesSeen,
    },
    {
      description: "Pending discovery suggestions waiting for review.",
      icon: SparklesIcon,
      iconTone: "bg-amber-100 text-amber-700",
      label: "Discovered",
      value: summary.discoveredServices,
    },
  ];

  const navItems = [
    { count: null, href: "#overview", icon: OverviewIcon, label: "Overview" },
    {
      count: bookmarks?.length ?? 0,
      href: "#bookmarks-home",
      icon: BookmarkIcon,
      label: "Bookmarks",
    },
    {
      count: summary.totalServices,
      href: "#services",
      icon: ServicesIcon,
      label: "Services",
    },
    {
      count: summary.runningContainers,
      href: "#containers",
      icon: DiscoveryIcon,
      label: "Containers",
    },
    {
      count: settings?.dockerEndpoints?.length ?? 0,
      href: "#discovery",
      icon: DiscoveryIcon,
      label: "Discovery",
    },
    {
      count: summary.discoveredServices,
      href: "#discovered-services",
      icon: SparklesIcon,
      label: "Suggestions",
    },
    {
      count: summary.devicesSeen,
      href: "#devices",
      icon: DevicesIcon,
      label: "Devices",
    },
    {
      count: dashboard?.recentEvents?.length ?? 0,
      href: "#activity",
      icon: ActivityIcon,
      label: "Activity",
    },
    {
      count: settings?.apiAccess?.tokens?.length ?? 0,
      href: "#settings",
      icon: TokenIcon,
      label: "Settings",
    },
  ];

  const statusItems = [
    {
      className: "border-slate-200 bg-white text-slate-600",
      icon: SparklesIcon,
      label: "Realtime updates",
    },
    {
      className:
        issuesCount > 0
          ? "border-warn/20 bg-warn/10 text-warn-strong"
          : "border-ok/15 bg-ok/10 text-ok-strong",
      icon: ShieldIcon,
      label:
        issuesCount > 0
          ? `${issuesCount} services need attention`
          : "Service health clear",
    },
    {
      className: canManageUI
        ? "border-accent/15 bg-accent/10 text-accent-strong"
        : "border-slate-200 bg-white text-slate-600",
      icon: ClockIcon,
      label: canManageUI ? "Trusted LAN writes enabled" : "Read-only network",
    },
  ];

  async function submitAndClose(action, payload) {
    const successful = await action(payload);
    if (successful) {
      setActiveModal("");
    }
    return successful;
  }

  async function createTokenAndClose(payload) {
    const created = await onCreateAPIToken(payload);
    if (!created) {
      return false;
    }
    setCreatedToken(created);
    setActiveModal("");
    return true;
  }

  async function submitDiscoveredBookmark(item, payload) {
    return onSaveBookmarkFromDiscoveredService(item.id, payload);
  }

  const quickActions = canManageUI
    ? [
        {
          description: "Create a manual service entry.",
          icon: ServicesIcon,
          label: "Add service",
          onSelect: () => setActiveModal("service"),
        },
        {
          description: "Save a dashboard or docs link.",
          icon: BookmarkIcon,
          label: "Add bookmark",
          onSelect: () => setBookmarkComposerToken((current) => current + 1),
        },
        {
          description: "Connect another Docker engine.",
          icon: DiscoveryIcon,
          label: "Add Docker endpoint",
          onSelect: () => setActiveModal("endpoint"),
        },
        {
          description: "Register a new scan target.",
          icon: DevicesIcon,
          label: "Add scan target",
          onSelect: () => setActiveModal("target"),
        },
        {
          description: "Create a token for external automation.",
          icon: TokenIcon,
          label: "Create API token",
          onSelect: () => setActiveModal("apiToken"),
        },
      ]
    : [];

  const currentModal = modalConfig(activeModal);

  return (
    <>
      <DashboardLayout
        alerts={<Alerts error={error} notice={notice} />}
        metrics={metrics}
        navItems={navItems}
        sidebarMeta={{
          apiTokenCount: settings?.apiAccess?.tokens?.length ?? 0,
          applianceName: settings?.appSettings?.applianceName,
          trustedNetwork: canManageUI,
        }}
        statusItems={statusItems}
        subtitle="Bookmarks lead the workspace, while discovery, health, and automation stay one scroll away."
        title="Navigation"
        toolbar={
          <div className="grid gap-3 xl:grid-cols-[repeat(3,auto)_auto]">
            <Button
              disabled={!canManageUI}
              leadingIcon={DiscoveryIcon}
              onClick={() => void onRunDiscovery()}
              variant="secondary"
            >
              Run discovery
            </Button>
            <Button
              disabled={!canManageUI}
              leadingIcon={ShieldIcon}
              onClick={() => void onRunMonitoring()}
              variant="secondary"
            >
              Run checks
            </Button>
            <Button leadingIcon={RefreshIcon} onClick={() => void onRefresh()} variant="ghost">
              Refresh
            </Button>
            {quickActions.length > 0 ? (
              <DropdownMenu items={quickActions} label="Quick actions" leadingIcon={PlusIcon} />
            ) : null}
          </div>
        }
      >
        <BookmarksHome
          bookmarks={bookmarks}
          canManage={canManageUI}
          devices={dashboard?.devices ?? []}
          folders={folders}
          onDeleteBookmark={onDeleteBookmark}
          onDeleteFolder={onDeleteFolder}
          onExportBookmarks={onExportBookmarks}
          onImportBookmarks={onImportBookmarks}
          onReorderBookmarks={onReorderBookmarks}
          onReorderFolders={onReorderFolders}
          onSaveBookmark={onSaveBookmark}
          onSaveFolder={onSaveFolder}
          onUploadBookmarkIcon={onUploadBookmarkIcon}
          openBookmarkComposerToken={bookmarkComposerToken}
          services={dashboard?.services ?? []}
          tags={tags}
        />
        <DashboardHeader
          canManageUI={canManageUI}
          metrics={metrics}
          onOpenModal={(modal) => {
            if (modal === "bookmark") {
              setBookmarkComposerToken((current) => current + 1);
              return;
            }
            setActiveModal(modal);
          }}
          settings={settings}
        />
        <DiscoveredServicesPanel
          canManage={canManageUI}
          items={pendingDiscoveredServices}
          onCreateBookmark={(item) => setSelectedDiscoveredService(item)}
          onIgnore={(item) => void onIgnoreDiscoveredService(item.id)}
          onRestore={(item) => void onRestoreDiscoveredService(item.id)}
        />
        <ServicesSection
          bookmarkedServiceIds={new Set((bookmarks ?? []).map((bookmark) => bookmark.serviceId).filter(Boolean))}
          canManage={canManageUI}
          onAdd={() => setActiveModal("service")}
          onAddBookmark={(service) =>
            void onSaveBookmarkFromService({
              isFavorite: false,
              serviceId: service.id,
            })
          }
          services={dashboard?.services ?? []}
        />
        <ContainersSection containers={dashboard?.containers ?? []} />
        <DiscoverySection
          canManage={canManageUI}
          discoverySettings={settings?.discovery}
          dockerEndpoints={settings?.dockerEndpoints ?? []}
          onAddDockerEndpoint={() => setActiveModal("endpoint")}
          onAddScanTarget={() => setActiveModal("target")}
          onSaveSettings={onSaveDiscoveryPolicy}
          scanTargets={settings?.scanTargets ?? []}
        />
        <DevicesSection
          devices={dashboard?.devices ?? []}
          discoveryCounts={discoveryCounts}
          serviceCounts={serviceCounts}
        />
        <WorkersSection
          jobState={settings?.jobState ?? []}
          recentEvents={dashboard?.recentEvents ?? []}
        />
        <ApiAccessSection
          canManage={canManageUI}
          createdToken={createdToken}
          legacyTokenAlive={settings?.apiAccess?.legacyAdminTokenAlive ?? false}
          onCreate={() => setActiveModal("apiToken")}
          onDismissCreatedToken={() => setCreatedToken(null)}
          onRevoke={onRevokeAPIToken}
          tokens={settings?.apiAccess?.tokens ?? []}
        />
      </DashboardLayout>

      <Modal
        description={currentModal.description}
        onClose={() => setActiveModal("")}
        open={Boolean(activeModal)}
        title={currentModal.title}
      >
        {activeModal === "service" ? (
          <ManualServiceForm
            onSubmit={(payload) => submitAndClose(onSaveManualService, payload)}
          />
        ) : null}
        {activeModal === "endpoint" ? (
          <DockerEndpointForm
            onSubmit={(payload) => submitAndClose(onSaveDockerEndpoint, payload)}
          />
        ) : null}
        {activeModal === "target" ? (
          <ScanTargetForm
            onSubmit={(payload) => submitAndClose(onSaveScanTarget, payload)}
          />
        ) : null}
        {activeModal === "apiToken" ? (
          <APITokenForm onSubmit={createTokenAndClose} />
        ) : null}
      </Modal>

      <BookmarkSuggestionDialog
        folders={folders}
        item={selectedDiscoveredService}
        onClose={() => setSelectedDiscoveredService(null)}
        onSubmit={submitDiscoveredBookmark}
        open={Boolean(selectedDiscoveredService)}
      />
    </>
  );
}
