import request from "../lib/request";

// 标签列表
export function getTagList(data: Record<string, any>): Promise<any> {
  return request.post("/v1/tag/list", data);
}

// 新增标签
export function addTag(data: Record<string, any>): Promise<any> {
  return request.post("/v1/tag/create", data);
}

// 修改标签
export function editTag(data: Record<string, any>): Promise<any> {
  return request.post("/v1/tag/update", data);
}

// 删除标签
export function delTag(data: Record<string, any>): Promise<any> {
  return request.post("/v1/tag/delete", data);
}

// 批量删除标签
export function batchDelTag(data: Record<string, any>): Promise<any> {
  return request.post("/v1/tag/batch-delete", data);
}

