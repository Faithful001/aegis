import React, { createContext, useContext, useState, useEffect } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ProviderCredential, ChatThread, ChatMessage, DiscoveredModel, ProviderMeta } from "../types/api";
import { credentialsApi } from "../api/credentials";
import { useAuth } from "./AuthContext";

interface AegisContextType {
  supportedProviders: ProviderMeta[];
  credentials: ProviderCredential[];
  refreshCredentials: () => Promise<void>;
  hasCredential: (provider: string) => boolean;
  models: DiscoveredModel[];
  refreshModels: () => Promise<void>;
  isLoadingModels: boolean;
  selectedModel: string;
  setSelectedModel: (model: string) => void;
  chatThreads: ChatThread[];
  activeThreadId: string | null;
  setActiveThreadId: (id: string | null) => void;
  createNewThread: () => string;
  addMessageToThread: (threadId: string, msg: ChatMessage) => void;
  updateLastMessageInThread: (threadId: string, content: string) => void;
}

const AegisContext = createContext<AegisContextType | undefined>(undefined);

export const AegisProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  const queryClient = useQueryClient();
  const [selectedModel, setSelectedModel] = useState<string>("");

  const [chatThreads, setChatThreads] = useState<ChatThread[]>([]);
  const [activeThreadId, setActiveThreadId] = useState<string | null>("default-thread-1");

  // TanStack Query for supported providers metadata
  const {
    data: rawSupportedProviders = [],
  } = useQuery<ProviderMeta[]>({
    queryKey: ["providers", "supported"],
    queryFn: async () => {
      const list = await credentialsApi.getProviders();
      return Array.isArray(list) ? list : [];
    },
    staleTime: 1000 * 60 * 30, // 30 minutes
  });

  const supportedProviders = Array.isArray(rawSupportedProviders) ? rawSupportedProviders : [];

  // TanStack Query for credentials
  const {
    data: rawCredentials = [],
    refetch: refetchCredentialsQuery,
  } = useQuery<ProviderCredential[]>({
    queryKey: ["credentials", "list"],
    queryFn: async () => {
      const list = await credentialsApi.list();
      return Array.isArray(list) ? list : [];
    },
    enabled: isAuthenticated,
    staleTime: 1000 * 60,
  });

  const credentials = Array.isArray(rawCredentials) ? rawCredentials : [];

  // TanStack Query for models
  const {
    data: rawModels = [],
    isLoading: isLoadingModels,
    refetch: refetchModelsQuery,
  } = useQuery<DiscoveredModel[]>({
    queryKey: ["credentials", "models"],
    queryFn: async () => {
      const fetched = await credentialsApi.listModels();
      return Array.isArray(fetched) ? fetched : [];
    },
    enabled: isAuthenticated,
    staleTime: 1000 * 60,
  });

  const models = Array.isArray(rawModels) ? rawModels : [];

  // Automatically adjust selectedModel when models update
  useEffect(() => {
    if (models.length > 0) {
      setSelectedModel((prev) => {
        if (prev && models.some((m) => m.id === prev)) {
          return prev;
        }
        return models[0]?.id || "";
      });
    } else if (!isAuthenticated) {
      setSelectedModel("");
    }
  }, [models, isAuthenticated]);

  const refreshCredentials = async () => {
    await refetchCredentialsQuery();
    await refetchModelsQuery();
  };

  const refreshModels = async () => {
    await refetchModelsQuery();
  };

  const hasCredential = (provider: string): boolean => {
    return credentials.some((c) => c.provider.toLowerCase() === provider.toLowerCase());
  };

  const createNewThread = (): string => {
    const newId = "thread-" + Date.now();
    const newThread: ChatThread = {
      id: newId,
      title: "New Exploration",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      messages: [],
    };
    setChatThreads((prev) => [newThread, ...prev]);
    setActiveThreadId(newId);
    return newId;
  };

  const addMessageToThread = (threadId: string, msg: ChatMessage) => {
    setChatThreads((prev) =>
      prev.map((t) => {
        if (t.id === threadId) {
          const updatedMessages = [...t.messages, msg];
          let title = t.title;
          if (t.title === "New Exploration" && msg.role === "user") {
            title = msg.content.slice(0, 30) + (msg.content.length > 30 ? "..." : "");
          }
          return {
            ...t,
            title,
            updated_at: new Date().toISOString(),
            messages: updatedMessages,
          };
        }
        return t;
      })
    );
  };

  const updateLastMessageInThread = (threadId: string, content: string) => {
    setChatThreads((prev) =>
      prev.map((t) => {
        if (t.id === threadId && t.messages.length > 0) {
          const msgs = [...t.messages];
          const lastIndex = msgs.length - 1;
          msgs[lastIndex] = {
            ...msgs[lastIndex],
            content: msgs[lastIndex].content + content,
          };
          return {
            ...t,
            updated_at: new Date().toISOString(),
            messages: msgs,
          };
        }
        return t;
      })
    );
  };

  return (
    <AegisContext.Provider
      value={{
        supportedProviders,
        credentials,
        refreshCredentials,
        hasCredential,
        models,
        refreshModels,
        isLoadingModels,
        selectedModel,
        setSelectedModel,
        chatThreads,
        activeThreadId,
        setActiveThreadId,
        createNewThread,
        addMessageToThread,
        updateLastMessageInThread,
      }}
    >
      {children}
    </AegisContext.Provider>
  );
};

export const useAegis = () => {
  const ctx = useContext(AegisContext);
  if (!ctx) throw new Error("useAegis must be used within AegisProvider");
  return ctx;
};
