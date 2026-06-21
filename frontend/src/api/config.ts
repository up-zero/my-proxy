import request from "../lib/request";

export interface ConfigItem {
  key: string;
  value: string;
  default_value: string;
}

export interface UpdateItem {
  key: string;
  value: string;
}

// 获取系统设置
export function getSystemSettings(): Promise<any> {
  return request.post("/v1/settings/get");
}

// 更新系统设置
export function updateSystemSettings(items: UpdateItem[]): Promise<any> {
  return request.post("/v1/settings/update", { items });
}
