import { createRouter, createWebHistory } from "vue-router";

// 权限常量（按菜单项）
export const PERMS = {
  dashboardView: "dashboard.view",        // 仪表盘
  proxyView: "proxy.view",                // 代理列表
  tagManage: "tag.manage",                // 标签管理
  trafficPolicy: "traffic_policy.manage", // 限速配额
  terminalView: "terminal.view",          // Web 终端
  alertView: "alert.view",                // 告警通知
  auditView: "audit.view",                // 日志审计
  userManage: "user.manage",              // 用户列表
  roleManage: "role.manage",              // 权限策略
};

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/login",
      name: "Login",
      component: () => import("../views/LoginView.vue"),
    },
    {
      path: "/",
      redirect: "/dashboard",
    },
    {
      path: "/dashboard",
      name: "仪表盘",
      meta: {
        isMenu: true,
        fullPage: true,
        titleKey: "routes.dashboard",
        icon: "DashboardOutlined",
        perm: PERMS.dashboardView,
      },
      component: () => import("../views/dashboard/index.vue"),
    },
    {
      path: "/proxyManage",
      name: "我的代理",
      meta: {
        isMenu: true,
        titleKey: "routes.proxyManage",
        icon: "ApartmentOutlined",
        perm: PERMS.proxyView,
      },
      redirect: "/proxyManage/index",
      children: [
        {
          path: "/proxyManage/index",
          name: "代理列表",
          meta: {
            isMenu: true,
            titleKey: "routes.proxyList",
            perm: PERMS.proxyView,
          },
          component: () => import("../views/proxyManage/index.vue"),
        },
        {
          path: "/proxyManage/tag",
          name: "标签管理",
          meta: {
            isMenu: true,
            titleKey: "routes.tagManage",
            perm: PERMS.tagManage,
          },
          component: () => import("../views/proxyManage/tag/index.vue"),
        },
        {
          path: "/proxyManage/trafficPolicy",
          name: "限速配额",
          meta: {
            isMenu: true,
            titleKey: "routes.trafficPolicy",
            perm: PERMS.trafficPolicy,
          },
          component: () => import("../views/proxyManage/trafficPolicy/index.vue"),
        },
        {
          path: "/proxyManage/capture",
          name: "抓包分析",
          meta: {
            hidden: true,
            fullPage: true,
            titleKey: "routes.captureAnalyze",
          },
          component: () => import("../views/proxyManage/capture.vue"),
        },
      ],
    },
    {
      path: "/operation",
      name: "运维中心",
      meta: {
        isMenu: true,
        titleKey: "routes.operationCenter",
        icon: "ToolOutlined",
        perm: PERMS.alertView,
      },
      redirect: "/operation/alert",
      children: [
        {
          path: "/operation/terminal",
          name: "Web 终端",
          meta: {
            isMenu: true,
            fullPage: true,
            titleKey: "routes.terminal",
            perm: PERMS.terminalView,
          },
          component: () => import("../views/terminal/index.vue"),
        },
        {
          path: "/operation/alert",
          name: "告警通知",
          meta: {
            isMenu: true,
            titleKey: "routes.alertNotify",
            perm: PERMS.alertView,
          },
          component: () => import("../views/operation/alert/index.vue"),
        },
        {
          path: "/operation/audit",
          name: "日志审计",
          meta: {
            isMenu: true,
            titleKey: "routes.auditLog",
            perm: PERMS.auditView,
          },
          component: () => import("../views/operation/audit/index.vue"),
        },
      ],
    },
    {
      path: "/userManage",
      name: "用户管理",
      meta: {
        isMenu: true,
        titleKey: "routes.userManage",
        icon: "TeamOutlined",
        perm: PERMS.userManage,
      },
      redirect: "/userManage/index",
      children: [
        {
          path: "/userManage/index",
          name: "用户列表",
          meta: {
            isMenu: true,
            titleKey: "routes.userList",
            perm: PERMS.userManage,
          },
          component: () => import("../views/userManage/user/index.vue"),
        },
        {
          path: "/userManage/role",
          name: "权限策略",
          meta: {
            isMenu: true,
            titleKey: "routes.permPolicy",
            perm: PERMS.roleManage,
          },
          component: () => import("../views/userManage/role/index.vue"),
        },
      ],
    },
    {
      path: "/changePassword",
      name: "修改密码",
      meta: {
        titleKey: "routes.changePassword",
      },
      component: () => import("../views/userManage/changePassword.vue"),
    },
  ],
});

export default router;
