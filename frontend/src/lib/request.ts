import config from "@/config";
import axios from "axios";
import { ApiError } from "./error";

import { toast } from "./util";
import router from "@/router";

const request = axios.create({
  baseURL: config.apiPrefix,
  timeout: 5000,
});

request.interceptors.request.use((conf) => {
  let token = localStorage.getItem(`${config.name}:token`);
  if (token) {
    conf.headers.set("Authorization", token);
  }

  return conf;
});

request.interceptors.response.use(
  (response) => {
    const headers = response.headers || {};
    const contentType = headers["content-type"];
    const disposition = headers["content-disposition"];

    if (
      response.config.responseType === "blob" ||
      disposition ||
      (contentType && !contentType.includes("application/json"))
    ) {
      return response.data;
    }
    const data = response.data;
    if (data.code == 60400) {
      router.push("/login");
      toast("请登录!");
      return {};
    } else if (data.code !== 200) {
      toast(data.msg, "error");
      throw new ApiError("系统错误，api链接：" + response.config.url);
    }
    return data;
  },
  (err) => {
    if (err.response.status == 401) {
      router.push("/login");
      toast("请登录!");
      return {};
    }
    throw new ApiError(err.response.message);
  }
);

export default request;
