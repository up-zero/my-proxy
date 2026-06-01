import request from "../lib/request";

export function getTrafficPolicyList(data: Record<string, any>): Promise<any> {
  return request.post("/v1/traffic-policy/list", data);
}

export function addTrafficPolicy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/traffic-policy/create", data);
}

export function editTrafficPolicy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/traffic-policy/update", data);
}

export function enableTrafficPolicy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/traffic-policy/enable", data);
}

export function disableTrafficPolicy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/traffic-policy/disable", data);
}

export function deleteTrafficPolicy(data: Record<string, any>): Promise<any> {
  return request.post("/v1/traffic-policy/delete", data);
}


