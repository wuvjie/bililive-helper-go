import { ref } from "vue";
import router from "@/router";

export interface SSELine {
  time: string;
  text: string;
  type: "info" | "success" | "error";
}

function classifyLine(text: string): SSELine["type"] {
  if (/❌|错误|失败|error|failed|fatal/i.test(text)) return "error";
  if (/✅|完成|成功|success|done|finished/i.test(text)) return "success";
  return "info";
}

export function useSSE() {
  const lines = ref<SSELine[]>([]);
  const isRunning = ref(false);
  const error = ref<string | null>(null);
  let abortController: AbortController | null = null;

  function addLine(text: string) {
    lines.value.push({
      time: new Date().toLocaleTimeString("zh-CN"),
      text,
      type: classifyLine(text),
    });
  }

  function clear() {
    lines.value = [];
    error.value = null;
  }

  /** Last SSE call parameters, cached so callers can implement retry. */
  const lastRequest = ref<{ url: string; body?: Record<string, unknown> } | null>(null);

  async function startSSE(url: string, body?: Record<string, unknown>) {
    abort(); // cancel any in-flight request
    clear();
    isRunning.value = true;
    lastRequest.value = { url, body };
    abortController = new AbortController();

    try {
      const res = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: body ? JSON.stringify(body) : undefined,
        signal: abortController.signal,
        credentials: "same-origin", // required for session cookie auth
      });

      if (!res.ok) {
        if (res.status === 401) {
          router.push("/login");
          return;
        }
        const text = await res.text();
        throw new Error(text || `HTTP ${res.status}`);
      }

      if (!res.body) {
        throw new Error("响应体为空，无法读取 SSE 流");
      }

      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines_raw = buffer.split("\n");
        buffer = lines_raw.pop() || "";

        for (const line of lines_raw) {
          const trimmed = line.trim();
          if (!trimmed || !trimmed.startsWith("data:")) continue;
          const data = trimmed.slice(5).trim();
          if (data === "[END]") {
            isRunning.value = false;
            return;
          }
          addLine(data);
        }
      }
    } catch (e: any) {
      if (e.name === "AbortError") return;

      const msg = (e.message || "").toLowerCase();
      const isNetworkError =
        msg.includes("failed to fetch") ||
        msg.includes("networkerror") ||
        msg.includes("network request failed") ||
        msg.includes("network") ||
        msg.includes("load failed") ||
        msg.includes("err_internet_disconnected") ||
        msg.includes("err_name_not_resolved");

      const friendlyMsg = isNetworkError ? "网络连接中断，请检查网络后重试" : e.message;

      error.value = friendlyMsg;
      addLine(`❌ ${friendlyMsg}`);
    } finally {
      isRunning.value = false;
    }
  }

  function retryLast() {
    if (!lastRequest.value) return;
    const { url, body } = lastRequest.value;
    startSSE(url, body);
  }

  function abort() {
    abortController?.abort();
    isRunning.value = false;
  }

  return { lines, isRunning, error, lastRequest, addLine, clear, startSSE, abort, retryLast };
}
