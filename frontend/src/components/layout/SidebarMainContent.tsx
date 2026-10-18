import React, { useState } from "react";
import { Link, useLocation } from "@tanstack/react-router";
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
  PlusIcon,
  Trash2,
} from "lucide-react";
import { useAegis } from "../../context/AegisContext";
import { useAuth } from "../../context/AuthContext";
import { AddCredentialModal } from "../modals/AddCredentialModal";
import { router } from "@/router";
import { MdOutlineKeyboardArrowRight } from "react-icons/md";
import { SidebarProps } from "./Sidebar";

export const SidebarMainContent: React.FC<Omit<SidebarProps, "setCollapsed">> = ({ collapsed }) => {
  const { user } = useAuth();
  const { chatThreads, activeThreadId, setActiveThreadId, createNewThread, deleteThread } =
    useAegis();

  const { pathname } = useLocation();

  const isChat = pathname === "/" || pathname.startsWith("/chat") || pathname.startsWith("/c");
  const isAnalytics = pathname.startsWith("/analytics");
  const isSettings = pathname.startsWith("/settings");

  const [addCredOpen, setAddCredOpen] = useState(false);
  return (
    <>
      {/* Main Action Links */}
      <div className="p-2 space-y-1">
        <button
          onClick={() => {
            setActiveThreadId(null);
            router.navigate({ to: "/" });
          }}
          className={`w-full flex items-center ${collapsed ? "justify-center" : "justify-start"} gap-3 px-3 py-2 rounded-md font-medium text-sm transition-all hover:bg-zinc-800 ${
            isChat && !activeThreadId
              ? "bg-zinc-800 text-white font-semibold"
              : "text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <PlusIcon className="w-4 h-4 text-zinc-300 flex-shrink-0" />
          {!collapsed && <span>New Chat</span>}
        </button>

        <Link
          to="/analytics"
          className={`w-full flex items-center ${collapsed ? "justify-center" : "justify-start"} gap-3 px-3 py-2 rounded-md font-medium text-sm transition-all hover:bg-zinc-800/60 ${
            isAnalytics
              ? "bg-zinc-800 text-white font-semibold"
              : "text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <BarChart2
            className={`w-4 h-4 flex-shrink-0 ${isAnalytics ? "text-white" : "text-zinc-400"}`}
          />
          {!collapsed && <span>Usage Analytics</span>}
        </Link>

        <Link
          to="/settings"
          search={{ tab: "profile" }}
          className={`w-full flex items-center ${collapsed ? "justify-center" : "justify-between"} gap-3 px-3 py-2 rounded-md font-medium text-sm transition-all hover:bg-zinc-800/60 ${
            isSettings
              ? "bg-zinc-800 text-white font-semibold"
              : "text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <div className="flex items-center gap-3">
            <Settings
              className={`w-4 h-4 flex-shrink-0 ${isSettings ? "text-white" : "text-zinc-400"}`}
            />
            {!collapsed && <span>Settings</span>}
          </div>
          {!collapsed && <MdOutlineKeyboardArrowRight />}
        </Link>
      </div>

      {!collapsed && <div className="mx-3 my-2 border-t border-surface-border/50"></div>}

      {/* Chat Threads Section */}
      {!collapsed && (
        <div className="px-3 py-2 flex-1 overflow-y-auto space-y-3">
          <div className="flex items-center justify-between text-xs text-zinc-400 font-medium">
            <span>Recent Chats</span>
            <button
              onClick={() => {
                setActiveThreadId(null);
                router.navigate({ to: "/" });
              }}
              className="p-1 text-zinc-400 hover:text-white rounded hover:bg-zinc-800/60"
              title="New Chat"
            >
              <Plus className="w-3.5 h-3.5" />
            </button>
          </div>

          <div className="space-y-0.5">
            {chatThreads.length === 0 ? (
              <p className="text-xs text-zinc-500 italic px-2 py-1">No chats yet</p>
            ) : (
              chatThreads.map((thread) => (
                <div
                  key={thread.id}
                  className={`group flex items-center justify-between w-full px-2.5 py-1.5 rounded-md text-xs transition-all ${
                    activeThreadId === thread.id
                      ? "bg-zinc-800 text-white font-medium"
                      : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40"
                  }`}
                >
                  <button
                    onClick={() => {
                      setActiveThreadId(thread.id);
                      router.navigate({ to: `/c/${thread.id}` });
                    }}
                    className="flex-1 text-left truncate mr-1.5"
                    title={thread.title}
                  >
                    {thread.title}
                  </button>
                  <button
                    onClick={async (e) => {
                      e.stopPropagation();
                      await deleteThread(thread.id);
                      if (activeThreadId === thread.id) {
                        router.navigate({ to: "/" });
                      }
                    }}
                    className="opacity-0 group-hover:opacity-100 p-1 text-zinc-500 hover:text-red-400 rounded transition-opacity"
                    title="Delete chat"
                  >
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>
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
          className="flex items-center gap-3 px-3 py-2 rounded-md text-zinc-300 text-xs font-medium hover:bg-zinc-800/50 transition-all cursor-pointer"
        >
          <div className="w-6 h-6 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-zinc-300 text-xs font-semibold flex-shrink-0">
            {user?.first_name && user?.last_name
              ? user.first_name.charAt(0).toUpperCase() + user.last_name.charAt(0).toUpperCase()
              : user?.email
                ? user.email.charAt(0).toUpperCase() + user.email.charAt(0).toUpperCase()
                : "FE"}
          </div>
          {!collapsed && (
            <span className="truncate">{user?.name || user?.email || "User Profile"}</span>
          )}
        </Link>
      </div>

      <AddCredentialModal open={addCredOpen} onOpenChange={setAddCredOpen} />
    </>
  );
};
