import request from "../lib/request";

export function getAlertList(data: Record<string, any>): Promise<any> {
  return request.post("/v1/alert/list", data);
}


