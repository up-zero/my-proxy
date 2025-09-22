<template>
  <a-config-provider>
    <!-- 登录页面 -->
    <RouterView v-if="route.path == '/login'" />

    <!-- 其他页面 -->
    <template v-else>
      <!-- header -->
      <a-affix :offset="0">
        <div class="m-header f-between">
          <div class="u-icon f-icon icon-logo">
            <img src="@/assets/favicon.ico" alt="" />My Proxy
          </div>
          <a-dropdown trigger="hover">
            <div class="m-user f-between f-flex-aligm-center">
              <div class="u-img">
                <img src="@/assets/img/logo.png" alt="" />
              </div>
              <p class="w1">管理员</p>
            </div>
            <template #overlay>
              <a-menu>
                <a-menu-item>
                  <a-button type="link" @click="logout">退出登录</a-button>
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-affix>

      <div class="m-main f-flex">
        <!-- 菜单 -->
        <div class="m-side">
          <a-menu :default-active="route.path" router>
            <template v-for="item in menus">
              <a-menu-item
                v-if="!item.children || item.children.length == 0"
                :index="item.path"
                >{{ item.name }}</a-menu-item
              >
              <a-sub-menu v-else :key="item.path" :index="item.path">
                <template #title>
                  <!-- <div class="u-icon f-icon" :class="[`z-${menu.menu_icon}`]"></div> -->
                  <span>{{ item.name }}</span>
                </template>
                <a-menu-item
                  v-for="sub of item.children"
                  :key="sub.path"
                  v-show="!sub.meta?.hidden"
                  :index="sub.path"
                  >{{ sub.name }}</a-menu-item
                >
              </a-sub-menu>
            </template>
          </a-menu>
        </div>

        <!-- 显示内容 -->
        <div class="m-content">
          <a-breadcrumb>
            <!-- <a-breadcrumb-item>{{ route.matched[0]?.name }}</a-breadcrumb-item> -->
            <a-breadcrumb-item>{{ route.name }}</a-breadcrumb-item>
          </a-breadcrumb>

          <RouterView class="p-page" />
        </div>
      </div>
    </template>
  </a-config-provider>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import store from "./stores";
import { computed } from "vue";

const user = store.useUserStore();
const route = useRoute();
const router = useRouter();

const getMenu = (menus: any) => {
  let res: any = [];

  menus.forEach((item: any) => {
    if (item.meta?.isMenu || user.userinfo.isAdmin) {
      let menu = {
        name: item.name,
        path: item.path,
        children: [],
      };
      if (item.children && item.children.length > 0) {
        const childrenMenu = getMenu(item.children);
        if (childrenMenu.length == 1) {
          menu = childrenMenu[0];
          // 只有一个子菜单提升层级
        } else {
          menu.children = childrenMenu;
        }
      }
      res.push(menu);
    }
  });
  return res;
};
const menus = computed(() => {
  return getMenu(router.options.routes);
});

// 退出登录
async function logout() {
  await user.logout();
}
</script>

<style lang="less">
@import "@/assets/less/global.less";

.text-center {
  text-align: center;
}
.border {
  border: 1px solid #eee;
}
</style>

<style lang="less" scoped>
@headerHeight: 56px;

.m-header {
  height: @headerHeight;
  background: #ffffff;
  box-shadow: inset 0px -1px 0px 0px #e7e7e7;
  padding-left: 20px;
  padding-right: 24px;
  .u-icon {
    display: flex;
    align-items: center;
    width: 180px;
    img {
      width: 25px;
      margin-right: 10px;
    }
  }
  .m-user {
    outline: none;
    .u-img {
      width: 32px;
      height: 32px;
      img {
        width: 100%;
        border-radius: 32px;
        height: 100%;
      }
      border: 1px solid #f6f6f6;

      background-color: #ececec;
    }
    .w1 {
      font-size: 14px;

      line-height: 20px;

      margin-left: 8px;
      margin-bottom: 0;
    }
  }
}
.m-side {
  width: 232px;
  height: calc(100vh - @headerHeight);
  overflow-y: auto;
  padding-top: 12px;
  :deep(.a-menu) {
    border: none;
  }
}
.m-content {
  flex: 1;
  height: calc(100vh - @headerHeight);
  padding: 20px;
  background-color: #f2f4f7;
  .p-page {
    background-color: #fff;
    padding: 20px;
    margin-top: 20px;
    height: calc(100vh - @headerHeight - 70px);
    overflow-y: auto;
  }
}
</style>
