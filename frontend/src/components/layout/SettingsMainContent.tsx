import React, { useState, useEffect, useRef } from "react";
import {
  ChevronLeft,
  Search,
  User,
  Shield,
  Link2,
  Lock,
  Bell,
  Palette,
  ArrowLeftRight,
  Blocks,
} from "lucide-react";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { SettingsTab } from "@/pages/settings";
import { SidebarProps } from "./Sidebar";

export interface SettingsNavItem {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}

export const DEFAULT_SETTINGS_NAV_ITEMS: SettingsNavItem[] = [
  { id: "profile", label: "Profile", icon: User },
  { id: "providers", label: "Providers", icon: Blocks },
  { id: "security", label: "Security & sign-in", icon: Shield },
  { id: "connected-accounts", label: "Connected accounts", icon: Link2 },
  { id: "authorized-apps", label: "Authorized apps", icon: Lock },
  { id: "notifications", label: "Notifications", icon: Bell },
  { id: "appearance", label: "Appearance", icon: Palette },
  { id: "project-transfers", label: "Project transfers", icon: ArrowLeftRight },
];

export const SettingsMainContent: React.FC<SidebarProps> = ({ collapsed, setCollapsed }) => {
  const navigate = useNavigate();
  const searchParams = useSearch({ strict: false }) as { tab?: string } | undefined;
  const [searchQuery, setSearchQuery] = useState("");
  const searchInputRef = useRef<HTMLInputElement>(null);

  const currentTab = searchParams?.tab || "profile";

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (
        e.key === "/" &&
        document.activeElement !== searchInputRef.current &&
        !["INPUT", "TEXTAREA"].includes((document.activeElement as HTMLElement)?.tagName)
      ) {
        e.preventDefault();
        setCollapsed(false);
        setTimeout(() => {
          searchInputRef.current?.focus();
        }, 0);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [setCollapsed]);

  const handleItemClick = (id: string) => {
    navigate({
      to: "/settings",
      search: { tab: id as SettingsTab },
    });
  };

  const handleBack = () => {
    navigate({ to: "/" });
  };

  const filteredItems = DEFAULT_SETTINGS_NAV_ITEMS.filter((item) =>
    item.label.toLowerCase().includes(searchQuery.toLowerCase().trim())
  );

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden">
      {!collapsed && (
        <div className="pt-4 px-4 pb-2">
          <button
            onClick={handleBack}
            className="group inline-flex items-center gap-2 text-zinc-400 hover:text-white transition-colors text-sm font-medium focus:outline-none"
          >
            <ChevronLeft className="w-4 h-4 text-zinc-400 group-hover:text-white transition-transform group-hover:-translate-x-0.5" />
            <span>Back to app</span>
          </button>
        </div>
      )}

      <div
        onClick={() => {
          setCollapsed(false);
          setTimeout(() => {
            searchInputRef.current?.focus();
          }, 0);
        }}
        className="px-[10px] pt-2 pb-3 cursor-pointer"
      >
        <div className="relative flex items-center bg-zinc-900/90 hover:bg-zinc-900 border border-zinc-800 focus-within:border-zinc-700 rounded-lg px-2.5 py-1.5 transition-colors">
          <Search className="w-3.5 h-3.5 text-zinc-500 mr-2 flex-shrink-0" />
          <input
            ref={searchInputRef}
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search settings"
            className="w-full bg-transparent text-xs text-zinc-200 placeholder-zinc-500 outline-none pr-1"
          />
          {!collapsed && (
            <kbd className="flex-shrink-0 text-[10px] text-zinc-500 bg-zinc-800/90 border border-zinc-700/60 rounded px-1.5 py-0.5 font-mono leading-none">
              /
            </kbd>
          )}
        </div>
      </div>

      {/* Navigation Menu List */}
      <div className="flex-1 overflow-y-auto px-2 space-y-1">
        {filteredItems.length === 0 ? (
          <p className="text-xs text-zinc-500 italic px-3 py-4 text-center">No settings found</p>
        ) : (
          filteredItems.map((item) => {
            const Icon = item.icon;
            const isActive = currentTab === item.id;

            return (
              <button
                key={item.id}
                onClick={() => handleItemClick(item.id)}
                title={collapsed ? item.label : undefined}
                className={`w-full flex items-center ${
                  collapsed ? "justify-center px-2" : "gap-3 px-3"
                } py-2 rounded-lg text-sm font-medium transition-all group ${
                  isActive
                    ? "bg-[#1f1f23] text-white shadow-sm font-semibold"
                    : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40"
                }`}
              >
                <Icon
                  className={`w-4 h-4 flex-shrink-0 transition-colors ${
                    isActive ? "text-white" : "text-zinc-500 group-hover:text-zinc-300"
                  }`}
                />
                {!collapsed && <span className="truncate">{item.label}</span>}
              </button>
            );
          })
        )}
      </div>
    </div>
  );
};

export default SettingsMainContent;
