import { useState } from "react";

import { formatDate } from "../../lib/format";
import Badge from "../ui/Badge";
import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import { ShieldIcon, TokenIcon } from "../ui/Icons";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/Table";

function SecretReveal({ createdToken, onDismiss }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(createdToken.secret);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1800);
    } catch {
      setCopied(false);
    }
  }

  return (
    <div className="rounded-3xl border border-accent/20 bg-accent/5 p-5">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-accent-strong">
            Save this token now
          </p>
          <p className="mt-2 text-sm leading-6 text-slate-600">
            This secret is shown once. Use it with the external automation API,
            not in the browser UI.
          </p>
          <code className="mt-4 block overflow-x-auto rounded-2xl border border-white bg-white px-4 py-3 text-sm font-medium text-slate-900">
            {createdToken.secret}
          </code>
        </div>
        <div className="flex gap-3">
          <Button onClick={handleCopy} variant="secondary">
            {copied ? "Copied" : "Copy token"}
          </Button>
          <Button onClick={onDismiss} variant="ghost">
            Dismiss
          </Button>
        </div>
      </div>
    </div>
  );
}

export default function ApiAccessSection({
  canManage = true,
  createdToken,
  legacyTokenAlive,
  onCreate,
  onDismissCreatedToken,
  onRevoke,
  tokens,
}) {
  return (
    <section id="settings">
      <Card>
        <CardHeader
          action={
            <Button disabled={!canManage} leadingIcon={TokenIcon} onClick={onCreate}>
              Create token
            </Button>
          }
          description="Generate and revoke bearer tokens for external scripts, dashboards, and automations."
          title="API access"
        />
        <CardContent className="grid gap-5">
          {createdToken ? (
            <SecretReveal
              createdToken={createdToken}
              onDismiss={onDismissCreatedToken}
            />
          ) : null}

          <div className="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(280px,0.8fr)]">
            <div className="rounded-3xl border border-slate-200 bg-slate-50 p-5">
              <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                Access model
              </p>
              <p className="mt-3 text-sm leading-6 text-slate-600">
                The local dashboard remains open for trusted LAN browsers.
                External clients should use bearer tokens against the external
                API and rotate them periodically.
              </p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-5">
              <div className="flex items-center gap-3">
                <span className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-slate-100 text-slate-600">
                  <ShieldIcon className="h-5 w-5" />
                </span>
                <div>
                  <p className="text-sm font-medium text-slate-900">
                    Legacy compatibility
                  </p>
                  <p className="text-sm text-slate-500">
                    {legacyTokenAlive
                      ? "A legacy admin token is still accepted by the external API."
                      : "No legacy admin token is active."}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {tokens.length === 0 ? (
            <EmptyState
              action={canManage ? onCreate : undefined}
              actionLabel="Create token"
              body="Use scoped bearer tokens for external apps, integrations, and custom scripts."
              title="No external API tokens created yet"
            />
          ) : (
            <Table>
              <TableHead>
                <tr>
                  <TableHeader>Token</TableHeader>
                  <TableHeader>Scope</TableHeader>
                  <TableHeader>Last used</TableHeader>
                  <TableHeader>Created</TableHeader>
                  <TableHeader className="text-right">Actions</TableHeader>
                </tr>
              </TableHead>
              <TableBody>
                {tokens.map((token) => (
                  <TableRow key={token.id}>
                    <TableCell className="min-w-[220px]">
                      <p className="font-medium text-slate-900">{token.name}</p>
                      <p className="mt-1 text-sm text-slate-500">{token.prefix}...</p>
                    </TableCell>
                    <TableCell>
                      <Badge tone={token.scope === "write" ? "accent" : "neutral"}>
                        {token.scope}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatDate(token.lastUsedAt)}</TableCell>
                    <TableCell>{formatDate(token.createdAt)}</TableCell>
                    <TableCell className="text-right">
                      {token.revokedAt ? (
                        <Badge tone="danger">revoked</Badge>
                      ) : (
                        <Button
                          disabled={!canManage}
                          onClick={() => void onRevoke(token.id)}
                          size="sm"
                          variant="ghost"
                        >
                          Revoke
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
