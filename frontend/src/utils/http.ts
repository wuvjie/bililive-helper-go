import axios from "axios";
import type { AxiosInstance, AxiosRequestConfig } from "axios";
import { ElMessage } from "element-plus";
import router from "@/router";
import { markUnauthenticated } from "@/router";

const http: AxiosInstance = axios.create({
  baseURL: "/api",
  timeout: 15000,
  withCredentials: true // send session cookie with every request
});

http.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      markUnauthenticated(); // 重置路由守卫状态，防止 auth 循环
      router.push("/login");
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
