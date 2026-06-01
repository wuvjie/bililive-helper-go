import { post, get } from "@/utils/http";

export function login(password: string) {
  return post("/login", { password });
}

export function logout() {
  window.location.href = "/logout";
}
