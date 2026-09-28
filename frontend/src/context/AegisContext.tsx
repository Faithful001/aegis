import React, { createContext, useContext, useState, useEffect } from "react";
import { Organization, Project, APIKey, ChatThread, ChatMessage } from "../types/api";
import { orgsApi } from "../api/orgs";
import { projectsApi } from "../api/projects";
import { useAuth } from "./AuthContext";

interface AegisContextType {
  organizations: Organization[];
  activeOrg: Organization | null;
  setActiveOrg: (org: Organization | null) => void;
  projects: Project[];
  activeProject: Project | null;
  setActiveProject: (proj: Project | null) => void;
  apiKeys: APIKey[];
  activeApiKey: string;
  setActiveApiKey: (key: string) => void;
  selectedModel: string;
  setSelectedModel: (model: string) => void;
  chatThreads: ChatThread[];
  activeThreadId: string | null;
  setActiveThreadId: (id: string | null) => void;
  createNewThread: () => string;
  addMessageToThread: (threadId: string, msg: ChatMessage) => void;
  updateLastMessageInThread: (threadId: string, content: string) => void;
  refreshOrgs: () => Promise<void>;
  refreshProjects: () => Promise<void>;
  refreshKeys: () => Promise<void>;
}

const AegisContext = createContext<AegisContextType | undefined>(undefined);

export const AVAILABLE_MODELS = [
  {
    id: "aegis-mistral-7b",
    name: "Aegis Mistral 7B",
    description: "Distributed self-hosted AI worker",
    fast: true,
  },
  {
    id: "openrouter/auto",
    name: "OpenRouter Auto",
    description: "BYOK high performance router",
    fast: true,
  },
  {
    id: "gpt-4o-mini",
    name: "OpenAI GPT-4o Mini",
    description: "Direct BYOK provider route",
    fast: true,
  },
  {
    id: "claude-3-5-sonnet",
    name: "Claude 3.5 Sonnet",
    description: "Reasoning and code generation",
    fast: false,
  },
];

export const AegisProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [activeOrg, setActiveOrg] = useState<Organization | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProject, setActiveProject] = useState<Project | null>(null);
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [activeApiKey, setActiveApiKey] = useState<string>(
    localStorage.getItem("aegis_user_api_key") || ""
  );
  const [selectedModel, setSelectedModel] = useState<string>("aegis-mistral-7b");

  const [chatThreads, setChatThreads] = useState<ChatThread[]>([]);
  const [activeThreadId, setActiveThreadId] = useState<string | null>("default-thread-1");

  const refreshOrgs = async () => {
    if (!isAuthenticated) return;
    try {
      const list = await orgsApi.list();
      setOrganizations(list);
      if (list.length > 0 && !activeOrg) {
        setActiveOrg(list[0]);
      }
    } catch (e) {
      console.error("Error fetching orgs:", e);
    }
  };

  const refreshProjects = async () => {
    if (!activeOrg) return;
    try {
      const list = await projectsApi.listByOrg(activeOrg.id);
      setProjects(list);
      if (list.length > 0 && (!activeProject || activeProject.organization_id !== activeOrg.id)) {
        setActiveProject(list[0]);
      }
    } catch (e) {
      console.error("Error fetching projects:", e);
    }
  };

  const refreshKeys = async () => {
    if (!activeProject) return;
    try {
      const keys = await projectsApi.listApiKeys(activeProject.id);
      setApiKeys(keys);
    } catch (e) {
      console.error("Error fetching API keys:", e);
    }
  };

  useEffect(() => {
    if (isAuthenticated) refreshOrgs();
  }, [isAuthenticated]);

  useEffect(() => {
    if (activeOrg) refreshProjects();
  }, [activeOrg]);

  useEffect(() => {
    if (activeProject) refreshKeys();
  }, [activeProject]);

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
        organizations,
        activeOrg,
        setActiveOrg,
        projects,
        activeProject,
        setActiveProject,
        apiKeys,
        activeApiKey,
        setActiveApiKey: (k) => {
          localStorage.setItem("aegis_user_api_key", k);
          setActiveApiKey(k);
        },
        selectedModel,
        setSelectedModel,
        chatThreads,
        activeThreadId,
        setActiveThreadId,
        createNewThread,
        addMessageToThread,
        updateLastMessageInThread,
        refreshOrgs,
        refreshProjects,
        refreshKeys,
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
