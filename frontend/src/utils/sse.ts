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
      type: classifyLine(text)
    });
  }

  function clear() {
    lines.value = [];
    error.value = null;
  }

  async function startSSE(url: string, body?: Record<string, any>) {
    abort(); // cancel any in-flight request
    clear();
    isRunning.value = true;
    abortController = new AbortController();

    try {
      const res = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: body ? JSON.stringify(body) : undefined,
        signal: abortController.signal,
        credentials: "same-origin" // required for session cookie auth
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
      error.value = e.message;
      addLine(`❌ 错误: ${e.message}`);
    } finally {
      isRunning.value = false;
    }
  }

  function abort() {
    abortController?.abort();
    isRunning.value = false;
  }

  return { lines, isRunning, error, addLine, clear, startSSE, abort };
}
