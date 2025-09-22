import request from "../lib/request";

export function login(data: any) {
    return request({
        url: '/v1/login',
        method: 'post',
        data
    })
}