import DevicesSection from "../../components/dashboard/DevicesSection";

export default function DevicesScreen({
  dashboard,
  discoveryCounts,
  serviceCounts,
}) {
  return (
    <DevicesSection
      devices={dashboard?.devices ?? []}
      discoveryCounts={discoveryCounts}
      serviceCounts={serviceCounts}
    />
  );
}
