import React from "react";
import { authApi } from "@/api/auth";
import { credentialsApi } from "@/api/credentials";
import { userApi } from "@/api/user";
import { useAuth } from "@/context/AuthContext";
import { ProviderCredential, ProviderMeta, UserSession } from "@/types/api";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Card } from "../../components/ui/Card";
import { Bell } from "lucide-react";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";

const NotificationsTab = () => {
  const { user } = useAuth();

  const [budgetEnabled, setBudgetEnabled] = useState(true);
  const [budgetUSD, setBudgetUSD] = useState("50.00");
  const [threshold, setThreshold] = useState("80");
  const [alertEmail, setAlertEmail] = useState("");

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
      setAlertEmail(alertSettings.alert_email || user?.email || "");
    } else if (user?.email && !alertEmail) {
      setAlertEmail(user.email);
    }
  }, [alertSettings, user]);

  const updateAlertsMutation = useMutation({
    mutationFn: async () =>
      userApi.updateAlertSettings({
        budget_alert_enabled: budgetEnabled,
        monthly_budget_usd: parseFloat(budgetUSD) || 50,
        threshold_percentage: parseInt(threshold) || 80,
        alert_email: alertEmail,
      }),
    onSuccess: () => {
      toast.success("Budget alert preferences saved!");
      refetchAlerts();
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to update alert settings");
    },
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-white tracking-tight">Notifications</h1>
        <p className="text-xs text-zinc-400 mt-1">
          Receive automatic email warnings when token expenditure approaches your monthly limit
        </p>
      </div>

      <Card hover={false} className="p-5 space-y-4 !bg-sidebar">
        <div className="flex items-start justify-between">
          <div className="space-y-0.5">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <Bell className="w-4 h-4 text-amber-400" />
              Usage & Budget Alerts
            </h3>
            <p className="text-xs text-zinc-400">
              Receive automatic email warnings when token expenditure approaches your monthly limit
            </p>
          </div>
          <label className="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              checked={budgetEnabled}
              onChange={(e) => setBudgetEnabled(e.target.checked)}
              className="sr-only peer"
            />
            <div className="w-9 h-5 bg-zinc-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-zinc-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-500"></div>
          </label>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-1">
          <Input
            label="Monthly Budget ($ USD)"
            type="number"
            min="1"
            step="5"
            placeholder="50.00"
            className="rounded-full"
            value={budgetUSD}
            onChange={(e) => setBudgetUSD(e.target.value)}
            disabled={!budgetEnabled}
          />
          <Input
            label="Alert Threshold (%)"
            type="number"
            min="10"
            max="100"
            step="5"
            placeholder="80"
            className="rounded-full"
            value={threshold}
            onChange={(e) => setThreshold(e.target.value)}
            disabled={!budgetEnabled}
          />
          <Input
            label="Alert Email"
            type="email"
            placeholder="user@example.com"
            className="rounded-full"
            value={alertEmail}
            onChange={(e) => setAlertEmail(e.target.value)}
            disabled={!budgetEnabled}
          />
        </div>

        <div className="flex justify-end pt-1">
          <Button
            onClick={() => updateAlertsMutation.mutate()}
            disabled={updateAlertsMutation.isPending}
            className="rounded-full px-5 py-2 font-semibold text-xs bg-zinc-200 text-zinc-950 hover:bg-white"
          >
            {updateAlertsMutation.isPending ? "Saving..." : "Save Alert Settings"}
          </Button>
        </div>
      </Card>
    </div>
  );
};

export default NotificationsTab;
