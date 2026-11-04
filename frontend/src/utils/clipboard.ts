import { ElMessage } from "element-plus";

/**
 * Universal copy helper that supports both modern navigator.clipboard and legacy fallback
 * (Essential for accessing via HTTP IP addresses on LAN/NAS without HTTPS).
 */
export async function copyToClipboard(text: string, successMsg = "已复制到剪贴板"): Promise<boolean> {
  if (!text) return false;
  let success = false;

  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text);
      success = true;
    } catch {
      // Fallback below
    }
  }

  if (!success) {
    try {
      const textArea = document.createElement("textarea");
      textArea.value = text;
      textArea.style.position = "fixed";
      textArea.style.top = "-9999px";
      textArea.style.left = "-9999px";
      textArea.style.opacity = "0";
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      success = document.execCommand("copy");
      document.body.removeChild(textArea);
    } catch {
      success = false;
    }
  }

  if (success) {
    ElMessage.success(successMsg);
  } else {
    ElMessage.error("复制失败，请手动选中文本复制");
  }

  return success;
}
