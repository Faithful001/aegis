import React, { useState, useEffect } from "react";
import { MainLayout } from "../components/layout/MainLayout";
import { ChatPromptInput } from "../components/chat/ChatPromptInput";
import { useAegis } from "../context/AegisContext";
import { useAuth } from "../context/AuthContext";
import { AddCredentialModal } from "../components/modals/AddCredentialModal";
import { AuthModal } from "../components/modals/AuthModal";
import { toast } from "sonner";
import { router } from "@/router";

export const HomePage: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const {
    selectedModel,
    models,
    createNewThread,
    sendMessage,
    isStreaming,
    setActiveThreadId,
  } = useAegis();

  const [credModalOpen, setCredModalOpen] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);

  useEffect(() => {
    setActiveThreadId(null);
  }, []);

  const modelList = Array.isArray(models) ? models : [];
  const selectedModelMeta =
    modelList.find((m) => m.id === selectedModel) || (modelList.length > 0 ? modelList[0] : null);

  const handleSendMessage = async (promptText: string) => {
    if (!isAuthenticated) {
      toast.info("Please sign in to continue");
      setAuthModalOpen(true);
      return;
    }

    const newThreadId = await createNewThread(promptText, selectedModel);
    router.navigate({ to: `/c/${newThreadId}` });
    sendMessage(promptText, newThreadId);
  };

  return (
    <MainLayout>
      <div className="flex flex-col h-full justify-center pb-6">
        <div className="flex-1 flex flex-col items-center justify-center text-center px-4 pt-12">
          <h1 className="text-3xl font-bold tracking-tight text-white mb-8">
            What should we explore?
          </h1>
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

export default HomePage;
