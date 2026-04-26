import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import PublicStatusPageScreen from "./PublicStatusPageScreen";
import StatusPageAnnouncementsPanel from "./StatusPageAnnouncementsPanel";
import StatusPageEditor from "./StatusPageEditor";
import StatusPageServicePicker from "./StatusPageServicePicker";

const publicPage = {
  slug: "home",
  title: "Home Status",
  description: "Public health",
  overallStatus: "degraded",
  lastUpdatedAt: "2026-04-26T10:00:00Z",
  announcements: [],
  services: [
    { name: "NAS", status: "healthy", lastCheckedAt: "2026-04-26T10:00:00Z" },
    { name: "Router", status: "unhealthy", lastCheckedAt: "2026-04-26T10:00:00Z" },
  ],
};

describe("status pages", () => {
  it("renders public services grouped by status order", () => {
    render(<PublicStatusPageScreen page={publicPage} />);

    const unhealthy = screen.getAllByText("unhealthy")[0];
    const healthy = screen.getAllByText("healthy")[0];

    expect(unhealthy.compareDocumentPosition(healthy) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(screen.getByText("Router")).toBeInTheDocument();
    expect(screen.getByText("NAS")).toBeInTheDocument();
  });

  it("renders public missing state", () => {
    render(<PublicStatusPageScreen missing />);

    expect(screen.getByText("Status page unavailable")).toBeInTheDocument();
  });

  it("submits editor create and update payloads", async () => {
    const onSave = vi.fn();
    const page = { id: "stp_1", slug: "old", title: "Old", description: "", enabled: true };
    render(<StatusPageEditor canManage onDelete={vi.fn()} onOpenPublic={vi.fn()} onSave={onSave} page={page} />);

    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "New Title" } });
    fireEvent.change(screen.getByLabelText("Slug"), { target: { value: "new-title" } });
    fireEvent.click(screen.getByText("Save"));

    expect(onSave).toHaveBeenCalledWith(expect.objectContaining({
      id: "stp_1",
      slug: "new-title",
      title: "New Title",
    }));
  });

  it("preserves service picker ordering payload", () => {
    const onSave = vi.fn();
    const page = {
      id: "stp_1",
      services: [
        { serviceId: "svc_1", serviceName: "One", displayName: "" },
        { serviceId: "svc_2", serviceName: "Two", displayName: "Second" },
      ],
    };
    render(<StatusPageServicePicker canManage onSave={onSave} page={page} services={[]} />);

    fireEvent.click(screen.getAllByText("Up")[1]);
    fireEvent.click(screen.getByText("Save order"));

    expect(onSave).toHaveBeenCalledWith("stp_1", [
      { serviceId: "svc_2", sortOrder: 0, displayName: "Second" },
      { serviceId: "svc_1", sortOrder: 1, displayName: "" },
    ]);
  });

  it("submits announcement kind title message and window fields", () => {
    const onSave = vi.fn();
    render(
      <StatusPageAnnouncementsPanel
        canManage
        onDelete={vi.fn()}
        onSave={onSave}
        page={{ id: "stp_1", announcements: [] }}
      />,
    );

    fireEvent.change(screen.getByLabelText("Kind"), { target: { value: "incident" } });
    fireEvent.change(screen.getByLabelText("Starts"), { target: { value: "2026-04-26T10:00" } });
    fireEvent.change(screen.getByLabelText("Ends"), { target: { value: "2026-04-26T11:00" } });
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Outage" } });
    fireEvent.change(screen.getByLabelText("Message"), { target: { value: "Investigating" } });
    fireEvent.click(screen.getByText("Save announcement"));

    expect(onSave).toHaveBeenCalledWith("stp_1", expect.objectContaining({
      kind: "incident",
      title: "Outage",
      message: "Investigating",
      startsAt: expect.stringContaining("2026-04-26T"),
      endsAt: expect.stringContaining("2026-04-26T"),
    }));
  });
});
