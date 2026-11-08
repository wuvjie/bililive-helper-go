import { describe, it, expect, vi } from "vitest";
import { defineComponent, ref, nextTick } from "vue";
import { mount } from "@vue/test-utils";
import { useWheelOverscroll } from "../useWheelOverscroll";

describe("useWheelOverscroll", () => {
  it("初始化状态正确", () => {
    let composable: ReturnType<typeof useWheelOverscroll> | null = null;
    const onNext = vi.fn();
    const onPrev = vi.fn();

    const TestComponent = defineComponent({
      setup() {
        composable = useWheelOverscroll({
          getContainer: () => null,
          canGoNext: ref(true),
          canGoPrev: ref(true),
          onNext,
          onPrev
        });
        return {};
      },
      template: "<div />"
    });

    mount(TestComponent);

    expect(composable).not.toBeNull();
    expect(composable!.slideDirection.value).toBe("next");
    expect(composable!.pullOffset.value).toBe(0);
    expect(composable!.pullProgress.value).toBe(0);
    expect(composable!.isPullingNext.value).toBe(false);
    expect(composable!.isPullingPrev.value).toBe(false);
    expect(composable!.isTransitioning.value).toBe(false);
  });

  it("手动触发 triggerTransition 正确分发并在冷却期锁定", async () => {
    let composable: ReturnType<typeof useWheelOverscroll> | null = null;
    const onNext = vi.fn();
    const onPrev = vi.fn();

    const TestComponent = defineComponent({
      setup() {
        composable = useWheelOverscroll({
          getContainer: () => null,
          canGoNext: ref(true),
          canGoPrev: ref(true),
          onNext,
          onPrev,
          cooldownMs: 50
        });
        return {};
      },
      template: "<div />"
    });

    mount(TestComponent);

    composable!.triggerTransition("next");
    expect(onNext).toHaveBeenCalledOnce();
    expect(composable!.slideDirection.value).toBe("next");
    expect(composable!.isTransitioning.value).toBe(true);

    composable!.triggerTransition("prev");
    expect(onPrev).toHaveBeenCalledOnce();
    expect(composable!.slideDirection.value).toBe("prev");
  });
});
