import request from "../lib/request";

export function login(data: any) {
  return request({
    url: "/v1/login",
    method: "post",
    data,
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
