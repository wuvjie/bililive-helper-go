import type { Component } from "vue";

export interface DrawerOptions {
  id?: string;
  visible?: boolean;
  title?: string;
  size?: string;
  direction?: "rtl" | "ltr" | "ttb" | "btt";
  showFooter?: boolean;
  cancelText?: string;
  contentComponent?: Component;
  contentProps?: Record<string, any>;
  contentHtml?: string;
  closeCallback?: () => void;
}
