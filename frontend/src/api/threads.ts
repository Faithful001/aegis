import { apiClient } from "./client";
import { ChatThread, ChatMessage } from "../types/api";
import { BaseResponse } from ".";

export interface CreateThreadPayload {
  title?: string;
  model?: string;
}

export interface UpdateThreadPayload {
  title?: string;
}

export interface AddMessagePayload {
  role: "system" | "user" | "assistant";
  content: string;
  model?: string;
}

export const threadsApi = {
  list: async (): Promise<BaseResponse<ChatThread[]>> => {
    const res = await apiClient.get<BaseResponse<ChatThread[]>>("/threads");
    return res.data;
  },

  get: async (id: string): Promise<BaseResponse<ChatThread>> => {
    const res = await apiClient.get<BaseResponse<ChatThread>>(`/threads/${id}`);
    return res.data;
  },

  create: async (payload: CreateThreadPayload = {}): Promise<BaseResponse<ChatThread>> => {
    const res = await apiClient.post<BaseResponse<ChatThread>>("/threads", payload);
    return res.data;
  },

  update: async (id: string, payload: UpdateThreadPayload): Promise<BaseResponse<ChatThread>> => {
    const res = await apiClient.patch<BaseResponse<ChatThread>>(`/threads/${id}`, payload);
    return res.data;
  },

  delete: async (id: string): Promise<BaseResponse<void>> => {
    const res = await apiClient.delete<BaseResponse<void>>(`/threads/${id}`);
    return res.data;
  },

  listMessages: async (threadId: string): Promise<BaseResponse<ChatMessage[]>> => {
    const res = await apiClient.get<BaseResponse<ChatMessage[]>>(`/threads/${threadId}/messages`);
    return res.data;
  },

  addMessage: async (
    threadId: string,
    payload: AddMessagePayload
  ): Promise<BaseResponse<ChatMessage>> => {
    const res = await apiClient.post<BaseResponse<ChatMessage>>(
      `/threads/${threadId}/messages`,
      payload
    );
    return res.data;
  },
};
