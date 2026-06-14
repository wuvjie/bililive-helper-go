import { onMounted, onActivated } from "vue";

/**
 * 统一处理 onMounted + onActivated 的数据加载。
 * 解决 keep-alive 场景下每个 view 都要重复写两遍相同 fetch 逻辑的问题。
 *
 * @param fetchFn - 数据加载函数
 * @param opts.onMountedOnly - 仅在 onMounted 时执行（如 setupCheck 只需首次加载）
 */
export function useFetchOnActivate(
  fetchFn: () => Promise<void>,
  opts?: { onMountedOnly?: boolean }
) {
  onMounted(fetchFn);
  if (!opts?.onMountedOnly) {
    onActivated(fetchFn);
  }
}
