import { post } from "@/utils/http";

export function login(password: string) {
  return post("/login", { password });
}

export function logout() {
  localStorage.removeItem("token");
  window.location.href = "/logout";
}
