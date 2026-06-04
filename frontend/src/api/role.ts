import request from "../lib/request";

// 角色列表
export function getRoleList(data: Record<string, any>): Promise<any> {
  return request.post("/v1/role/list", data);
}

// 新增角色
export function addRole(data: Record<string, any>): Promise<any> {
  return request.post("/v1/role/create", data);
}

// 修改角色
export function editRole(data: Record<string, any>): Promise<any> {
  return request.post("/v1/role/update", data);
}

// 删除角色
export function delRole(data: Record<string, any>): Promise<any> {
  return request.post("/v1/role/delete", data);
}
