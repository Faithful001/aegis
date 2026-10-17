import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";

export const API_BASE_URL = "/api/v1";

export const getAccessToken = (): string | null => sessionStorage.getItem("aegis_jwt_token");
export const getRefreshToken = (): string | null => sessionStorage.getItem("aegis_refresh_token");

export const setTokens = (accessToken: string, refreshToken?: string) => {
  sessionStorage.setItem("aegis_jwt_token", accessToken);
  if (refreshToken) {
    sessionStorage.setItem("aegis_refresh_token", refreshToken);
  }
};

export const clearTokens = () => {
  sessionStorage.removeItem("aegis_jwt_token");
  sessionStorage.removeItem("aegis_refresh_token");
};

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

apiClient.interceptors.request.use(
  (config) => {
    const token = getAccessToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Handling token refresh & concurrent queue
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (err: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else if (token) {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as
      | (InternalAxiosRequestConfig & { _retry?: boolean })
      | undefined;

    if (!originalRequest) {
      return Promise.reject(error);
    }

    const isAuthEndpoint = originalRequest.url?.includes("/auth/");

    if (error.response?.status === 401 && !originalRequest._retry && !isAuthEndpoint) {
      const refreshToken = getRefreshToken();

      if (!refreshToken) {
        clearTokens();
        window.dispatchEvent(new Event("unauthorized"));
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise<string>((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return apiClient(originalRequest);
          })
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Direct axios post to avoid interceptor recursion
        const res = await axios.post(`${API_BASE_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });

        const newTokenData = res.data?.data?.token;
        const newAccessToken = newTokenData?.access_token;
        const newRefreshToken = newTokenData?.refresh_token;

        if (!newAccessToken) {
          throw new Error("No access token returned from refresh endpoint");
        }

        setTokens(newAccessToken, newRefreshToken);

        window.dispatchEvent(
          new CustomEvent("aegis_token_refreshed", {
            detail: {
              accessToken: newAccessToken,
              refreshToken: newRefreshToken,
              user: res.data?.data?.user,
            },
          })
        );

        processQueue(null, newAccessToken);
        originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
        return apiClient(originalRequest);
      } catch (refreshErr) {
        processQueue(refreshErr, null);
        clearTokens();
        window.dispatchEvent(new Event("unauthorized"));
        return Promise.reject(refreshErr);
      } finally {
        isRefreshing = false;
      }
    }

    // If 401 on an auth endpoint or already retried once
    if (error.response?.status === 401 && (isAuthEndpoint || originalRequest._retry)) {
      clearTokens();
      window.dispatchEvent(new Event("unauthorized"));
    }

    return Promise.reject(error);
  }
);
