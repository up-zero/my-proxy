/// <reference types="vite/client" />

interface ImportMeta {
  readonly env: Record<string, string>
}

import 'vue-router';
declare module 'vue-router' {
  interface RouteMeta {
    isMenu?: boolean;
    fullPage?: boolean;
    hidden?: boolean;
    titleKey?: string;
    icon?: string;
  }
}
