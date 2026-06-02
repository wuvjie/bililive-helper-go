import { post } from "@/utils/http";

export function login(password: string) {
  return post("/login", { password });
}

export function logout() {
  window.location.href = "/logout";
}

export function changePassword(old_password: string, new_password: string) {
  return post("/auth/change-password", { old_password, new_password });
}
