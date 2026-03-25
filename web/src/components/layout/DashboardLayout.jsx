import { useState } from "react";

import Navbar from "./Navbar";
import Sidebar from "./Sidebar";

export default function DashboardLayout({
  activeHref,
  alerts,
  children,
  metrics,
  navItems,
  onNavigate,
  sidebarMeta,
  statusItems,
  subtitle,
  title,
  toolbar,
}) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="grid min-h-screen gap-0 lg:grid-cols-[280px_minmax(0,1fr)]">
      <Sidebar
        activeHref={activeHref}
        metrics={metrics}
        navItems={navItems}
        onClose={() => setSidebarOpen(false)}
        onNavigate={onNavigate}
        open={sidebarOpen}
        sidebarMeta={sidebarMeta}
      />
      <div className="min-w-0">
        <Navbar
          onOpenSidebar={() => setSidebarOpen(true)}
          statusItems={statusItems}
          subtitle={subtitle}
          title={title}
          toolbar={toolbar}
        />
        <main className="px-4 py-6 sm:px-6 lg:px-8">
          {alerts ? <div className="mb-6">{alerts}</div> : null}
          <div className="grid gap-6">{children}</div>
        </main>
      </div>
    </div>
  );
}
