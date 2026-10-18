import { credentialsApi } from "@/api/credentials";
import { DiscoveredModel, PaginatedResult } from "@/types/api";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import {
  ArrowsClockwiseIcon,
  CaretLeftIcon,
  CaretRightIcon,
  MagnifyingGlassIcon,
} from "@phosphor-icons/react";
import { Badge } from "../ui/Badge";

const MODELS_PER_PAGE = 10;

export const ProviderModelDrawer: React.FC<{ providerId: string; providerName: string }> = ({
  providerId,
  providerName,
}) => {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [togglingId, setTogglingId] = useState<string | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1);
    }, 250);
    return () => clearTimeout(timer);
  }, [search]);

  const { data, isLoading, isFetching } = useQuery<PaginatedResult<DiscoveredModel>>({
    queryKey: ["provider-models", providerId, page, debouncedSearch],
    queryFn: () => credentialsApi.listModels(providerId, page, MODELS_PER_PAGE, debouncedSearch),
    staleTime: 1000 * 60 * 5,
  });

  const toggleMutation = useMutation({
    mutationFn: async ({ modelId, currentActive }: { modelId: string; currentActive: boolean }) => {
      setTogglingId(modelId);
      return credentialsApi.toggleModel(providerId, modelId, !currentActive);
    },
    onSuccess: (updatedModel) => {
      toast.success(
        `${updatedModel.name || updatedModel.id} is now ${updatedModel.is_active ? "active" : "inactive"}`
      );
      queryClient.invalidateQueries({ queryKey: ["provider-models", providerId] });
      queryClient.invalidateQueries({ queryKey: ["credentials", "models"] });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to toggle model status");
    },
    onSettled: () => {
      setTogglingId(null);
    },
  });

  const models = data?.items || [];
  const total = data?.total || 0;
  const totalPages = Math.max(1, data?.total_pages || 1);

  return (
    <div className="px-5 pb-4 pt-3 space-y-3">
      {/* Search */}
      <div className="relative">
        <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-zinc-500" />
        <input
          type="text"
          placeholder="Filter models…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-9 pr-8 py-1.5 text-xs bg-zinc-900 border border-zinc-700/60 text-zinc-200 rounded-lg placeholder:text-zinc-600 focus:outline-none focus:border-zinc-500 transition-all"
        />
        {/* {(isFetching || isLoading) && (
          <RefreshCw className="absolute right-3 top-1/2 -translate-y-1/2 w-3 h-3 text-zinc-500 animate-spin" /> 
        )} */}
      </div>

      {/* Model rows */}
      <>
        {isLoading ? (
          <div className="flex items-center gap-2 text-zinc-500 text-xs py-4 px-5">
            <ArrowsClockwiseIcon className="w-3.5 h-3.5 animate-spin" /> Loading models…
          </div>
        ) : total === 0 && !debouncedSearch ? (
          <p className="text-xs text-zinc-500 italic px-5 py-4">
            No models found for {providerName}.
          </p>
        ) : (
          <div className="space-y-0.5">
            {models.length === 0 ? (
              <p className="text-xs text-zinc-500 italic py-2">No models match your search.</p>
            ) : (
              models.map((model) => (
                <div
                  key={model.id}
                  className="flex items-center justify-between px-3 py-2 rounded-lg hover:bg-zinc-800/40 transition-colors"
                >
                  <div className="min-w-0">
                    <span className="text-xs font-medium text-zinc-200 truncate block">
                      {model.name}
                    </span>
                    <span className="text-[10px] text-zinc-500 font-mono truncate block">
                      {model.id}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 shrink-0 ml-3">
                    {model.fast && <Badge variant="info">Fast</Badge>}
                    <button
                      type="button"
                      role="switch"
                      aria-checked={model.is_active}
                      onClick={() =>
                        toggleMutation.mutate({
                          modelId: model.id,
                          currentActive: !!model.is_active,
                        })
                      }
                      disabled={togglingId === model.id}
                      className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full p-0.5 transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500/50 ${
                        model.is_active
                          ? "bg-emerald-500 shadow-sm shadow-emerald-500/30"
                          : "bg-zinc-700/80 hover:bg-zinc-600/80"
                      } ${togglingId === model.id ? "opacity-60 cursor-not-allowed" : ""}`}
                      title={
                        model.is_active
                          ? "Model is active. Click to deactivate"
                          : "Model is inactive. Click to activate"
                      }
                    >
                      <span className="sr-only">Toggle model active status</span>
                      <span
                        className={`pointer-events-none flex items-center justify-center h-4 w-4 transform rounded-full bg-white shadow-md transition-transform duration-200 ease-in-out ${
                          model.is_active ? "translate-x-4" : "translate-x-0"
                        }`}
                      >
                        {togglingId === model.id && (
                          <ArrowsClockwiseIcon className="w-2.5 h-2.5 text-zinc-600 animate-spin" />
                        )}
                      </span>
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between pt-1">
          <span className="text-[10px] text-zinc-500">
            {total} models · page {page}/{totalPages}
          </span>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1 || isFetching}
              className="p-1 rounded text-zinc-400 hover:text-zinc-100 disabled:opacity-30 hover:bg-zinc-800 transition-colors"
            >
              <CaretLeftIcon className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages || isFetching}
              className="p-1 rounded text-zinc-400 hover:text-zinc-100 disabled:opacity-30 hover:bg-zinc-800 transition-colors"
            >
              <CaretRightIcon className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
