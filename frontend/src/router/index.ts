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
      redirect: "/proxyManage/index",
    },
    {
      path: "/proxyManage",
      name: "我的代理",
      meta: {
        isMenu: true,
      },
      redirect: "/proxyManage/index",
      children: [
        {
          path: "/proxyManage/index",
          name: "代理列表",
          meta: {
            isMenu: true,
          },
          component: () => import("../views/proxyManage/index.vue"),
        },
      ],
    },
    {
      path: "/changePassword",
      name: "修改密码",
      component: () => import("../views/userManage/changePassword.vue"),
    },
  ],
});

export default router;
