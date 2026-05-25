import request from "../lib/request";

// 仪表盘概览
export function getDashboardOverview(): Promise<any> {
  return request.get("/v1/dashboard/overview");
}

