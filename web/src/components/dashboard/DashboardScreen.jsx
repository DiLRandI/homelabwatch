import { useState } from "react";

import DashboardLayout from "../layout/DashboardLayout";
import Alerts from "../ui/Alerts";
import Button from "../ui/Button";
import DropdownMenu from "../ui/DropdownMenu";
import Input from "../ui/Input";
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
import BookmarkForm from "../forms/BookmarkForm";
import DockerEndpointForm from "../forms/DockerEndpointForm";
import ManualServiceForm from "../forms/ManualServiceForm";
import ScanTargetForm from "../forms/ScanTargetForm";
import BookmarksSection from "./BookmarksSection";
import DashboardHeader from "./DashboardHeader";
import DevicesSection from "./DevicesSection";
import DiscoverySection from "./DiscoverySection";
import ServicesSection from "./ServicesSection";
import WorkersSection from "./WorkersSection";

const DEFAULT_SUMMARY = {
  totalServices: 0,
  healthyServices: 0,
  degradedServices: 0,
  unhealthyServices: 0,
  devicesSeen: 0,
  bookmarks: 0,
};

function modalConfig(activeModal) {
  switch (activeModal) {
    case "service":
      return {
        description: "Capture a stable URL that should be monitored alongside discovered infrastructure.",
        title: "Add manual service",
      };
    case "bookmark":
      return {
        description: "Save a frequently used external dashboard or reference link for the team.",
        title: "Add bookmark",
      };
    case "endpoint":
      return {
        description: "Connect a local or remote Docker engine so containers show up as first-class services.",
        title: "Add Docker endpoint",
      };
    case "target":
      return {
        description: "Register a CIDR range and port profile for ongoing network discovery.",
        title: "Add scan target",
      };
    default:
      return { description: "", title: "" };
  }
}

export default function DashboardScreen({
  adminToken,
  dashboard,
  error,
  notice,
  onAdminTokenChange,
  onRefresh,
  onRunDiscovery,
  onRunMonitoring,
  onSaveBookmark,
  onSaveDockerEndpoint,
  onSaveManualService,
  onSaveScanTarget,
  settings,
}) {
  const [activeModal, setActiveModal] = useState("");
  const summary = dashboard?.summary ?? DEFAULT_SUMMARY;
  const metrics = [
    {
      description: "Total tracked service endpoints across all sources.",
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
      description: "Detected issues that need operator attention.",
      icon: SparklesIcon,
      iconTone: "bg-warn/10 text-warn-strong",
      label: "Degraded",
      value: summary.degradedServices + summary.unhealthyServices,
    },
    {
      description: "Devices known to the control plane inventory.",
      icon: DevicesIcon,
      iconTone: "bg-slate-100 text-slate-700",
      label: "Devices",
      value: summary.devicesSeen,
    },
  ];

  const navItems = [
    { count: null, href: "#overview", icon: OverviewIcon, label: "Overview" },
    {
      count: summary.totalServices,
      href: "#services",
      icon: ServicesIcon,
      label: "Services",
    },
    {
      count: settings?.dockerEndpoints?.length ?? 0,
      href: "#discovery",
      icon: DiscoveryIcon,
      label: "Discovery",
    },
    {
      count: summary.devicesSeen,
      href: "#devices",
      icon: DevicesIcon,
      label: "Devices",
    },
    {
      count: summary.bookmarks,
      href: "#bookmarks",
      icon: BookmarkIcon,
      label: "Bookmarks",
    },
    {
      count: dashboard?.recentEvents?.length ?? 0,
      href: "#activity",
      icon: ActivityIcon,
      label: "Activity",
    },
  ];

  const statusItems = [
    {
      className: "border-slate-200 bg-white text-slate-600",
      icon: SparklesIcon,
      label: "Realtime updates",
    },
    {
      className: "border-sky-200 bg-sky-50 text-sky-700",
      icon: DiscoveryIcon,
      label: `${settings?.dockerEndpoints?.length ?? 0} Docker endpoints`,
    },
    {
      className: "border-slate-200 bg-white text-slate-600",
      icon: ClockIcon,
      label: settings?.appSettings?.autoScanEnabled ? "Auto scan enabled" : "Auto scan off",
    },
  ];

  async function submitAndClose(action, payload) {
    const successful = await action(payload);
    if (successful) {
      setActiveModal("");
    }
    return successful;
  }

  const quickActions = [
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
      onSelect: () => setActiveModal("bookmark"),
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
      description: "Run a fresh discovery sweep now.",
      icon: RefreshIcon,
      label: "Run discovery",
      onSelect: () => void onRunDiscovery(),
    },
    {
      description: "Execute health checks immediately.",
      icon: ShieldIcon,
      label: "Run checks",
      onSelect: () => void onRunMonitoring(),
    },
  ];

  const currentModal = modalConfig(activeModal);

  return (
    <>
      <DashboardLayout
        alerts={<Alerts error={error} notice={notice} />}
        metrics={metrics}
        navItems={navItems}
        statusItems={statusItems}
        subtitle="A clean operator view for discovery, monitoring, and day-two homelab workflows."
        title="Operations"
        tokenFile={settings?.appSettings?.adminTokenFile ?? ""}
        toolbar={
          <div className="grid gap-3 xl:grid-cols-[minmax(0,220px)_repeat(3,auto)_auto]">
            <Input
              autoComplete="off"
              compact
              containerClassName="min-w-0"
              inputClassName="bg-white"
              label="Admin token"
              onChange={onAdminTokenChange}
              placeholder="Paste write token"
              type="password"
              value={adminToken}
            />
            <Button
              leadingIcon={DiscoveryIcon}
              onClick={() => void onRunDiscovery()}
              variant="secondary"
            >
              Run discovery
            </Button>
            <Button
              leadingIcon={ShieldIcon}
              onClick={() => void onRunMonitoring()}
              variant="secondary"
            >
              Run checks
            </Button>
            <Button
              leadingIcon={RefreshIcon}
              onClick={() => void onRefresh()}
              variant="ghost"
            >
              Refresh
            </Button>
            <DropdownMenu
              items={quickActions}
              label="Quick actions"
              leadingIcon={PlusIcon}
            />
          </div>
        }
      >
        <DashboardHeader
          adminTokenFile={settings?.appSettings?.adminTokenFile ?? ""}
          metrics={metrics}
          onOpenModal={setActiveModal}
          settings={settings}
        />
        <ServicesSection
          onAdd={() => setActiveModal("service")}
          services={dashboard?.services ?? []}
        />
        <DiscoverySection
          dockerEndpoints={settings?.dockerEndpoints ?? []}
          onAddDockerEndpoint={() => setActiveModal("endpoint")}
          onAddScanTarget={() => setActiveModal("target")}
          scanTargets={settings?.scanTargets ?? []}
        />
        <DevicesSection devices={dashboard?.devices ?? []} />
        <BookmarksSection
          bookmarks={dashboard?.bookmarks ?? []}
          onAdd={() => setActiveModal("bookmark")}
        />
        <WorkersSection
          jobState={settings?.jobState ?? []}
          recentEvents={dashboard?.recentEvents ?? []}
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
        {activeModal === "bookmark" ? (
          <BookmarkForm
            onSubmit={(payload) => submitAndClose(onSaveBookmark, payload)}
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
      </Modal>
    </>
  );
}
