import {
  BookmarkIcon,
  DatabaseIcon,
  DevicesIcon,
  DiscoveryIcon,
  OverviewIcon,
  ServicesIcon,
  ShieldIcon,
  TokenIcon,
} from "../components/ui/Icons";

export const APP_ROUTES = [
  {
    countKey: "dashboard",
    icon: OverviewIcon,
    id: "dashboard",
    label: "Dashboard",
    path: "/",
    subtitle: "Favorites, fleet status, and recent control-plane activity.",
    title: "Dashboard",
  },
  {
    countKey: "bookmarks",
    icon: BookmarkIcon,
    id: "bookmarks",
    label: "Bookmarks",
    path: "/bookmarks",
    subtitle: "Curated links, folders, tags, and favorites for daily operations.",
    title: "Bookmarks",
  },
  {
    countKey: "services",
    icon: ServicesIcon,
    id: "services",
    label: "Services",
    path: "/services",
    subtitle: "Accepted services, Docker workloads, and bookmark promotion.",
    title: "Services",
  },
  {
    countKey: "health",
    icon: ShieldIcon,
    id: "health",
    label: "Health",
    path: "/health",
    subtitle: "Health checks, test flows, and monitored service status.",
    title: "Health",
  },
  {
    countKey: "discovery",
    icon: DiscoveryIcon,
    id: "discovery",
    label: "Discovery",
    path: "/discovery",
    subtitle: "Docker endpoints, scan targets, policy, and discovery review.",
    title: "Discovery",
  },
  {
    countKey: "devices",
    icon: DevicesIcon,
    id: "devices",
    label: "Devices",
    path: "/devices",
    subtitle: "Known devices, network addresses, open ports, and confidence.",
    title: "Devices",
  },
  {
    countKey: "definitions",
    icon: DatabaseIcon,
    id: "definitions",
    label: "Definitions",
    path: "/definitions",
    subtitle: "Fingerprint matchers and managed health-check templates.",
    title: "Service Definitions",
  },
  {
    countKey: "settings",
    icon: TokenIcon,
    id: "settings",
    label: "Settings",
    path: "/settings",
    subtitle: "API access, worker status, and appliance-level administration.",
    title: "Settings",
  },
];

export function normalizePath(pathname) {
  if (!pathname || pathname === "/") {
    return "/";
  }

  const trimmed = pathname.replace(/\/+$/, "");
  return trimmed || "/";
}

export function getRoute(pathname) {
  const normalized = normalizePath(pathname);
  return APP_ROUTES.find((route) => route.path === normalized) || APP_ROUTES[0];
}
