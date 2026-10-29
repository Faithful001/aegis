import { apiClient } from "./client";

export interface UploadedFileResponse {
  id: string;
  filename: string;
  mime_type: string;
  size: number;
  file_type: "image" | "document";
  url: string;
  content?: string;
}

export const uploadApi = {
  upload: async (file: File, type: "photo" | "file"): Promise<UploadedFileResponse> => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("type", type);

    const response = await apiClient.post<{
      success: boolean;
      message: string;
      data: UploadedFileResponse;
    }>("/upload", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });

    return response.data.data;
  },

  uploadMultiple: async (files: File[], type: "photo" | "file"): Promise<UploadedFileResponse[]> => {
    const formData = new FormData();
    files.forEach((file) => {
      formData.append("files", file);
    });
    formData.append("type", type);

    const response = await apiClient.post<{
      success: boolean;
      message: string;
      data: UploadedFileResponse[];
    }>("/upload/multiple", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });

    return response.data.data;
  },

  getFile: async (id: string): Promise<UploadedFileResponse> => {
    const response = await apiClient.get<{
      success: boolean;
      data: UploadedFileResponse;
    }>(`/upload/${id}`);
    return response.data.data;
  },
};

