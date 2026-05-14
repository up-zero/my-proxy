import request from "../lib/request";

// 分组列表
export function getGroupList(data: Record<string, any>): Promise<any> {
  return request.post("/v1/group/list", data);
}

// 新增分组
export function addGroup(data: Record<string, any>): Promise<any> {
  return request.post("/v1/group/create", data);
}

// 修改分组
export function editGroup(data: Record<string, any>): Promise<any> {
  return request.post("/v1/group/update", data);
}

// 删除分组
export function delGroup(data: Record<string, any>): Promise<any> {
  return request.post("/v1/group/delete", data);
}

