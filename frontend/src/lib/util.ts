/** -----------------纯工具函数--------------- */

import { message } from "ant-design-vue";

/** 延迟 */
export function timeout(ms: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

/* toast */
export function toast(
  msg: string,
  type: "success" | "error" | "warning" | "info" = "success",
  duration: number = 1
) {
  return new Promise((resolve) => {
    message[type]({
      content: msg,
      type,
      duration,
    });
    timeout(duration).then(resolve);
  });
}
