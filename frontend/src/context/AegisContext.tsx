import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import { ProviderCredential, ChatThread, ChatMessage, DiscoveredModel } from "../types/api";
import { credentialsApi } from "../api/credentials";
import { useAuth } from "./AuthContext";

interface AegisContextType {
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
  const [credentials, setCredentials] = useState<ProviderCredential[]>([]);
  const [models, setModels] = useState<DiscoveredModel[]>([]);
  const [isLoadingModels, setIsLoadingModels] = useState<boolean>(false);
  const [selectedModel, setSelectedModel] = useState<string>("");

  const [chatThreads, setChatThreads] = useState<ChatThread[]>([]);
  const [activeThreadId, setActiveThreadId] = useState<string | null>("default-thread-1");

  const refreshModels = useCallback(async () => {
    if (!isAuthenticated) {
      setModels([]);
      setSelectedModel("");
      return;
    }
    setIsLoadingModels(true);
    try {
      const fetched = await credentialsApi.listModels();
      const list = fetched || [];
      setModels(list);
      setSelectedModel((prev) => {
        if (prev && list.some((m) => m.id === prev)) {
          return prev;
        }
        return list[0]?.id || "";
      });
    } catch (e) {
      console.error("Error discovering models from providers:", e);
      setModels([]);
      setSelectedModel("");
    } finally {
      setIsLoadingModels(false);
    }
  }, [isAuthenticated]);

  const refreshCredentials = async () => {
    if (!isAuthenticated) {
      setCredentials([]);
      return;
    }
    try {
      const list = await credentialsApi.list();
      setCredentials(list);
      // Automatically refresh models whenever credentials change
      refreshModels();
    } catch (e) {
      console.error("Error fetching user credentials:", e);
    }
  };

  const hasCredential = (provider: string): boolean => {
    return credentials.some((c) => c.provider.toLowerCase() === provider.toLowerCase());
  };

  useEffect(() => {
    if (isAuthenticated) {
      refreshCredentials();
      refreshModels();
    } else {
      setCredentials([]);
      setModels([]);
      setSelectedModel("");
    }
  }, [isAuthenticated, refreshModels]);

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
