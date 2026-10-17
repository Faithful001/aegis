import React, { useEffect, useState } from "react";
import { MainLayout } from "../../../components/layout/MainLayout";
import { ChatPromptInput } from "../../../components/chat/ChatPromptInput";
import { ChatMessages } from "../../../components/chat/ChatMessages";
import { useAegis } from "../../../context/AegisContext";
import { useAuth } from "../../../context/AuthContext";
import { AddCredentialModal } from "../../../components/modals/AddCredentialModal";
import { AuthModal } from "../../../components/modals/AuthModal";
import { toast } from "sonner";

interface Props {
  id?: string;
}

export const DynamicChatThreadPage: React.FC<Props> = ({ id }) => {
  const { isAuthenticated } = useAuth();
  const {
    chatThreads,
    activeThreadId,
    setActiveThreadId,
    fetchThreadById,
    selectedModel,
    models,
    sendMessage,
    isStreaming,
  } = useAegis();

  const [credModalOpen, setCredModalOpen] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);

  const targetId = id || activeThreadId || "";

  useEffect(() => {
    if (!id) return;
    setActiveThreadId(id);

    if (isAuthenticated) {
      fetchThreadById(id);
    }
  }, [id, isAuthenticated, setActiveThreadId, fetchThreadById]);

  const activeThread = chatThreads.find((t) => t.id === targetId);
  const messages = activeThread?.messages || [];

  const modelList = Array.isArray(models) ? models : [];
  const selectedModelMeta =
    modelList.find((m) => m.id === selectedModel) || (modelList.length > 0 ? modelList[0] : null);

  const handleSendMessage = async (promptText: string) => {
    if (!isAuthenticated) {
      toast.info("Please sign in to continue");
      setAuthModalOpen(true);
      return;
    }

    await sendMessage(promptText, targetId);
  };

  return (
    <MainLayout>
      <div className="flex flex-col h-full justify-between pb-6">
        <div className="flex-1 overflow-y-auto">
          <ChatMessages messages={messages} isStreaming={isStreaming} />
        </div>

        <div className="pt-4 bg-gradient-to-t from-background via-background to-transparent sticky bottom-0 z-10">
          <ChatPromptInput onSendMessage={handleSendMessage} disabled={isStreaming} />
        </div>
      </div>

      <AddCredentialModal
        open={credModalOpen}
        onOpenChange={setCredModalOpen}
        defaultProvider={selectedModelMeta?.provider}
      />
      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} initialMode="login" />
    </MainLayout>
  );
};

export default DynamicChatThreadPage;
