import request from "../lib/request";
import { encryptPassword } from "../lib/crypto";

export async function login(data: any) {
  // 密码 RSA 加密后传输（后端仅接受密文）
  const password = await encryptPassword(data?.password ?? "");
  return request({
    url: "/v1/login",
    method: "post",
    data: { ...data, password },
  });
}

// 修改密码
export function changePassword(data: Record<string, any>): Promise<any> {
  return request.post("/v1/edit/password", data);
}

// 用户列表
export function getUserList(data: Record<string, any>): Promise<any> {
  return request.post("/v1/user/list", data);
}

// 新增用户
export function addUser(data: Record<string, any>): Promise<any> {
  return request.post("/v1/user/create", data);
}
// 修改用户
export function editUser(data: Record<string, any>): Promise<any> {
  return request.post("/v1/user/update", data);
}

// 删除用户
export function delUser(data: Record<string, any>): Promise<any> {
  return request.post("/v1/user/delete", data);
}
