import { buildTerminalWebSocketUrl } from "../lib/request";

export interface TerminalConnInfo {
  id: string;
  name: string;
  host: string;
  port: string;
  username: string;
  password: string;
}

export interface ProxyTerminalInfo {
  uuid: string;
  name: string;
  listen_address: string;
}

/**
 * 创建终端 WebSocket 连接
 */
export function createTerminalSocket(info: TerminalConnInfo): WebSocket {
  const url = buildTerminalWebSocketUrl(info);
  return new WebSocket(url);
}
