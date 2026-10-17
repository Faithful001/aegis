import React, { createContext, useContext, useState, useEffect } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ChatThread, ChatMessage, DiscoveredModel, ProviderMeta } from "../types/api";
import { credentialsApi } from "../api/credentials";
import { threadsApi } from "../api/threads";
import { useAuth } from "./AuthContext";

interface AegisContextType {
  supportedProviders: ProviderMeta[];
  refreshProviders: () => Promise<void>;
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
  createNewThread: (title?: string, model?: string) => Promise<string>;
  deleteThread: (id: string) => Promise<void>;
  addMessageToThread: (threadId: string, msg: ChatMessage) => void;
  updateLastMessageInThread: (threadId: string, content: string) => void;
  persistMessage: (
    threadId: string,
    role: "system" | "user" | "assistant",
    content: string,
    model?: string
  ) => Promise<ChatMessage | null>;
  refetchThreads: () => Promise<void>;
}

const AegisContext = createContext<AegisContextType | undefined>(undefined);

export const AegisProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  const queryClient = useQueryClient();
  const [selectedModel, setSelectedModel] = useState<string>("");

  const [chatThreads, setChatThreads] = useState<ChatThread[]>([]);
  const [activeThreadId, setActiveThreadId] = useState<string | null>(null);

  // TanStack Query for threads
  const {
    data: rawThreads = [],
    isLoading: isLoadingThreads,
    refetch: refetchThreadsQuery,
  } = useQuery<ChatThread[]>({
    queryKey: ["threads"],
    queryFn: async () => {
      const res = await threadsApi.list();
      const list = Array.isArray(res?.data)
        ? res.data
        : Array.isArray(res)
        ? res
        : [];
      return list.map((t: ChatThread) => ({
        ...t,
        messages: Array.isArray(t.messages) ? t.messages : [],
      }));
    },
    enabled: isAuthenticated,
    staleTime: 1000 * 30,
  });

  // Keep chatThreads in sync with fetched threads from backend
  useEffect(() => {
    if (rawThreads.length > 0) {
      setChatThreads(rawThreads);
      setActiveThreadId((prev) => {
        if (prev && rawThreads.some((t) => t.id === prev)) {
          return prev;
        }
        return rawThreads[0]?.id || null;
      });
    } else if (!isAuthenticated) {
      setChatThreads([]);
      setActiveThreadId(null);
    } else {
      setChatThreads([]);
      setActiveThreadId(null);
    }
  }, [rawThreads, isAuthenticated]);

  // TanStack Query for supported providers metadata (carries is_configured per user)
  const { data: rawSupportedProviders = [], refetch: refetchProvidersQuery } = useQuery<
    ProviderMeta[]
  >({
    queryKey: ["providers", "supported"],
    queryFn: async () => {
      const list = await credentialsApi.getProviders();
      return Array.isArray(list) ? list : [];
    },
    enabled: isAuthenticated,
    staleTime: 0, // always fresh after credential changes
  });

  const supportedProviders = Array.isArray(rawSupportedProviders) ? rawSupportedProviders : [];

  // TanStack Query for models (retrieves active models from DB level)
  const {
    data: rawModels = [],
    isLoading: isLoadingModels,
    refetch: refetchModelsQuery,
  } = useQuery<DiscoveredModel[]>({
    queryKey: ["credentials", "models", "active"],
    queryFn: async () => {
      const fetched = await credentialsApi.listAllModels(true);
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

  const refreshProviders = async () => {
    await refetchProvidersQuery();
  };

  const refreshCredentials = async () => {
    await refetchProvidersQuery();
    await refetchModelsQuery();
  };

  const refreshModels = async () => {
    await refetchModelsQuery();
  };

  const refetchThreads = async () => {
    await refetchThreadsQuery();
  };

  const hasCredential = (provider: string): boolean => {
    return supportedProviders.some(
      (p) => p.id.toLowerCase() === provider.toLowerCase() && p.is_configured
    );
  };

  const createNewThread = async (title?: string, model?: string): Promise<string> => {
    const threadTitle = title || "New Exploration";
    const threadModel = model || selectedModel;

    try {
      const response = await threadsApi.create({
        title: threadTitle,
        model: threadModel,
      });

      const created = response.data;

      if (!created) {
        throw new Error("Failed to create thread: " + response.message);
      }

      const newThread: ChatThread = {
        ...created,
        messages: created.messages || [],
      };
      console.log("created thread", created);
      console.log("created thread id", created.id);
      console.log("new thread", newThread);

      setChatThreads((prev) => [newThread, ...prev.filter((t) => t.id !== newThread.id)]);
      setActiveThreadId(newThread.id);
      queryClient.setQueryData<ChatThread[]>(["threads"], (prev = []) => [
        newThread,
        ...prev.filter((t) => t.id !== newThread.id),
      ]);
      return newThread.id;
    } catch (err) {
      // Fallback local thread
      const newId = "thread-" + Date.now();
      const newThread: ChatThread = {
        id: newId,
        title: threadTitle,
        model: threadModel,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        messages: [],
      };
      setChatThreads((prev) => [newThread, ...prev]);
      setActiveThreadId(newId);
      return newId;
    }
  };

  const deleteThread = async (id: string): Promise<void> => {
    try {
      await threadsApi.delete(id);
    } catch (err) {
      console.error("Failed to delete thread on server:", err);
    }

    setChatThreads((prev) => {
      const filtered = prev.filter((t) => t.id !== id);
      if (activeThreadId === id) {
        setActiveThreadId(filtered.length > 0 ? filtered[0].id : null);
      }
      return filtered;
    });

    queryClient.setQueryData<ChatThread[]>(["threads"], (prev = []) =>
      prev.filter((t) => t.id !== id)
    );
  };

  const addMessageToThread = (threadId: string, msg: ChatMessage) => {
    setChatThreads((prev) =>
      prev.map((t) => {
        if (t.id === threadId) {
          const updatedMessages = [...(t.messages || []), msg];
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
        if (t.id === threadId && t.messages && t.messages.length > 0) {
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

  const persistMessage = async (
    threadId: string,
    role: "system" | "user" | "assistant",
    content: string,
    model?: string
  ): Promise<ChatMessage | null> => {
    try {
      const response = await threadsApi.addMessage(threadId, {
        role,
        content,
        model,
      });

      const saved = response.data;

      if (!saved) {
        throw new Error("Failed to persist message to backend thread: " + response.message);
      }

      // Update thread title and timestamp in background query cache
      refetchThreadsQuery();
      return saved;
    } catch (err) {
      console.error("Failed to persist message to backend thread:", err);
      return null;
    }
  };

  return (
    <AegisContext.Provider
      value={{
        supportedProviders,
        refreshProviders,
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
        deleteThread,
        addMessageToThread,
        updateLastMessageInThread,
        persistMessage,
        refetchThreads,
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
