import React from "react";
import { ProfileTab } from "@/pages/settings/ProfileTab";
import { ProvidersTab } from "@/pages/settings/ProvidersTab";
import { useSearch, useNavigate } from "@tanstack/react-router";
import { GearIcon } from "@phosphor-icons/react";
import NotificationsTab from "./NotificationsTab";
import SecurityTab from "./SecurityTab";

export type SettingsTab =
  | "profile"
  | "providers"
  | "security"
  | "connected-accounts"
  | "authorized-apps"
  | "notifications"
  | "appearance";

export const SettingsPage: React.FC = () => {
  const search = useSearch({ from: "/settings" });
  const navigate = useNavigate();

  const activeTab: SettingsTab = (search?.tab as SettingsTab) || "profile";

  const handleTabChange = (tab: string) => {
    navigate({
      to: "/settings",
      search: { tab: tab as SettingsTab },
    });
  };

  return (
    <div className="flex-1 overflow-y-auto h-full px-5 py-8">
      <div className="max-w-3xl mx-auto">
        {activeTab === "profile" && <ProfileTab />}
        {activeTab === "providers" && <ProvidersTab />}
        {activeTab === "notifications" && <NotificationsTab />}
        {activeTab === "security" && <SecurityTab />}

        {activeTab !== "profile" &&
          activeTab !== "providers" &&
          activeTab !== "notifications" &&
          activeTab !== "security" && (
            <div className="flex flex-col items-center justify-center py-24 text-center border border-zinc-800/80 rounded-2xl bg-zinc-900/30 p-8 mt-6">
              <div className="w-12 h-12 rounded-xl bg-zinc-800/70 border border-zinc-700/60 flex items-center justify-center mb-4 text-zinc-300 shadow-sm">
                <GearIcon className="w-6 h-6" />
              </div>
              <h3 className="text-base font-semibold text-zinc-200 capitalize tracking-wide">
                {activeTab.replace(/-/g, " ")}
              </h3>
              <p className="text-xs text-zinc-500 mt-1.5 max-w-sm">
                Configuration and management for {activeTab.replace(/-/g, " ")} will be available
                soon.
              </p>
            </div>
          )}
      </div>
    </div>
  );
};

export default SettingsPage;
