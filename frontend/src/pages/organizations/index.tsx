import React, { useState } from 'react';
import { MainLayout } from '../../components/layout/MainLayout';
import { useAegis } from '../../context/AegisContext';
import { Card } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { Badge } from '../../components/ui/Badge';
import { Plus, Layers, Key, Users, ArrowRight } from 'lucide-react';
import { CreateOrgModal } from '../../components/modals/CreateOrgModal';
import { AddCredentialModal } from '../../components/modals/AddCredentialModal';

export const OrganizationsPage: React.FC = () => {
  const { organizations, activeOrg, setActiveOrg } = useAegis();
  const [createOrgOpen, setCreateOrgOpen] = useState(false);
  const [addCredOpen, setAddCredOpen] = useState(false);

  return (
    <MainLayout>
      <div className="max-w-5xl mx-auto p-8 space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white tracking-tight">Organizations & BYOK Vault</h1>
            <p className="text-xs text-zinc-400 mt-1">
              Manage organization workspaces, team access, and upstream LLM provider keys.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Button variant="outline" onClick={() => setAddCredOpen(true)} className="flex items-center gap-2">
              <Key className="w-4 h-4 text-amber-400" /> Add Provider Key
            </Button>
            <Button onClick={() => setCreateOrgOpen(true)} className="flex items-center gap-2">
              <Plus className="w-4 h-4" /> New Organization
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {organizations.map((org) => (
            <Card
              key={org.id}
              className={`p-5 flex flex-col justify-between cursor-pointer group ${
                activeOrg?.id === org.id ? 'border-amber-500/50 bg-surface-hover' : ''
              }`}
              onClick={() => {
                setActiveOrg(org);
                window.location.href = `/organizations/${org.id}`;
              }}
            >
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="w-10 h-10 rounded-xl bg-zinc-900 border border-zinc-800 flex items-center justify-center text-amber-400">
                    <Layers className="w-5 h-5" />
                  </div>
                  {activeOrg?.id === org.id && <Badge variant="warning">Active Workspace</Badge>}
                </div>
                <div>
                  <h3 className="text-base font-bold text-white group-hover:text-amber-400 transition-colors">
                    {org.name}
                  </h3>
                  <p className="text-xs text-zinc-400 mt-1 font-mono">slug: {org.slug || 'default'}</p>
                </div>
              </div>

              <div className="pt-4 border-t border-surface-border mt-4 flex items-center justify-between text-xs text-zinc-400">
                <span className="flex items-center gap-1.5">
                  <Users className="w-3.5 h-3.5 text-zinc-400" /> Shared Access
                </span>
                <span className="flex items-center gap-1 text-zinc-300 group-hover:translate-x-1 transition-transform font-medium">
                  Manage Credentials <ArrowRight className="w-3.5 h-3.5" />
                </span>
              </div>
            </Card>
          ))}
        </div>

        <CreateOrgModal open={createOrgOpen} onOpenChange={setCreateOrgOpen} />
        <AddCredentialModal open={addCredOpen} onOpenChange={setAddCredOpen} />
      </div>
    </MainLayout>
  );
};

export default OrganizationsPage;
