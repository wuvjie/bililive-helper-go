import type { Component } from "vue";

export interface DialogOptions {
  id?: string;
  visible?: boolean;
  title?: string;
  width?: string;
  fullscreen?: boolean;
  closeOnClickModal?: boolean;
  closeOnPressEscape?: boolean;
  draggable?: boolean;
  showFooter?: boolean;
  showCancel?: boolean;
  cancelText?: string;
  sureText?: string;
  sureLoading?: boolean;
  contentComponent?: Component;
  contentProps?: Record<string, any>;
  closeCallback?: () => void;
  beforeCancel?: () => void;
  beforeSure?: () => void;
}
