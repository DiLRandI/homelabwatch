import PublicStatusPageScreen from "./PublicStatusPageScreen";

export default function StatusPagePreview({ page }) {
  if (!page) return null;
  const publicPage = {
    slug: page.slug,
    title: page.title,
    description: page.description,
    overallStatus: page.services?.length ? page.services.some((item) => item.status !== "healthy") ? "degraded" : "healthy" : "unknown",
    lastUpdatedAt: page.updatedAt,
    announcements: page.announcements || [],
    services: (page.services || []).map((item) => ({
      name: item.displayName || item.serviceName,
      status: item.status,
      lastCheckedAt: item.lastCheckedAt,
      latestCheck: item.latestCheck,
    })),
  };
  return <PublicStatusPageScreen embedded page={publicPage} />;
}
