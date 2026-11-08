import { ref, type Ref, type ComputedRef, onMounted, onUnmounted, onActivated, onDeactivated } from "vue";

export interface WheelOverscrollOptions {
  getContainer: () => HTMLElement | null;
  canGoNext: ComputedRef<boolean> | Ref<boolean> | (() => boolean);
  canGoPrev: ComputedRef<boolean> | Ref<boolean> | (() => boolean);
  onNext: () => void;
  onPrev: () => void;
  threshold?: number;
  maxPullOffset?: number;
  cooldownMs?: number;
  debounceMs?: number;
}

export function useWheelOverscroll(options: WheelOverscrollOptions) {
  const {
    getContainer,
    threshold = 70,
    maxPullOffset = 16,
    cooldownMs = 440,
    debounceMs = 220
  } = options;

  const slideDirection = ref<"next" | "prev">("next");
  const pullOffset = ref(0);
  const pullProgress = ref(0);
  const isPullingNext = ref(false);
  const isPullingPrev = ref(false);
  const isTransitioning = ref(false);

  let accumulatedDeltaYDown = 0;
  let accumulatedDeltaYUp = 0;
  let resetTimer: ReturnType<typeof setTimeout> | null = null;

  function resolveCanNext(): boolean {
    return typeof options.canGoNext === "function" ? options.canGoNext() : options.canGoNext.value;
  }

  function resolveCanPrev(): boolean {
    return typeof options.canGoPrev === "function" ? options.canGoPrev() : options.canGoPrev.value;
  }

  function scheduleAccumulatorReset() {
    if (resetTimer) clearTimeout(resetTimer);
    resetTimer = setTimeout(() => {
      accumulatedDeltaYDown = 0;
      accumulatedDeltaYUp = 0;
      pullOffset.value = 0;
      pullProgress.value = 0;
      isPullingNext.value = false;
      isPullingPrev.value = false;
    }, debounceMs);
  }

  function triggerTransition(direction: "next" | "prev") {
    slideDirection.value = direction;
    isTransitioning.value = true;
    pullOffset.value = 0;
    pullProgress.value = 0;
    isPullingNext.value = false;
    isPullingPrev.value = false;

    if (direction === "next") {
      options.onNext();
    } else {
      options.onPrev();
    }

    setTimeout(() => {
      isTransitioning.value = false;
    }, cooldownMs);
  }

  function handleWheel(e: WheelEvent) {
    if (isTransitioning.value) {
      return;
    }

    const container = getContainer();
    if (!container) return;

    const { scrollTop, clientHeight, scrollHeight } = container;
    // 无论是超长内容滚到底/顶，还是内容短于视口未出现滚动条，均能准确响应滚动翻页
    const isAtTop = scrollTop <= 12 || scrollHeight <= clientHeight;
    const isAtBottom = (scrollTop + clientHeight >= scrollHeight - 12) || scrollHeight <= clientHeight;

    // 向下滚动到底部 -> 触发切到下一模块
    if (e.deltaY > 0 && isAtBottom) {
      if (resolveCanNext()) {
        accumulatedDeltaYDown += e.deltaY;
        pullOffset.value = -Math.min(maxPullOffset, Math.round(Math.sqrt(accumulatedDeltaYDown) * 1.8));
        pullProgress.value = Math.min(100, Math.round((accumulatedDeltaYDown / threshold) * 100));
        isPullingNext.value = true;
        scheduleAccumulatorReset();

        if (accumulatedDeltaYDown >= threshold) {
          accumulatedDeltaYDown = 0;
          accumulatedDeltaYUp = 0;
          triggerTransition("next");
        }
      }
    }
    // 向上滚动到顶部 -> 触发切到上一模块
    else if (e.deltaY < 0 && isAtTop) {
      if (resolveCanPrev()) {
        accumulatedDeltaYUp += Math.abs(e.deltaY);
        pullOffset.value = Math.min(maxPullOffset, Math.round(Math.sqrt(accumulatedDeltaYUp) * 1.8));
        pullProgress.value = Math.min(100, Math.round((accumulatedDeltaYUp / threshold) * 100));
        isPullingPrev.value = true;
        scheduleAccumulatorReset();

        if (accumulatedDeltaYUp >= threshold) {
          accumulatedDeltaYDown = 0;
          accumulatedDeltaYUp = 0;
          triggerTransition("prev");
        }
      }
    } else {
      accumulatedDeltaYDown = 0;
      accumulatedDeltaYUp = 0;
      pullOffset.value = 0;
      pullProgress.value = 0;
      isPullingNext.value = false;
      isPullingPrev.value = false;
    }
  }

  function attachWheelListener() {
    const container = getContainer();
    if (container) {
      container.removeEventListener("wheel", handleWheel);
      container.addEventListener("wheel", handleWheel, { passive: true });
    }
  }

  function detachWheelListener() {
    const container = getContainer();
    if (container) {
      container.removeEventListener("wheel", handleWheel);
    }
    if (resetTimer) clearTimeout(resetTimer);
  }

  onMounted(() => {
    attachWheelListener();
  });

  onActivated(() => {
    attachWheelListener();
  });

  onDeactivated(() => {
    detachWheelListener();
  });

  onUnmounted(() => {
    detachWheelListener();
  });

  function lockOverscroll(durationMs = 600) {
    isTransitioning.value = true;
    accumulatedDeltaYDown = 0;
    accumulatedDeltaYUp = 0;
    pullOffset.value = 0;
    pullProgress.value = 0;
    isPullingNext.value = false;
    isPullingPrev.value = false;
    setTimeout(() => {
      isTransitioning.value = false;
    }, durationMs);
  }

  return {
    slideDirection,
    pullOffset,
    pullProgress,
    isPullingNext,
    isPullingPrev,
    isTransitioning,
    triggerTransition,
    lockOverscroll,
    attachWheelListener,
    detachWheelListener
  };
}
