import { useState } from "react";

import ApiAccessSection from "../../components/dashboard/ApiAccessSection";
import WorkersSection from "../../components/dashboard/WorkersSection";
import APITokenForm from "../../components/forms/APITokenForm";
import Modal from "../../components/ui/Modal";

export default function SettingsScreen({
  canManageUI,
  dashboard,
  onCreateAPIToken,
  onRevokeAPIToken,
  settings,
}) {
  const [creatingToken, setCreatingToken] = useState(false);
  const [createdToken, setCreatedToken] = useState(null);

  async function handleCreateToken(payload) {
    const created = await onCreateAPIToken(payload);
    if (!created) {
      return false;
    }

    setCreatedToken(created);
    setCreatingToken(false);
    return true;
  }

  return (
    <>
      <ApiAccessSection
        canManage={canManageUI}
        createdToken={createdToken}
        legacyTokenAlive={settings?.apiAccess?.legacyAdminTokenAlive ?? false}
        onCreate={() => setCreatingToken(true)}
        onDismissCreatedToken={() => setCreatedToken(null)}
        onRevoke={onRevokeAPIToken}
        tokens={settings?.apiAccess?.tokens ?? []}
      />
      <WorkersSection
        jobState={settings?.jobState ?? []}
        recentEvents={dashboard?.recentEvents ?? []}
        showRecentEvents={false}
      />
      <Modal
        description="Create a scoped bearer token for scripts, integrations, or external dashboards."
        onClose={() => setCreatingToken(false)}
        open={creatingToken}
        title="Create external API token"
      >
        <APITokenForm onSubmit={handleCreateToken} />
      </Modal>
    </>
  );
}
