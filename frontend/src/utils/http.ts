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
    const url = error.config?.url || "";
    const isAuthOp = url.includes("/login") || url.includes("/auth/change-password");

    // 针对常规业务接口的 401：会话已过期，重置认证状态并重定向至登录页
    if (error.response?.status === 401 && !isAuthOp) {
      useAuthStore().markUnauthenticated(); // 重置认证状态，防止 auth 循环
      if (router.currentRoute.value.path !== "/login") {
        router.push("/login");
      }
      return Promise.reject(error);
    }
    const data = error.response?.data;
    const msg = typeof data === "string" ? data : data?.message || data?.error || error.message;
    ElMessage.error(msg || "操作失败");
    return Promise.reject(error);
  }
);

export function get<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
  return http.get(url, { params, ...config });
}

export function post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
  return http.post(url, data, config);
}
export function del<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
  return http.delete(url, { params, ...config });
}
