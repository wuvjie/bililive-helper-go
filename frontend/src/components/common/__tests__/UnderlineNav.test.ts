import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import UnderlineNav from "../UnderlineNav.vue";

describe("UnderlineNav.vue", () => {
  const items = [
    { label: "全部", value: "all", count: 18 },
    { label: "已合并", value: "merged", count: 12 },
    { label: "原始录像", value: "raw", count: 6 },
  ];

  it("renders all items with counts", () => {
    const wrapper = mount(UnderlineNav, {
      props: {
        modelValue: "all",
        items,
      },
    });

    const buttons = wrapper.findAll(".underline-nav-item");
    expect(buttons.length).toBe(3);
    expect(buttons[0].text()).toContain("全部");
    expect(buttons[0].text()).toContain("18");
    expect(buttons[0].classes()).toContain("is-active");
    expect(buttons[0].attributes("aria-selected")).toBe("true");
    expect(buttons[1].classes()).not.toContain("is-active");
    expect(buttons[1].attributes("aria-selected")).toBe("false");
  });

  it("emits update:modelValue and change when an item is clicked", async () => {
    const wrapper = mount(UnderlineNav, {
      props: {
        modelValue: "all",
        items,
      },
    });

    const buttons = wrapper.findAll(".underline-nav-item");
    await buttons[1].trigger("click");

    expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["merged"]);
    expect(wrapper.emitted("change")).toBeTruthy();
    expect(wrapper.emitted("change")?.[0]).toEqual(["merged"]);
  });
});
