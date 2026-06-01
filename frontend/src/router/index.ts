import { createRouter, createWebHistory } from "vue-router";

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
      },
      component: () => import("../views/dashboard/index.vue"),
    },
    {
      path: "/proxyManage",
      name: "我的代理",
      meta: {
        isMenu: true,
        titleKey: "routes.proxyManage",
      },
      redirect: "/proxyManage/index",
      children: [
        {
          path: "/proxyManage/index",
          name: "代理列表",
          meta: {
            isMenu: true,
            titleKey: "routes.proxyList",
          },
          component: () => import("../views/proxyManage/index.vue"),
        },
        {
          path: "/proxyManage/tag",
          name: "标签管理",
          meta: {
            isMenu: true,
            titleKey: "routes.tagManage",
          },
          component: () => import("../views/proxyManage/tag/index.vue"),
        },
        {
          path: "/proxyManage/trafficPolicy",
          name: "限速配额",
          meta: {
            isMenu: true,
            titleKey: "routes.trafficPolicy",
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
      },
      redirect: "/operation/alert",
      children: [
        {
          path: "/operation/alert",
          name: "告警通知",
          meta: {
            isMenu: true,
            titleKey: "routes.alertNotify",
          },
          component: () => import("../views/operation/alert/index.vue"),
        },
      ],
    },
    {
      path: "/userManage",
      name: "用户管理",
      meta: {
        isMenu: true,
        titleKey: "routes.userManage",
      },
      redirect: "/userManage/index",
      children: [
        {
          path: "/userManage/index",
          name: "用户列表",
          meta: {
            isMenu: true,
            titleKey: "routes.userList",
          },
          component: () => import("../views/userManage/user/index.vue"),
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
