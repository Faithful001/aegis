import React, { useState } from 'react';
import { 
  MessageSquare, 
  Image as ImagineIcon, 
  BookOpen, 
  Cpu, 
  Plus, 
  Search, 
  Sidebar as SidebarIcon, 
  Folder, 
  Sparkles, 
  User as UserIcon, 
  Grid,
  ChevronDown,
  Layers
} from 'lucide-react';
import { useAegis } from '../../context/AegisContext';
import { useAuth } from '../../context/AuthContext';
import { CreateProjectModal } from '../modals/CreateProjectModal';
import { CreateOrgModal } from '../modals/CreateOrgModal';

interface SidebarProps {
  collapsed: boolean;
  setCollapsed: (collapsed: boolean) => void;
  activeTab?: string;
}

export const Sidebar: React.FC<SidebarProps> = ({ collapsed, setCollapsed }) => {
  const { user } = useAuth();
  const {
    organizations,
    activeOrg,
    setActiveOrg,
    projects,
    activeProject,
    setActiveProject,
    chatThreads,
    activeThreadId,
    setActiveThreadId,
    createNewThread,
  } = useAegis();

  const [createProjectOpen, setCreateProjectOpen] = useState(false);
  const [createOrgOpen, setCreateOrgOpen] = useState(false);
  const [showOrgDropdown, setShowOrgDropdown] = useState(false);

  return (
    <aside
      className={`h-screen bg-sidebar border-r border-surface-border flex flex-col transition-all duration-300 z-30 select-none ${
        collapsed ? 'w-16' : 'w-64'
      }`}
    >
      {/* Top Header Logo & Search Controls */}
      <div className="p-3.5 flex items-center justify-between border-b border-surface-border/40">
        <div className="flex items-center gap-2.5 overflow-hidden">
          <div className="w-8 h-8 rounded-full bg-zinc-900 border border-zinc-700/80 flex items-center justify-center flex-shrink-0 text-white shadow-sm">
            <span className="font-bold text-lg tracking-tighter">Ø</span>
          </div>
          {!collapsed && (
            <span className="font-semibold text-sm tracking-wide text-zinc-100">AEGIS</span>
          )}
        </div>

        {!collapsed && (
          <div className="flex items-center gap-1">
            <button className="p-1.5 rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60 transition-colors">
              <Search className="w-4 h-4" />
            </button>
            <button
              onClick={() => setCollapsed(true)}
              className="p-1.5 rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60 transition-colors"
            >
              <SidebarIcon className="w-4 h-4" />
            </button>
          </div>
        )}

        {collapsed && (
          <button
            onClick={() => setCollapsed(false)}
            className="p-1.5 rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60 transition-colors mx-auto"
          >
            <SidebarIcon className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Main Navigation links */}
      <div className="p-2 space-y-1">
        <button
          onClick={() => {
            createNewThread();
          }}
          className="w-full flex items-center gap-3 px-3 py-2 rounded-xl bg-surface-hover text-zinc-100 font-medium text-sm transition-all hover:bg-zinc-800"
        >
          <MessageSquare className="w-4 h-4 text-zinc-300 flex-shrink-0" />
          {!collapsed && <span>Chat</span>}
        </button>

        <button className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 font-medium text-sm transition-all">
          <div className="flex items-center gap-3">
            <ImagineIcon className="w-4 h-4 text-zinc-400 flex-shrink-0" />
            {!collapsed && <span>Imagine</span>}
          </div>
          {!collapsed && <span className="w-1.5 h-1.5 rounded-full bg-blue-500"></span>}
        </button>

        <button className="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 font-medium text-sm transition-all">
          <BookOpen className="w-4 h-4 text-zinc-400 flex-shrink-0" />
          {!collapsed && <span>Library</span>}
        </button>

        <button className="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 font-medium text-sm transition-all">
          <Cpu className="w-4 h-4 text-zinc-400 flex-shrink-0" />
          {!collapsed && <span>Automations</span>}
        </button>
      </div>

      {!collapsed && <div className="mx-3 my-2 border-t border-surface-border/50"></div>}

      {/* Organization & Projects Section */}
      {!collapsed && (
        <div className="px-3 py-2 flex-1 overflow-y-auto space-y-4">
          {/* Org Selector */}
          <div>
            <div className="flex items-center justify-between mb-1.5 text-xs text-zinc-400 font-medium">
              <span>Organization</span>
              <button
                onClick={() => setCreateOrgOpen(true)}
                className="text-zinc-400 hover:text-white p-0.5 hover:bg-zinc-800 rounded"
              >
                <Plus className="w-3.5 h-3.5" />
              </button>
            </div>
            <div className="relative">
              <button
                onClick={() => setShowOrgDropdown(!showOrgDropdown)}
                className="w-full flex items-center justify-between px-3 py-1.5 bg-surface-card border border-surface-border rounded-xl text-xs text-zinc-200 hover:bg-zinc-800/50"
              >
                <div className="flex items-center gap-2 truncate">
                  <Layers className="w-3.5 h-3.5 text-amber-400 flex-shrink-0" />
                  <span className="truncate">{activeOrg ? activeOrg.name : 'Select Organization'}</span>
                </div>
                <ChevronDown className="w-3.5 h-3.5 text-zinc-400 flex-shrink-0" />
              </button>

              {showOrgDropdown && (
                <div className="absolute left-0 right-0 top-full mt-1 bg-surface border border-surface-border rounded-xl p-1 shadow-2xl z-50 space-y-0.5">
                  {organizations.map((org) => (
                    <button
                      key={org.id}
                      onClick={() => {
                        setActiveOrg(org);
                        setShowOrgDropdown(false);
                      }}
                      className={`w-full text-left px-2.5 py-1.5 text-xs rounded-lg transition-colors ${
                        activeOrg?.id === org.id
                          ? 'bg-zinc-800 text-white font-medium'
                          : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'
                      }`}
                    >
                      {org.name}
                    </button>
                  ))}
                  <button
                    onClick={() => {
                      setShowOrgDropdown(false);
                      setCreateOrgOpen(true);
                    }}
                    className="w-full text-left px-2.5 py-1.5 text-xs text-amber-400 hover:bg-zinc-800/50 rounded-lg flex items-center gap-1.5 font-medium border-t border-surface-border mt-1 pt-1.5"
                  >
                    <Plus className="w-3 h-3" /> New Organization
                  </button>
                </div>
              )}
            </div>
          </div>

          {/* Projects */}
          <div>
            <div className="flex items-center justify-between mb-1.5 text-xs text-zinc-400 font-medium">
              <span>Projects</span>
              <button
                onClick={() => setCreateProjectOpen(true)}
                className="flex items-center gap-1 text-zinc-400 hover:text-white text-xs px-1.5 py-0.5 rounded hover:bg-zinc-800/60"
              >
                <Plus className="w-3 h-3" /> Add project
              </button>
            </div>

            <div className="space-y-0.5">
              {projects.length === 0 ? (
                <p className="text-xs text-zinc-400 italic px-2 py-1">No projects yet</p>
              ) : (
                projects.map((proj) => (
                  <button
                    key={proj.id}
                    onClick={() => setActiveProject(proj)}
                    className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-xl text-xs font-medium transition-all ${
                      activeProject?.id === proj.id
                        ? 'bg-zinc-800/90 text-white'
                        : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40'
                    }`}
                  >
                    <Folder className="w-3.5 h-3.5 text-zinc-400 flex-shrink-0" />
                    <span className="truncate">{proj.name}</span>
                  </button>
                ))
              )}
            </div>
          </div>

          {/* Chats list */}
          <div>
            <div className="text-xs text-zinc-400 font-medium mb-1.5">Chats</div>
            <div className="space-y-0.5">
              {chatThreads.map((thread) => (
                <button
                  key={thread.id}
                  onClick={() => setActiveThreadId(thread.id)}
                  className={`w-full text-left px-2.5 py-1.5 rounded-xl text-xs truncate transition-all ${
                    activeThreadId === thread.id
                      ? 'bg-zinc-800 text-white font-medium'
                      : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40'
                  }`}
                >
                  {thread.title}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {collapsed && <div className="flex-1"></div>}

      {/* Bottom Footer User Info & Plugins */}
      <div className="p-2.5 border-t border-surface-border/60 space-y-1">
        <a
          href="/settings"
          className="flex items-center gap-3 px-3 py-2 rounded-xl text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/50 text-xs font-medium transition-all"
        >
          <Grid className="w-4 h-4 text-zinc-400 flex-shrink-0" />
          {!collapsed && <span>Plugins</span>}
        </a>

        <div className="flex items-center gap-3 px-3 py-2 rounded-xl text-zinc-300 text-xs font-medium">
          <div className="w-6 h-6 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-zinc-300 text-xs font-semibold flex-shrink-0">
            {user?.name ? user.name.slice(0, 2).toUpperCase() : 'FE'}
          </div>
          {!collapsed && <span className="truncate">{user?.name || user?.email || 'Faithful Eromosele'}</span>}
        </div>
      </div>

      <CreateProjectModal open={createProjectOpen} onOpenChange={setCreateProjectOpen} />
      <CreateOrgModal open={createOrgOpen} onOpenChange={setCreateOrgOpen} />
    </aside>
  );
};
