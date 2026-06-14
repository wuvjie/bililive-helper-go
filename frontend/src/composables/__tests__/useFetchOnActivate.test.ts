import { describe, it, expect, vi } from "vitest";
import { defineComponent, nextTick } from "vue";
import { mount } from "@vue/test-utils";
import { useFetchOnActivate } from "../useFetchOnActivate";

describe("useFetchOnActivate", () => {
  it("在 onMounted 时调用 fetchFn", async () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined);

    const TestComponent = defineComponent({
      setup() {
        useFetchOnActivate(fetchFn);
        return {};
      },
      template: "<div />",
    });

    mount(TestComponent);
    await nextTick();

    expect(fetchFn).toHaveBeenCalledOnce();
  });

  it("onMountedOnly 模式下只在 onMounted 调用", async () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined);

    const TestComponent = defineComponent({
      setup() {
        useFetchOnActivate(fetchFn, { onMountedOnly: true });
        return {};
      },
      template: "<div />",
    });

    mount(TestComponent);
    await nextTick();

    expect(fetchFn).toHaveBeenCalledOnce();
  });
});
