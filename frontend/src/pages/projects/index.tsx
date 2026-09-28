import React, { useState } from 'react';
import { MainLayout } from '../../components/layout/MainLayout';
import { useAegis } from '../../context/AegisContext';
import { Card } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { Plus, Folder, Key, ArrowRight, ShieldCheck } from 'lucide-react';
import { CreateProjectModal } from '../../components/modals/CreateProjectModal';

export const ProjectsPage: React.FC = () => {
  const { projects, activeOrg, setActiveProject } = useAegis();
  const [createOpen, setCreateOpen] = useState(false);

  return (
    <MainLayout>
      <div className="max-w-5xl mx-auto p-8 space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white tracking-tight">Projects</h1>
            <p className="text-xs text-zinc-400 mt-1">
              Active projects under organization <span className="text-amber-400 font-semibold">{activeOrg?.name}</span>.
            </p>
          </div>
          <Button onClick={() => setCreateOpen(true)} className="flex items-center gap-2">
            <Plus className="w-4 h-4" /> New Project
          </Button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {projects.map((proj) => (
            <Card
              key={proj.id}
              className="p-5 flex flex-col justify-between cursor-pointer group hover:border-zinc-600"
              onClick={() => {
                setActiveProject(proj);
                window.location.href = `/projects/${proj.id}`;
              }}
            >
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="w-10 h-10 rounded-xl bg-zinc-900 border border-zinc-800 flex items-center justify-center text-amber-400">
                    <Folder className="w-5 h-5" />
                  </div>
                  <span className="text-[10px] font-mono text-zinc-500 uppercase tracking-wider bg-zinc-900 px-2 py-0.5 rounded border border-zinc-800">
                    {proj.slug || 'active'}
                  </span>
                </div>
                <div>
                  <h3 className="text-base font-bold text-white group-hover:text-amber-400 transition-colors">
                    {proj.name}
                  </h3>
                  <p className="text-xs text-zinc-400 mt-1">
                    {proj.description || 'Distributed inference project worker group.'}
                  </p>
                </div>
              </div>

              <div className="pt-4 border-t border-surface-border mt-4 flex items-center justify-between text-xs text-zinc-400">
                <span className="flex items-center gap-1.5">
                  <ShieldCheck className="w-3.5 h-3.5 text-emerald-400" /> Protected Endpoint
                </span>
                <span className="flex items-center gap-1 text-zinc-300 group-hover:translate-x-1 transition-transform font-medium">
                  View Details <ArrowRight className="w-3.5 h-3.5" />
                </span>
              </div>
            </Card>
          ))}
        </div>

        <CreateProjectModal open={createOpen} onOpenChange={setCreateOpen} />
      </div>
    </MainLayout>
  );
};

export default ProjectsPage;
