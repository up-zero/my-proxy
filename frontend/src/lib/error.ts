import logger from "./logger";
import { toast } from "./util";

/* *********接口相关错误********** */
export class ApiError extends Error {
  name = "api错误";
}

/* *********全局错误处理方法********** */
export function errorHandle(err: Error | any, info: string = "") {
  let name = "",
    msg = "";
  if (err instanceof Error) {
    name = err.name;
    msg = err.message;
  } else {
    name = "未知类型";
    try {
      msg = JSON.stringify(err) || "未知错误";
    } catch (e) {
      msg = "未知错误";
    }
  }
  logger.error(name, err, info);
  toast(msg, "error");
}
