import React, { useEffect, useState } from "react";
import { useLocation } from "@tanstack/react-router";
import { SidebarSimpleIcon as SidebarIcon } from "@phosphor-icons/react";
import { motion, AnimatePresence } from "framer-motion";
import { SidebarMainContent } from "./SidebarMainContent";
import SettingsMainContent from "./SettingsMainContent";

export interface SidebarProps {
  collapsed: boolean;
  setCollapsed: (collapsed: boolean) => void;
  activeTab?: string;
}

function useIsSmallScreen() {
  const [small, setSmall] = useState(
    () => typeof window !== "undefined" && window.matchMedia("(max-width: 767px)").matches
  );
  useEffect(() => {
    const mq = window.matchMedia("(max-width: 767px)");
    const handler = (e: MediaQueryListEvent) => setSmall(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);
  return small;
}

export const Sidebar: React.FC<SidebarProps> = ({ collapsed, setCollapsed }) => {
  const { pathname } = useLocation();
  const isSettings = pathname.startsWith("/settings");
  const isSmallScreen = useIsSmallScreen();

  return (
    <>
      {/* Floating open button — small screens only, shown while sidebar is off-screen */}
      {collapsed && (
        <button
          onClick={() => setCollapsed(false)}
          className="fixed top-3 left-3 z-40 flex md:hidden items-center justify-center w-9 h-9 rounded-xl bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors shadow-lg"
        >
          <SidebarIcon className="w-4 h-4" />
        </button>
      )}

      <motion.aside
        initial={false}
        animate={{
          // On small screens: slide off-screen left when collapsed, slide in when open
          // On md+: always x=0, width handles the slim-rail toggle via CSS
          x: isSmallScreen && collapsed ? "-100%" : 0,
        }}
        transition={{ type: "spring", stiffness: 300, damping: 30, mass: 0.8 }}
        className={`h-screen bg-sidebar border-r border-surface-border flex flex-col select-none z-30
          fixed md:relative
          transition-[width] duration-300
          ${collapsed ? "w-64 md:w-14" : "w-64"}`}
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

          {/* Sidebar's own close button — shown when expanded */}
          {!collapsed && (
            <button
              onClick={() => setCollapsed(true)}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60 transition-colors"
            >
              <SidebarIcon className="w-4 h-4" />
            </button>
          )}

          {/* Expand button — only meaningful on md+ slim rail */}
          {collapsed && (
            <button
              onClick={() => setCollapsed(false)}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60 transition-colors mx-auto"
            >
              <SidebarIcon className="w-4 h-4" />
            </button>
          )}
        </div>

        <div className="relative flex-1 overflow-hidden flex flex-col">
          <AnimatePresence mode="wait" initial={false}>
            {isSettings ? (
              <motion.div
                key="settings-sidebar"
                initial={{ x: 24, opacity: 0 }}
                animate={{ x: 0, opacity: 1 }}
                exit={{ x: 24, opacity: 0 }}
                transition={{ duration: 0.18, ease: "easeInOut" }}
                className="flex-1 flex flex-col h-full overflow-hidden"
              >
                <SettingsMainContent collapsed={collapsed} setCollapsed={setCollapsed} />
              </motion.div>
            ) : (
              <motion.div
                key="main-sidebar"
                initial={{ x: -24, opacity: 0 }}
                animate={{ x: 0, opacity: 1 }}
                exit={{ x: -24, opacity: 0 }}
                transition={{ duration: 0.18, ease: "easeInOut" }}
                className="flex-1 flex flex-col h-full overflow-hidden"
              >
                <SidebarMainContent collapsed={collapsed} setCollapsed={setCollapsed} />
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.aside>
    </>
  );
};
