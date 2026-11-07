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
      textArea.style.left = "-9999px";
      textArea.style.top = "0";
      textArea.style.opacity = "0";
      textArea.setAttribute("readonly", "");

      // If inside an active el-drawer or el-dialog, mount inside it so focus trap doesn't break copy
      const activeModal = document.querySelector(".el-drawer:not([style*='display: none'])") ||
                          document.querySelector(".el-dialog:not([style*='display: none'])") ||
                          document.body;
      activeModal.appendChild(textArea);
      textArea.focus({ preventScroll: true });
      textArea.select();
      textArea.setSelectionRange(0, text.length);
      success = document.execCommand("copy");
      activeModal.removeChild(textArea);
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
