import axios from "axios";
import type { AxiosInstance, AxiosRequestConfig } from "axios";
import { ElMessage } from "element-plus";
import router from "@/router";
import { useAuthStore } from "@/store/modules/auth";

const http: AxiosInstance = axios.create({
  baseURL: "/api",
  timeout: 15000,
  withCredentials: true, // send session cookie with every request
});

http.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore().markUnauthenticated(); // 重置认证状态，防止 auth 循环
      router.push("/login");
      return Promise.reject(error);
    }
    const data = error.response?.data;
    const msg = typeof data === "string" ? data : data?.message || data?.error || error.message;
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
