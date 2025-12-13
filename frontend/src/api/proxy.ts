import request from "../lib/request";

// 代理状态
export function getProxyStatus(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/status", data);
}

// 创建代理
export function addProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/create", data);
}
// 编辑代理
export function editProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/edit", data);
}

// 删除代理
export function delProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/delete", data);
}
// 批量删除
export function delBacthProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/batch/delete", data);
}
// 启动代理
export function startProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/start", data);
}

// 停止代理
export function stopProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/stop", data);
}
// 重启代理
export function restartProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/restart", data);
}
// 导出代理
export function exportProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/export", data);
}
// 导入代理
export function importProxy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/proxy/import", data,);
}