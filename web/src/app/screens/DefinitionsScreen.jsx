import ServiceDefinitionsSection from "../../components/dashboard/ServiceDefinitionsSection";

export default function DefinitionsScreen({
  canManageUI,
  onDeleteServiceDefinition,
  onReapplyServiceDefinition,
  onSaveServiceDefinition,
  settings,
}) {
  return (
    <ServiceDefinitionsSection
      canManage={canManageUI}
      definitions={settings?.serviceDefinitions ?? []}
      onDeleteDefinition={onDeleteServiceDefinition}
      onReapplyDefinition={onReapplyServiceDefinition}
      onSaveDefinition={onSaveServiceDefinition}
    />
  );
}
