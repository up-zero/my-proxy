import logger from "./logger";
import { toast } from "./util";
import { t } from "@/i18n";

/* *********接口相关错误********** */
export class ApiError extends Error {
  name = t("errors.api");
}

/* *********全局错误处理方法********** */
export function errorHandle(err: Error | any, info: string = "") {
  let name: string;
  let msg: string;
  if (err instanceof Error) {
    name = err.name;
    msg = err.message;
  } else {
    name = t("errors.unknownType");
    try {
      msg = JSON.stringify(err) || t("errors.unknownError");
    } catch (e) {
      msg = t("errors.unknownError");
    }
  }
  logger.error(name, err, info);
  toast(msg, "error");
}
