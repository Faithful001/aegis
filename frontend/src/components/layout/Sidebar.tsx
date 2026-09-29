import React, { useState } from "react";
import { Link } from "@tanstack/react-router";
import {
  MessageSquare,
  Key,
  BarChart2,
  Plus,
  Search,
  Sidebar as SidebarIcon,
  Grid,
  Shield,
  Settings,
} from "lucide-react";
import { useAegis } from "../../context/AegisContext";
import { useAuth } from "../../context/AuthContext";
import { AddCredentialModal } from "../modals/AddCredentialModal";

interface SidebarProps {
  collapsed: boolean;
  setCollapsed: (collapsed: boolean) => void;
  activeTab?: string;
}

export const Sidebar: React.FC<SidebarProps> = ({ collapsed, setCollapsed }) => {
  const { user } = useAuth();
  const { credentials, chatThreads, activeThreadId, setActiveThreadId, createNewThread } =
    useAegis();

  const [addCredOpen, setAddCredOpen] = useState(false);

  return (
    <aside
      className={`h-screen bg-sidebar border-r border-surface-border flex flex-col transition-all duration-300 z-30 select-none ${
        collapsed ? "w-16" : "w-64"
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

      {/* Main Action Links */}
      <div className="p-2 space-y-1">
        <button
          onClick={() => {
            createNewThread();
          }}
          className="w-full flex items-center gap-3 px-3 py-2 rounded-xl bg-surface-hover text-zinc-100 font-medium text-sm transition-all hover:bg-zinc-800"
        >
          <MessageSquare className="w-4 h-4 text-zinc-300 flex-shrink-0" />
          {!collapsed && <span>New Chat</span>}
        </button>

        <button
          onClick={() => setAddCredOpen(true)}
          className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 font-medium text-sm transition-all"
        >
          <div className="flex items-center gap-3">
            <Key className="w-4 h-4 text-amber-400 flex-shrink-0" />
            {!collapsed && <span>BYOK Vault</span>}
          </div>
          {!collapsed && (
            <span className="text-[10px] font-mono bg-zinc-900 border border-zinc-800 px-1.5 py-0.5 rounded text-zinc-400">
              {credentials.length} Keys
            </span>
          )}
        </button>

        <Link
          to="/analytics"
          className="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 font-medium text-sm transition-all"
        >
          <BarChart2 className="w-4 h-4 text-zinc-400 flex-shrink-0" />
          {!collapsed && <span>Usage Analytics</span>}
        </Link>

        <Link
          to="/settings"
          className="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 font-medium text-sm transition-all"
        >
          <Settings className="w-4 h-4 text-zinc-400 flex-shrink-0" />
          {!collapsed && <span>Settings</span>}
        </Link>
      </div>

      {!collapsed && <div className="mx-3 my-2 border-t border-surface-border/50"></div>}

      {/* Chat Threads Section */}
      {!collapsed && (
        <div className="px-3 py-2 flex-1 overflow-y-auto space-y-3">
          <div className="flex items-center justify-between text-xs text-zinc-400 font-medium">
            <span>Recent Chats</span>
            <button
              onClick={createNewThread}
              className="p-1 text-zinc-400 hover:text-white rounded hover:bg-zinc-800/60"
            >
              <Plus className="w-3.5 h-3.5" />
            </button>
          </div>

          <div className="space-y-0.5">
            {chatThreads.length === 0 ? (
              <p className="text-xs text-zinc-500 italic px-2 py-1">No chats yet</p>
            ) : (
              chatThreads.map((thread) => (
                <button
                  key={thread.id}
                  onClick={() => setActiveThreadId(thread.id)}
                  className={`w-full text-left px-2.5 py-1.5 rounded-xl text-xs truncate transition-all ${
                    activeThreadId === thread.id
                      ? "bg-zinc-800 text-white font-medium"
                      : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40"
                  }`}
                >
                  {thread.title}
                </button>
              ))
            )}
          </div>
        </div>
      )}

      {collapsed && <div className="flex-1"></div>}

      {/* Bottom Footer User Info */}
      <div className="p-2.5 border-t border-surface-border/60 space-y-1">
        <Link
          to="/settings"
          className="flex items-center gap-3 px-3 py-2 rounded-xl text-zinc-300 text-xs font-medium hover:bg-zinc-800/50 transition-all cursor-pointer"
        >
          <div className="w-6 h-6 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-zinc-300 text-xs font-semibold flex-shrink-0">
            {user?.name
              ? user.name.slice(0, 2).toUpperCase()
              : user?.email
                ? user.email.slice(0, 2).toUpperCase()
                : "FE"}
          </div>
          {!collapsed && (
            <span className="truncate">{user?.name || user?.email || "User Profile"}</span>
          )}
        </Link>
      </div>

      <AddCredentialModal open={addCredOpen} onOpenChange={setAddCredOpen} />
    </aside>
  );
};
