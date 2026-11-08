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
      e.preventDefault();
      return;
    }

    const container = getContainer();
    if (!container) return;

    const { scrollTop, clientHeight, scrollHeight } = container;
    const isScrollable = scrollHeight > clientHeight + 40;
    const isAtBottom = isScrollable && (scrollTop + clientHeight >= scrollHeight - 8);
    const isAtTop = isScrollable && (scrollTop <= 8);

    // Downward wheel at bottom -> Pull towards next
    if (e.deltaY > 0 && isAtBottom) {
      if (resolveCanNext()) {
        accumulatedDeltaYDown += e.deltaY;
        pullOffset.value = -Math.min(maxPullOffset, Math.round(Math.sqrt(accumulatedDeltaYDown) * 1.8));
        pullProgress.value = Math.min(100, Math.round((accumulatedDeltaYDown / threshold) * 100));
        isPullingNext.value = true;
        scheduleAccumulatorReset();

        if (accumulatedDeltaYDown >= threshold) {
          e.preventDefault();
          accumulatedDeltaYDown = 0;
          accumulatedDeltaYUp = 0;
          triggerTransition("next");
        }
      }
    }
    // Upward wheel at top -> Pull towards prev
    else if (e.deltaY < 0 && isAtTop) {
      if (resolveCanPrev()) {
        accumulatedDeltaYUp += Math.abs(e.deltaY);
        pullOffset.value = Math.min(maxPullOffset, Math.round(Math.sqrt(accumulatedDeltaYUp) * 1.8));
        pullProgress.value = Math.min(100, Math.round((accumulatedDeltaYUp / threshold) * 100));
        isPullingPrev.value = true;
        scheduleAccumulatorReset();

        if (accumulatedDeltaYUp >= threshold) {
          e.preventDefault();
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
      container.addEventListener("wheel", handleWheel, { passive: false });
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
