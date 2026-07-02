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

function getCurrentNodeUuid(): string | null {
  try {
    const raw = localStorage.getItem(`${config.name}:currentNode`);
    if (raw) {
      const parsed = JSON.parse(raw);
      if (parsed.uuid && parsed.uuid !== "node-local" && !parsed.isLocal) {
        return parsed.uuid;
      }
    }
  } catch {
    // ignore
  }
  return null;
}

export interface TerminalConnParams {
  host: string;
  port: string;
  username: string;
  password: string;
  id: string;
}

export function buildTerminalWebSocketUrl(params: TerminalConnParams): string {
  const apiBase = config.apiPrefix || "/api";
  const url = new URL(apiBase, window.location.origin);

  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = `${url.pathname.replace(/\/$/, "")}/v1/ws/terminal`;
  url.searchParams.set("host", params.host);
  url.searchParams.set("port", params.port || "22");
  url.searchParams.set("username", params.username);
  url.searchParams.set("password", params.password || "");
  url.searchParams.set("session_id", params.id);
  url.searchParams.set("language", getCurrentLocale());

  const token = getStoredToken();
  if (token) {
    url.searchParams.set("token", token);
  }

  // 非本地节点时传递 node_uuid，服务端根据此参数代理到子节点
  const nodeUuid = getCurrentNodeUuid();
  if (nodeUuid) {
    url.searchParams.set("node_uuid", nodeUuid);
  }

  return url.toString();
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

  // 非本地节点时注入 X-Node-Id（具体是否转发由后端中间件判断）
  const nodeUuid = getCurrentNodeUuid();
  if (nodeUuid) {
    conf.headers.set("X-Node-Id", nodeUuid);
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
