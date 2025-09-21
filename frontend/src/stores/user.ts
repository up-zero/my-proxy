import { login } from "@/api/user";
import config from "@/config";
import { toast } from "@/lib/util";
import router from "@/router";
import { defineStore } from "pinia";

const USERINFO = () => ({
  isAdmin: false,
  menu: {
    ids: [],
    tree: [],
  },
});

export default defineStore("user", {
  persist: true,
  state: () => ({
    userinfo: USERINFO(),
  }),
  getters: {},
  actions: {
    // 初始化操作
    _init(result: any) {
      console.log(result, "222222");

      // 设置token
      if (result.token)
        localStorage.setItem(`${config.name}:token`, result.token);
    },
    // 登录
    async login(loginForm: any) {
      const res = await login(loginForm);

      if (!res.data) return;
      this._init(res.data);
      // 跳转首页
      router.replace("/");
      toast("登录成功");
    },
    // 退出登录
    async logout() {
      // 设置用户信息
      this.userinfo = USERINFO();
      // 移除Token
      localStorage.removeItem(`${config.name}:token`);
      // 跳转登录页面
      router.replace("/login");
      toast("已退出登录");
    },
  },
});
