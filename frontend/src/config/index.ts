export default {
  name: "my-proxy", // 项目名称
  apiPrefix: import.meta.env.VITE_API_BASE_URL, // 接口前缀
  logLevel: "debug", // 日志类型
};
