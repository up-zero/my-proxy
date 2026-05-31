import config from "@/config";
import { getCurrentLocale, t } from "@/i18n";
import axios from "axios";
import { ApiError } from "./error";

import { toast } from "./util";
import router from "@/router";

const request = axios.create({
  baseURL: config.apiPrefix,
  timeout: 5000,
});

export function getStoredToken(): string {
  return localStorage.getItem(`${config.name}:token`) || "";
}

export function buildCaptureWebSocketUrl(taskUuid: string): string {
  const apiBase = config.apiPrefix || "/api";
  const url = new URL(apiBase, window.location.origin);

  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = `${url.pathname.replace(/\/$/, "")}/v1/ws/capture`;
  url.searchParams.set("task_uuid", taskUuid);
  url.searchParams.set("language", getCurrentLocale());

  const token = getStoredToken();
  if (token) {
    url.searchParams.set("token", token);
  }

  return url.toString();
}

request.interceptors.request.use((conf) => {
  let token = getStoredToken();
  conf.headers.set("language", getCurrentLocale());
  if (token) {
    conf.headers.set("Authorization", token);
  }

  return conf;
});

request.interceptors.response.use(
  (response) => {
    const headers = response.headers || {};
    const contentType = headers["content-type"];
    const disposition = headers["content-disposition"];

    if (
      response.config.responseType === "blob" ||
      disposition ||
      (contentType && !contentType.includes("application/json"))
    ) {
      return response.data;
    }
    const data = response.data;
    if (data.code == 60400) {
      router.push("/login");
      toast(data.msg || t("auth.loginRequired"));
      return {};
    } else if (data.code !== 200) {
      toast(data.msg, "error");
      throw new ApiError(t("request.systemErrorWithUrl", { url: response.config.url || "" }));
    }
    return data;
  },
  (err) => {
    if (err.response.status == 401) {
      router.push("/login");
      toast(t("auth.loginRequired"));
      return {};
    }
    throw new ApiError(err.response.message);
  }
);

export default request;
