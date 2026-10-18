import React, { useEffect, useState } from "react";
import { userApi } from "@/api/user";
import { useAuth } from "@/context/AuthContext";
import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import SettingsLayout from "@/components/layout/SettingsLayout";

interface NotificationItem {
  id: string;
  title: string;
  description: string;
  enabled: boolean;
  toggle: () => void;
  expandableContent?: React.ReactNode;
}

const NotificationsTab: React.FC = () => {
  const { user } = useAuth();

  // Budget Alerts state
  const [budgetEnabled, setBudgetEnabled] = useState(true);
  const [budgetUSD, setBudgetUSD] = useState("50.00");
  const [threshold, setThreshold] = useState("80");
  // const [alertEmail, setAlertEmail] = useState("");

  // Other notification toggles matching modern platform alerts
  const [failedDeployments, setFailedDeployments] = useState(true);
  const [successfulDeployments, setSuccessfulDeployments] = useState(false);
  const [rateLimitAlerts, setRateLimitAlerts] = useState(true);
  const [keyRevokedAlerts, setKeyRevokedAlerts] = useState(false);

  const { data: alertSettings, refetch: refetchAlerts } = useQuery({
    queryKey: ["user", "budget-alerts"],
    queryFn: userApi.getAlertSettings,
  });

  useEffect(() => {
    if (alertSettings) {
      setBudgetEnabled(alertSettings.budget_alert_enabled);
      setBudgetUSD(
        alertSettings.monthly_budget_usd ? alertSettings.monthly_budget_usd.toString() : "50.00"
      );
      setThreshold(
        alertSettings.threshold_percentage ? alertSettings.threshold_percentage.toString() : "80"
      );
      // setAlertEmail(alertSettings.alert_email || user?.email || "");
    }
    // else if (user?.email && !alertEmail) {
    //   // setAlertEmail(user.email);
    // }
  }, [alertSettings, user]);

  const updateAlertsMutation = useMutation({
    mutationFn: async () =>
      userApi.updateAlertSettings({
        budget_alert_enabled: budgetEnabled,
        monthly_budget_usd: parseFloat(budgetUSD) || 50,
        threshold_percentage: parseInt(threshold) || 80,
        alert_email: user?.email,
      }),
    onSuccess: () => {
      toast.success("Notification preferences saved successfully!");
      refetchAlerts();
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to update notification settings");
    },
  });

  const notificationSections: NotificationItem[] = [
    {
      id: "budget-alerts",
      title: "Usage & Budget alerts",
      description: "When token expenditure approaches or reaches your monthly spend threshold.",
      enabled: budgetEnabled,
      toggle: () => setBudgetEnabled(!budgetEnabled),
      expandableContent: budgetEnabled ? (
        <div className="pt-4 mt-3 border-t border-zinc-800/60">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <Input
              label="Monthly Budget ($ USD)"
              type="number"
              min="1"
              step="5"
              placeholder="50.00"
              className="rounded-lg bg-zinc-900/60 border-zinc-800 text-sm focus:border-zinc-600"
              value={budgetUSD}
              onChange={(e) => setBudgetUSD(e.target.value)}
            />
            <Input
              label="Alert Threshold (%)"
              type="number"
              min="10"
              max="100"
              step="5"
              placeholder="80"
              className="rounded-lg bg-zinc-900/60 border-zinc-800 text-sm focus:border-zinc-600"
              value={threshold}
              onChange={(e) => setThreshold(e.target.value)}
            />
          </div>
        </div>
      ) : null,
    },
    {
      id: "rate-limit-alerts",
      title: "Rate limit warnings",
      description:
        "When your API calls encounter provider rate limits (HTTP 429) or token capacity caps.",
      enabled: rateLimitAlerts,
      toggle: () => setRateLimitAlerts(!rateLimitAlerts),
    },
    {
      id: "failed-requests",
      title: "Failed requests & model outages",
      description:
        "When a connected model provider returns persistent 5xx errors or connection timeouts.",
      enabled: failedDeployments,
      toggle: () => setFailedDeployments(!failedDeployments),
    },
    {
      id: "key-revocation",
      title: "Credential status changes",
      description: "When an active API key is invalidated, expired, or modified in the vault.",
      enabled: keyRevokedAlerts,
      toggle: () => setKeyRevokedAlerts(!keyRevokedAlerts),
    },
    {
      id: "weekly-digest",
      title: "Weekly usage digest",
      description:
        "Receive a summary of token consumption, model breakdown, and cost estimates every Monday.",
      enabled: successfulDeployments,
      toggle: () => setSuccessfulDeployments(!successfulDeployments),
    },
  ];

  return (
    <SettingsLayout
      title="Notifications"
      subTitle="Manage how and when you receive alerts about your models, usage, and credentials."
    >
      <div className="space-y-6">
        {/* Card with list rows matching the sample */}
        <Card
          hover={false}
          className="p-0 border border-zinc-800/80 bg-[#121214] rounded-xl overflow-hidden divide-y divide-zinc-800/70"
        >
          {notificationSections.map((item) => (
            <div key={item.id} className="px-6 py-5 transition-colors hover:bg-zinc-900/30">
              <div className="flex items-start justify-between gap-4">
                <div className="space-y-1 pr-4">
                  <h3 className="text-sm font-medium text-white tracking-tight">{item.title}</h3>
                  <p className="text-xs text-zinc-400 leading-relaxed">{item.description}</p>
                </div>

                {/* Maintained toggle switch */}
                <label className="relative inline-flex items-center cursor-pointer flex-shrink-0 mt-0.5">
                  <input
                    type="checkbox"
                    checked={item.enabled}
                    onChange={item.toggle}
                    className="sr-only peer"
                  />
                  <div className="w-9 h-5 bg-zinc-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-zinc-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-500"></div>
                </label>
              </div>

              {item.expandableContent}
            </div>
          ))}
        </Card>

        {/* Action Button container aligned to right */}
        <div className="flex justify-end pt-2">
          <Button
            onClick={() => updateAlertsMutation.mutate()}
            disabled={updateAlertsMutation.isPending}
            className="rounded-lg px-4 py-2 text-sm font-semibold bg-white text-zinc-950 hover:bg-zinc-200 transition-all shadow-sm"
          >
            {updateAlertsMutation.isPending ? "Saving changes..." : "Save changes"}
          </Button>
        </div>
      </div>
    </SettingsLayout>
  );
};

export default NotificationsTab;
