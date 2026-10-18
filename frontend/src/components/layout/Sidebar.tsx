import React from "react";
import { useLocation } from "@tanstack/react-router";
import { Sidebar as SidebarIcon } from "lucide-react";
import { SidebarMainContent } from "./SidebarMainContent";
import SettingsSidebar from "./SettingsSidebar";

export interface SidebarProps {
  collapsed: boolean;
  setCollapsed: (collapsed: boolean) => void;
  activeTab?: string;
}

export const Sidebar: React.FC<SidebarProps> = ({ collapsed, setCollapsed }) => {
  const { pathname } = useLocation();
  const isSettings = pathname.startsWith("/settings");

  return (
    <aside
      className={`h-screen bg-sidebar border-r border-surface-border flex flex-col transition-all duration-300 z-30 select-none ${
        collapsed ? "w-14" : "w-64"
      }`}
    >
      <div className="p-3.5 flex items-center justify-between border-b border-surface-border/40">
        <div className="flex items-center gap-2.5 overflow-hidden">
          <div className="w-8 h-8 rounded-full bg-zinc-900 border border-zinc-700/80 flex items-center justify-center flex-shrink-0 text-white shadow-sm">
            <span className="font-bold text-lg tracking-tighter">Ø</span>
          </div>
          {!collapsed && (
            <span className="font-semibold text-sm tracking-wide text-zinc-100">Aegis</span>
          )}
        </div>

        {!collapsed && (
          <div className="flex items-center gap-1">
            <button
              onClick={() => setCollapsed(true)}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60 transition-colors"
            >
              <SidebarIcon className="w-4 h-4" />
            </button>
          </div>
        )}

        {collapsed && (
          <div className="w-full flex items-center justify-between">
            <button
              onClick={() => setCollapsed(false)}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60 transition-colors mx-auto"
            >
              <SidebarIcon className="w-4 h-4" />
            </button>
          </div>
        )}
      </div>
      {isSettings ? <SettingsSidebar /> : <SidebarMainContent collapsed={collapsed} />}
    </aside>
  );
};
