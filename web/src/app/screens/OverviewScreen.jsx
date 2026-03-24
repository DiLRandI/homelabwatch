import { useMemo, useState } from "react";

import FavoritesStrip from "../../components/bookmarks/FavoritesStrip";
import DashboardHeader from "../../components/dashboard/DashboardHeader";
import WorkersSection from "../../components/dashboard/WorkersSection";
import BookmarkSuggestionDialog from "../../components/discovery/BookmarkSuggestionDialog";
import DiscoveredServicesPanel from "../../components/discovery/DiscoveredServicesPanel";
import { bookmarkOpenURL } from "../../lib/api";

function sortFavorites(bookmarks) {
  return [...bookmarks]
    .filter((bookmark) => bookmark.isFavorite)
    .sort((left, right) => {
      if (left.favoritePosition !== right.favoritePosition) {
        return left.favoritePosition - right.favoritePosition;
      }
      return left.name.localeCompare(right.name);
    });
}

export default function OverviewScreen({
  bookmarks = [],
  canManageUI,
  dashboard,
  folders = [],
  metrics,
  onIgnoreDiscoveredService,
  onNavigate,
  onRestoreDiscoveredService,
  onSaveBookmarkFromDiscoveredService,
  settings,
}) {
  const [selectedDiscoveredService, setSelectedDiscoveredService] = useState(null);
  const favorites = useMemo(() => sortFavorites(bookmarks), [bookmarks]);
  const pendingDiscoveredServices = (dashboard?.discoveredServices ?? []).filter(
    (item) => item.state === "pending" || item.state === "ignored",
  );

  function handleOpenBookmark(bookmark) {
    window.open(bookmarkOpenURL(bookmark.id), "_blank", "noopener,noreferrer");
  }

  function handleDashboardAction(target) {
    switch (target) {
      case "service":
        onNavigate("/services");
        break;
      case "bookmark":
        onNavigate("/bookmarks");
        break;
      case "apiToken":
        onNavigate("/settings");
        break;
      default:
        onNavigate("/");
        break;
    }
  }

  async function handleCreateBookmark(item, payload) {
    return onSaveBookmarkFromDiscoveredService(item.id, payload);
  }

  return (
    <>
      <DashboardHeader
        canManageUI={canManageUI}
        metrics={metrics}
        onOpenModal={handleDashboardAction}
        settings={settings}
      />
      <FavoritesStrip bookmarks={favorites} onOpen={handleOpenBookmark} />
      <DiscoveredServicesPanel
        canManage={canManageUI}
        items={pendingDiscoveredServices}
        onCreateBookmark={(item) => setSelectedDiscoveredService(item)}
        onIgnore={(item) => void onIgnoreDiscoveredService(item.id)}
        onRestore={(item) => void onRestoreDiscoveredService(item.id)}
      />
      <WorkersSection
        jobState={settings?.jobState ?? []}
        recentEvents={dashboard?.recentEvents ?? []}
        showWorkers={false}
      />
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
