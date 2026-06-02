import request from "../lib/request";

export function getAuditList(data: Record<string, any>): Promise<any> {
  return request.post("/v1/audit/list", data);
}
