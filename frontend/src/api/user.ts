import request from "../lib/request";

export function login(data: any) {
    return request({
        url: '/v1/login',
        method: 'post',
        data
    })
}

// 修改密码
export function changePassword(data: Record<string, any>): Promise<any> {
  return request.post("/v1/edit/password", data);
}