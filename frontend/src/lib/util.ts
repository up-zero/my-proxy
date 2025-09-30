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

export const parseTime = (date: any, fmt?: string) => {
  // 表格格式化日期

  date = new Date(date);
  if (typeof fmt === "undefined") {
    fmt = "yyyy-MM-dd hh:mm:ss";
  }
  if (/(y+)/.test(fmt)) {
    fmt = fmt.replace(
      RegExp.$1,
      (date.getFullYear() + "").substr(4 - RegExp.$1.length)
    );
  }
  let o: any = {
    "M+": date.getMonth() + 1,
    "d+": date.getDate(),
    "h+": date.getHours(),
    "m+": date.getMinutes(),
    "s+": date.getSeconds(),
  };
  for (let k in o) {
    if (new RegExp(`(${k})`).test(fmt)) {
      let str = o[k] + "";
      fmt = fmt.replace(
        RegExp.$1,
        RegExp.$1.length === 1 ? str : ("00" + str).substr(str.length)
      );
    }
  }
  return fmt;
};
