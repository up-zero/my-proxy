<template>
  <a-config-provider :locale="locale" :theme="antTheme">
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
          <div class="m-header-right">
            <ThemeSwitcher />
            <LanguageSwitcher variant="compact" />
            <a-dropdown trigger="hover">
              <div class="m-user f-between f-flex-aligm-center">
                <div class="u-img">
                  <img src="@/assets/img/logo.png" alt="" />
                </div>
                <p class="w1">{{ t("common.administrator") }}</p>
              </div>
              <template #overlay>
                <a-menu @click="topMenuClick">
                  <a-menu-item key="cp">
                    <a-button type="link">{{ t("header.editPassword") }}</a-button>
                  </a-menu-item>
                  <a-menu-item key="out">
                    <a-button type="link">{{ t("header.logout") }}</a-button>
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>
        </div>
      </a-affix>

      <div class="m-main f-flex">
        <!-- 菜单 -->
        <div :class="['m-side', { 'm-side-collapsed': collapsed }]">
          <div class="m-side-scroll">
            <a-menu
              mode="inline"
              :selectedKeys="[route.path]"
              v-model:openKeys="menuOpenKeys"
              :inlineCollapsed="collapsed"
              triggerSubMenuAction="hover"
              @click="routerTo"
            >
              <template v-for="item in menus">
                <a-menu-item
                  v-if="!item.children || item.children.length == 0"
                  :key="item.path"
                >
                  <template #icon>
                    <component :is="iconMap[item.icon as string]" v-if="item.icon" />
                  </template>
                  {{ item.name }}
                </a-menu-item>
                <template v-else>
                  <a-sub-menu :key="item.path">
                    <template #icon>
                      <component :is="iconMap[item.icon as string]" v-if="item.icon" />
                    </template>
                    <template #title>
                      <span>{{ item.name }}</span>
                    </template>
                    <a-menu-item
                      v-for="sub of item.children"
                      :key="sub.path"
                      v-show="!sub.meta?.hidden"
                      :index="sub.path"
                    >
                      <template #icon>
                        <component :is="iconMap[sub.icon as string]" v-if="sub.icon" />
                      </template>
                      {{ sub.name }}
                    </a-menu-item>
                  </a-sub-menu>
                </template>
              </template>
            </a-menu>
          </div>
          <div class="m-side-trigger" @click="collapsed = !collapsed" :title="collapsed ? t('common.expand') : t('common.collapse')">
            <MenuFoldOutlined v-if="!collapsed" />
            <MenuUnfoldOutlined v-else />
          </div>
        </div>

        <!-- 显示内容 -->
        <div :class="['m-content', { 'full-layout': isFullPage }]">
          <a-breadcrumb>
            <template v-if="breadcrumbItems.length > 1">
              <a-breadcrumb-item v-for="(item, index) in breadcrumbItems" :key="index">
                {{ item }}
              </a-breadcrumb-item>
            </template>
            <template v-else>
              <a-breadcrumb-item>{{ currentRouteTitle }}</a-breadcrumb-item>
            </template>
          </a-breadcrumb>

          <RouterView :class="['p-page', { 'full-page-host': isFullPage }]" />
          <div class="footer">
            <a href="https://github.com/up-zero/my-proxy" target="_blank">© 2025-{{ currentYear }} up-zero</a></div>
        </div>
      </div>
      <passwordBox ref="passwordBoxRef" />
    </template>
  </a-config-provider>
</template>

<script setup lang="ts">
import passwordBox from '@/views/userManage/changePassword.vue'
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import ThemeSwitcher from "@/components/ThemeSwitcher.vue";
import { useAppI18n } from "@/i18n";
import { useRoute, useRouter } from "vue-router";
import store from "./stores";
import { computed, ref, watch } from "vue";
import antThemeLib from "ant-design-vue/es/theme";
import {
  DashboardOutlined,
  ApartmentOutlined,
  ToolOutlined,
  UserOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  KeyOutlined,
  LogoutOutlined,
  UnorderedListOutlined,
  TagsOutlined,
  ControlOutlined,
  BellOutlined,
  AuditOutlined,
  TeamOutlined,
  CodeOutlined,
  SettingOutlined,
} from "@ant-design/icons-vue";

const iconMap: Record<string, any> = {
  DashboardOutlined,
  ApartmentOutlined,
  ToolOutlined,
  UserOutlined,
  UnorderedListOutlined,
  TagsOutlined,
  ControlOutlined,
  BellOutlined,
  AuditOutlined,
  TeamOutlined,
  KeyOutlined,
  CodeOutlined,
  SettingOutlined,
};

const collapsed = ref(false);

// 主题
const themeStore = store.useThemeStore();
const antTheme = computed(() => ({
  algorithm: themeStore.isDark ? antThemeLib.darkAlgorithm : antThemeLib.defaultAlgorithm,
}));

// 监听主题变化，切换 html class
watch(() => themeStore.mode, (mode) => {
  document.documentElement.classList.toggle("theme-dark", mode === "dark");
}, { immediate: true });

// 版权年份，动态获取当前年份
const currentYear = new Date().getFullYear();

// 根据当前路由自动展开对应子菜单
const menuOpenKeys = ref<string[]>([]);

const { antLocale: locale, t } = useAppI18n();
const user = store.useUserStore();
const route = useRoute();
const router = useRouter();
const passwordBoxRef = ref();

// 监听路由变化，自动更新展开的子菜单
watch(() => route.path, () => {
  const matched = route.matched;
  const keys: string[] = [];
  for (const record of matched) {
    if (record.path && record.path !== '/' && record.path !== route.path) {
      keys.push(record.path);
    }
  }
  menuOpenKeys.value = keys;
}, { immediate: true });

const getMenu = (menus: any) => {
  let res: any = [];

  menus.forEach((item: any) => {
    // 标记为 isMenu 才显示在菜单中
    const hasPerm = !item.meta?.perm || user.hasPermission(item.meta.perm);
    if (item.meta?.isMenu && hasPerm) {
      let menu = {
        name: item.meta?.titleKey ? t(item.meta.titleKey) : item.name,
        path: item.path,
        icon: item.meta?.icon || null,
        children: [],
      };
      if (item.children && item.children.length > 0) {
        const childrenMenu = getMenu(item.children);
        if (childrenMenu.length == 1) {
          const child = childrenMenu[0];
          // 继承父级 icon（如果子级没有定义）
          if (!child.icon && item.meta?.icon) {
            child.icon = item.meta.icon;
          }
          menu = child;
          // 只有一个子菜单提升层级
        } else if (childrenMenu.length > 0) {
          menu.children = childrenMenu;
        } else {
          // 所有子菜单都被过滤掉，不显示父菜单
          return;
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
const isFullPage = computed(() => Boolean(route.meta?.fullPage));
const currentRouteTitle = computed(() => (route.meta?.titleKey ? t(route.meta.titleKey as string) : route.name));

// 生成面包屑路径数组
const breadcrumbItems = computed(() => {
  const matched = route.matched;
  const items: string[] = [];
  
  for (let i = 0; i < matched.length; i++) {
    const record = matched[i];
    // 跳过根路由和登录页
    if (record.path === '/' || record.path === '/login') {
      continue;
    }
    
    // 只添加有标题的路由
    if (record.meta?.titleKey) {
      items.push(t(record.meta.titleKey as string));
    } else if (record.name && typeof record.name === 'string') {
      items.push(record.name);
    }
  }
  
  return items;
});

// 退出登录
async function logout() {
  await user.logout();
}
const routerTo = ({ key }: any) => {
  router.push(key);
  // 移动端或小屏点击菜单后自动折叠
  if (window.innerWidth < 768) {
    collapsed.value = true;
  }
};
const topMenuClick = ({ key }: any) => {
  console.log(key);
  if (key === "out") {
    logout();
  } else if (key === "cp") {
    passwordBoxRef.value.init();

    // router.push("/changePassword");
  }
};
</script>

<style lang="less">
@import "@/assets/less/global.less";
@import "@/assets/less/theme.less";

html,
body,
#app {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--color-bg-app, #f2f4f7);
}

.text-center {
  text-align: center;
}
.border {
  border: 1px solid var(--color-border-light, #eee);
}

// 深色主题 —— 静态方法组件(message / notification)
html.theme-dark {
  .ant-message-notice-content,
  .ant-notification-notice {
    box-shadow:
      0 0 0 1px rgba(255, 255, 255, 0.06),   // 微弱亮边勾勒轮廓
      0 8px 24px rgba(0, 0, 0, 0.6);           // 加重投影突出层次
  }
}
</style>

<style lang="less" scoped>
@headerHeight: 56px;

.m-header {
  height: @headerHeight;
  background: var(--color-bg-header, #ffffff);
  box-shadow: inset 0px -1px 0px 0px var(--color-border-header, #e7e7e7);
  padding-left: 20px;
  padding-right: 24px;
  .u-icon {
    display: flex;
    align-items: center;
    width: 180px;
    color: var(--color-text-on-header);
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
      border: 1px solid var(--color-user-avatar-border, #f6f6f6);

      background-color: var(--color-user-avatar-bg, #ececec);
    }
    .w1 {
      font-size: 14px;
      color: var(--color-text-on-header);
      line-height: 20px;

      margin-left: 8px;
      margin-bottom: 0;
    }
  }
}
.m-header-right {
  display: flex;
  align-items: center;
  gap: 14px;
}
.m-side {
  position: relative;
  width: 232px;
  height: calc(100vh - @headerHeight);
  overflow: hidden;
  padding-top: 12px;
  background-color: var(--color-bg-sidebar);
  display: flex;
  flex-direction: column;
  transition: width 0.2s;
  flex-shrink: 0;
  border-right: 1px solid var(--color-border-header);
  .m-side-scroll {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    :deep(.a-menu) {
      border: none !important;
      background: transparent;
    }
    :deep(.a-menu-inline) {
      border-inline-end: none !important;
      background: transparent;
    }
    :deep(.a-menu-submenu) {
      background: transparent;
    }
  }
  &.m-side-collapsed {
    width: 80px;
  }
}
.m-side-trigger {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  cursor: pointer;
  color: var(--color-sidebar-trigger-color, #999);
  font-size: 14px;
  background: var(--color-sidebar-trigger-bg, #fff);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  transition: all 0.2s;
  z-index: 10;
  // 展开时：右下角
  right: 8px;
  bottom: 8px;
  &:hover {
    color: #1890ff;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  }
}
.m-side-collapsed .m-side-trigger {
  // 折叠时：下方居中
  right: auto;
  bottom: 8px;
  left: 50%;
  transform: translateX(-50%);
}
.m-content {
  flex: 1;
  height: calc(100vh - @headerHeight);
  padding: 16px;
  background-color: var(--color-bg-app, #f2f4f7);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  .p-page {
    background-color: var(--color-bg-card, #fff);
    padding: 16px;
    margin-top: 16px;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    &.full-page-host {
      overflow: hidden;
      display: flex;
      flex-direction: column;
    }
  }
  .footer{
    flex-shrink: 0;

   line-height: 30px;
    text-align: center;
    font-size: 12px;
    a{
      text-decoration: none;
    color: var(--color-footer-link, #999);

    }
  }
}
</style>
