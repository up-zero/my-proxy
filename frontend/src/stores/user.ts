import { login } from "@/api/user";
import config from "@/config";
import { t } from "@/i18n";
import { toast } from "@/lib/util";
import router from "@/router";
import { defineStore } from "pinia";

const USERINFO = () => ({
  isAdmin: false,
  roleID: "",
  roleName: "",
  permissions: [] as string[],
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

      // 设置 token
      if (result.token)
        localStorage.setItem(`${config.name}:token`, result.token);

      // 设置用户信息
      const isAdmin = result.level === "root";
      this.userinfo = {
        isAdmin,
        roleID: result.role_id || "",
        roleName: result.role_name || "",
        permissions: result.permissions || [],
      };
    },
    // 判断是否有某个权限
    hasPermission(perm: string): boolean {
      if (this.userinfo.isAdmin) return true;
      return this.userinfo.permissions.includes(perm);
    },
    // 登录
    async login(loginForm: any) {
      const res = await login(loginForm);

      if (!res.data) return;
      this._init(res.data);
      // 跳转首页
      router.replace("/");
      toast(t("auth.loginSuccess"));
    },
    // 退出登录
    async logout() {
      // 设置用户信息
      this.userinfo = USERINFO();
      // 移除Token
      localStorage.removeItem(`${config.name}:token`);
      // 跳转登录页面
      router.replace("/login");
      toast(t("auth.logoutSuccess"));
    },
  },
});
