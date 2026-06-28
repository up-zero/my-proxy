import request from "../lib/request";

export interface NodeItem {
  uuid: string;
  name: string;
  address: string;
  secret_key: string;
  enabled: boolean;
  is_local: boolean;
  created_at: number;
  updated_at: number;
}

// 获取节点列表
export function getNodeList(): Promise<any> {
  return request.post("/v1/node/list");
}

// 创建节点
export function createNode(data: {
  name: string;
  address: string;
  secret_key: string;
  enabled?: boolean;
}): Promise<any> {
  return request.post("/v1/node/create", data);
}

// 更新节点
export function updateNode(data: {
  uuid: string;
  name: string;
  address: string;
  secret_key: string;
  enabled?: boolean;
}): Promise<any> {
  return request.post("/v1/node/update", data);
}

// 删除节点
export function deleteNode(data: { uuid: string }): Promise<any> {
  return request.post("/v1/node/delete", data);
}
