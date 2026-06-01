import axios from "axios";
import type { AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from "axios";
import { ElMessage } from "element-plus";

const http: AxiosInstance = axios.create({
  baseURL: "/api",
  timeout: 30000
});

http.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

http.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      window.location.href = "/login";
      return Promise.reject(error);
    }
    const msg = error.response?.data?.message || error.response?.data?.error || error.message;
    ElMessage.error(msg);
    return Promise.reject(error);
  }
);

export function get<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
  return http.get(url, { params, ...config });
}

export function post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
  return http.post(url, data, config);
}

export default http;
