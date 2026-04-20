<template>
  <div v-if="i18nReady" :class="['w-full', 'h-full', 'flex', 'flex-col', { 'notice-show': noticeShow }]">
    <notice v-if="noticeShow" :api-url="apiUrl" @show-alert-change="showAlertChange" />
    <Navigation
      class="flex-1"
      navigation-type="top-bottom"
      :need-menu="!!subMenuData?.length"
      @toggle="handleNavToggle"
    >
      <template #side-header>
        <img
          :src="platformConfig.appLogo"
          class="w-[28px] h-[28px] mr-[12px]"
          @click="handleGotoHome"
        />
        <span
          class="text-[16px] text-[#FAFBFD] cursor-pointer"
          @click="handleGotoHome"
        >{{ platformConfig.i18n.productName || platformConfig.productName }}</span
        >
      </template>
      <template #header>
        <FlexRow class="w-full text-[#96A2B9] text-[14px]">
          <template #left>
            <span class="flex items-center text-[14px]">
              <RouterLink
                v-for="item in navData"
                :key="item.routeName"
                :to="{ name: item.routeName, params: item.params }"
                :class="[
                  'px-[16px] text-[#96A2B9]',
                  { 'text-[#fff]': item.routeName === currentMainMenu },
                ]"
              >
                {{ $t(item.title) }}
              </RouterLink>
            </span>
          </template>
          <template #right>
            <div class="flex items-center gap-[8px]">
              <!-- 语言切换 -->
              <Dropdown
                class="mr-[8px]"
                ref="langRef"
                theme="light"
                :popover-options="{
                  clickContentAutoHide: true,
                }">
                <span class="header-icon text-[18px]">
                  <i :class="curLang.icon"></i>
                </span>
                <template #content>
                  <Dropdown.DropdownMenu>
                    <Dropdown.DropdownItem
                      v-for="(item, index) in langs"
                      :key="index"
                      ext-cls="dropdown-item"
                      @click="handleChangeLang(item)"
                    >
                      <i :class="['text-[18px] mr-[3px]', item.icon]"></i>
                      {{item.name}}
                    </Dropdown.DropdownItem>
                  </Dropdown.DropdownMenu>
                </template>
              </Dropdown>
              <!-- 帮助文档 -->
              <Dropdown
                class="mr-[8px]"
                theme="light"
                :popover-options="{
                  clickContentAutoHide: true,
                }">
                <span id="siteHelp" class="header-icon !text-[16px]">
                  <i class="nodeman-icon nc-help-document-fill"></i>
                </span>
                <template #content>
                  <Dropdown.DropdownMenu>
                    <Dropdown.DropdownItem
                      v-for="(item, index) in helpList"
                      :key="index"
                      ext-cls="dropdown-item"
                      @click="handleGotoLink(item)"
                    >
                      {{item.name}}
                    </Dropdown.DropdownItem>
                  </Dropdown.DropdownMenu>
                </template>
              </Dropdown>
              <!-- 用户设置 -->
              <bk-popover
                theme="light"
                :arrow="false"
                placement="bottom-start"
                trigger="click"
              >
                <div class="flex items-center gap-[5px] cursor-pointer">
                  <span>{{ userStore.user?.username }}</span>
                  <angle-up-fill />
                </div>
                <template #content>
                  <ul>
                    <li class="dropdown-item cursor-pointer" @click="logout">{{ t('platform.logout') }}</li>
                  </ul>
                </template>
              </bk-popover>
            </div>
          </template>
        </FlexRow>
      </template>
      <template #menu>
        <BizSelector ref="bizSelectorRef" :expanded="navToggle" />
        <Menu :active-key="String(currentActive)">
          <Menu.Group
            v-for="item in subMenuData"
            :key="item.title"
            :name="$t(item.title)"
          >
            <Menu.Item
              v-for="subItem in item.children"
              :key="subItem.routeName"
              :need-icon="true"
              @click="handleChangeSubMenu(subItem)"
            >
              <template #icon>
                <i
                  v-if="subItem.icon"
                  :class="[
                    subItem.icon,
                    currentActive === subItem.routeName
                      ? 'text-[#3A84FF]'
                      : 'text-[#979BA5]',
                  ]"
                />
              </template>
              {{ $t(subItem.title) }}
            </Menu.Item>
          </Menu.Group>
        </Menu>
      </template>
      <RouterView />
    </Navigation>
    <log-version v-model:is-show="showLog"></log-version>
    <PermissionDialog />
  </div>
</template>

<script setup lang="ts">
import { Dropdown, Menu, Navigation } from 'bkui-vue';
import { AngleUpFill } from 'bkui-vue/lib/icon';
import { debounce } from 'lodash';
import { computed, onBeforeMount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { setDocumentTitle, setShortcutIcon  } from '@blueking/platform-config';
import { useHead } from '@vueuse/head';

import { TopoService } from '@/api/modules/topo';
import { logout } from '@/common/auth';
import { parseCookies } from '@/common/util';
import BizSelector from '@/components/biz-selector.vue';
import Notice from '@/components/notice.vue';
import LogVersion from '@/components/log-version.vue';
import PermissionDialog from '@/components/permission-dialog.vue';
import type { NavItem } from '@/composables/use-menu';
import useMenu from '@/composables/use-menu';
import usePlatform from '@/composables/use-platform';
import { getModuleAuthorizedItems, getPageAuthorizedItems, matchPageAuth, PAGE_AUTH_CONFIG, shouldDeferBizAuthCheck } from '@/constants/auth';

/** 包管理子路由名 → 包类型 id 映射（用于 authorized/verify 传入正确的资源实例 ID） */
const ROUTE_TO_PKG_TYPE_ID: Record<string, string> = {
  agentPackageMng: 'agent',
  proxyPackageMng: 'proxy',
  certPackageMng: 'cert',
  bintoolPackageMng: 'bintool',
  plugin_bintoolPackageMng: 'plugin_bintool',
  pluginPackageMng: 'plugin',
};

/** 根据当前路由名获取包类型 id（仅包管理页面有效） */
function getPkgTypeIdFromRoute(): string | undefined {
  const name = typeof route.name === 'string' ? route.name : undefined;
  return name ? ROUTE_TO_PKG_TYPE_ID[name] : undefined;
}
import { i18nReady } from '@/modules/i18n';
import { useAuthStore } from '@/stores/auth';
import { useMainStore } from '@/stores/main';
import useUserStore from '@/stores/user';

const { t } = useI18n();
const mainStore = useMainStore();
const authStore = useAuthStore();
// 路由信息
const route = useRoute();
const router = useRouter();

// 用户store
const userStore = useUserStore();

// 导航配置
const { navData, subMenuData } = useMenu();

// 蓝鲸平台相关配置hook
const { platformConfig, getPlatformInfo } = usePlatform();
const appName = computed(() => platformConfig.i18n.productName);
const navToggle = ref(false);
const currentMainMenu = computed(() => (route.meta.mainMenu || (route.query.mainMenu as string)));

// 跳转首页
function handleGotoHome() {
  router.push({
    name: 'nodeManager',
  });
}

const ensureCurrentRoutePermission = async () => {
  if (route.name === '403' || route.name === '404') return;

  // 包管理：路由级不做 verify/403 跳转，由 pkg/index.vue 渲染 NoPermission 占位页
  if (route.meta?.mainMenu === 'pkgManager') return;

  // 拓扑管理-操作记录：由 record.vue 渲染 NoPermission 占位页，不触发 verify
  if (route.name === 'record' && route.meta?.mainMenu === 'topoManager') return;

  const matched = matchPageAuth(route, PAGE_AUTH_CONFIG);
  if (!matched) return;
  const isRuleManagerRoute = route.path.includes('rule-manager');

  const hasExplicitlyClearedMultiBusiness = (() => {
    const cachedBizIds = localStorage.getItem('bk_biz_id');
    if (!cachedBizIds) return false;

    try {
      const parsedBizIds = JSON.parse(cachedBizIds);
      return Array.isArray(parsedBizIds) && parsedBizIds.length === 0;
    } catch {
      return false;
    }
  })();
  const persistedStrategyBizId = (() => {
    const cachedStrategyBizId = localStorage.getItem('strategy_biz_id');
    if (!cachedStrategyBizId) return undefined;

    const parsed = Number(cachedStrategyBizId);
    return Number.isNaN(parsed) || parsed <= 0 ? undefined : parsed;
  })();
  let selectedBizIds = isRuleManagerRoute && mainStore.strategyBizId
    ? [mainStore.strategyBizId]
    : [...mainStore.selectedBusinessId];
  let currentBizId = selectedBizIds[0];
  if (
    matched.resourceType === 'biz'
    && mainStore.isBusinessReady
    && (currentBizId === undefined || currentBizId === null)
    && (isRuleManagerRoute || !hasExplicitlyClearedMultiBusiness)
  ) {
    const defaultBizId = isRuleManagerRoute
      ? mainStore.strategyBizId || persistedStrategyBizId || mainStore.businessList[0]?.bk_biz_id
      : mainStore.businessList[0]?.bk_biz_id;
    if (defaultBizId !== undefined && defaultBizId !== null) {
      if (isRuleManagerRoute) {
        mainStore.updateStrategyBizId(defaultBizId);
      } else {
        mainStore.updateCurBusiness([defaultBizId]);
      }
      selectedBizIds = [defaultBizId];
      currentBizId = defaultBizId;
    }
  }

  const bizScope = selectedBizIds.length > 0
    ? selectedBizIds.map(id => String(id))
    : currentBizId ? [String(currentBizId)] : undefined;
  if (shouldDeferBizAuthCheck(matched, mainStore.isBusinessReady, currentBizId)) {
    authStore.refreshPermissions();
    return;
  }

  const pkgTypeId = (matched.resourceType === 'package_type' || matched.resourceType === 'package')
    ? getPkgTypeIdFromRoute() : undefined;

  if (
    authStore.needRefresh
    || authStore.isDifferentBiz(bizScope)
    || !authStore.hasPermissionCache(matched.id, bizScope, pkgTypeId)
    || authStore.isPermissionCacheExpired(matched.id, bizScope, pkgTypeId)
  ) {
    const verified = await authStore.batchVerify([matched], bizScope, pkgTypeId);
    if (!verified) return;
  }

  if (!authStore.hasPermission(matched.id, bizScope, pkgTypeId)) {
    authStore.setDeniedActionIds([matched.id]);
    router.replace({ name: '403' });
  }
};

// 切换通知
const apiUrl = '/api/v3/notice/announcements/current';
const noticeShow = computed(() => mainStore.noticeShow);
function showAlertChange(isShow: boolean) {
  mainStore.updateNoticeShow(isShow);
};

// 切换子菜单
function handleChangeSubMenu(item: Omit<NavItem, 'group'>) {
  router.push({
    name: item.routeName,
    params: item.params,
  });
}

// 切换左侧菜单的展开收起
const handleNavToggle = (value: boolean) => {
  navToggle.value = value;
};

// 业务选择器
const bizSelectorRef = ref<InstanceType<typeof BizSelector>>();
const currentActive = computed(() => route.meta.parentName || route.name);
const loading = ref(false);
const getBusinessList = async () => {
  loading.value = true;
  const res = await TopoService.BusinessList({
    page: {
      limit: 0,
    },
  }, {
    irrevocable: true,
  });
  mainStore.updateBusinessList(res.items);
  loading.value = false;
};
const helpList = computed(() => [
  {
    id: 'DOC',
    name: t('platform.productDoc'),
    href: window.PROJECT_CONFIG.BK_DOCS_CENTER_URL,
  },
  {
    id: 'VERSION',
    name: t('platform.releaseNotes'),
  },
  {
    id: 'FAQ',
    name: t('platform.feedback'),
    href: 'https://bk.tencent.com/s-mart/community',
  },
  {
    id: 'FAQ',
    name: t('platform.openSource'),
    href: window.PROJECT_CONFIG.BKAPP_NAV_OPEN_SOURCE_URL,
  },
]);
/**
 * 系统外链
 */
const helpListRef = ref();
const showLog = ref(false);
function handleGotoLink(item) {
  switch (item.id) {
    case 'DOC':
    case 'FAQ':
      item.href && window.open(item.href);
      break;
    case 'VERSION':
      showLog.value = true;
      break;
  }
  helpListRef.value && helpListRef.value.instance.hide();
}
// 语言
const langs = ref([
  {
    icon: 'nodeman-icon nc-lang-zh-cn',
    name: '中文',
    id: 'zh-CN', // 前端语言标识
  },
  {
    icon: 'nodeman-icon nc-lang-en',
    name: 'English',
    id: 'en-US',
  },
]);
const curLang = computed(() => {
  let currentLang = parseCookies().blueking_language
    ? parseCookies().blueking_language
    : 'zh-cn';

  if (['zh-CN', 'zh-cn', 'cn', 'zhCN', 'zhcn', 'None', 'none'].indexOf(currentLang) > -1) {
    currentLang = 'zh-CN';
  } else {
    currentLang = 'en-US';
  }
  mainStore.updateLanguage(currentLang);
  return langs.value.find(item => item.id === currentLang) || { id: 'zh-CN', icon: 'nodeman-icon nc-lang-zh-cn' };
});
// 切换语言
async function handleChangeLang(item) {
  if (item.id !== curLang.value) {
    const {
      BK_DOMAIN: domain = '',
      BK_TENANT: tenant_id = '',
      BK_USER_WEB_URL: apiBaseUrl = '',
    } = window.PROJECT_CONFIG;

    // URL安全检查
    if (!apiBaseUrl || typeof apiBaseUrl !== 'string') {
      console.error('Invalid apiBaseUrl parameter');
      return;
    }

    // URL消毒：验证协议和格式
    let safeApiBaseUrl = apiBaseUrl;
    try {
      const urlObj = new URL(apiBaseUrl);
      if (urlObj.protocol !== 'http:' && urlObj.protocol !== 'https:') {
        console.error('Invalid URL protocol');
        return;
      }
      safeApiBaseUrl = urlObj.href;
    } catch (error) {
      console.error('Invalid URL format');
      return;
    }

    // 参数消毒：只允许字母数字和下划线
    const safeLanguage = item.id === 'zh-CN' ? 'zh-cn' : 'en';
    if (!safeLanguage) {
      console.error('Invalid language parameter after sanitization');
      return;
    }

    try {
      const baseUrl = safeApiBaseUrl.endsWith('/') ? safeApiBaseUrl.slice(0, -1) : safeApiBaseUrl;
      const url = `${baseUrl}/api/v3/open-web/tenant/current-user/language/`;

      // 添加响应状态检查
      const response = await fetch(url, {
        method: 'PUT',
        headers: {
          'X-Bk-Tenant-Id': tenant_id,
          'Content-type': 'application/json',
        },
        body: JSON.stringify({
          language: safeLanguage,
        }),
        credentials: 'include',
      });

      // 检查HTTP响应状态
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // 设置合理的cookie过期时间（1小时）
      const today = new Date();
      today.setTime(today.getTime() + 1000 * 60 * 60);

      // 安全设置cookie，对值进行编码
      const encodedLang = encodeURIComponent(safeLanguage);
      document.cookie = `blueking_language=${encodedLang};path=/;domain=${domain};expires=${today.toUTCString()}`;

      // 更新HTML lang属性
      document.querySelector('html')?.setAttribute('lang', safeLanguage === 'zh-cn' ? 'zh-CN' : 'en-US');

      // 添加延迟让用户看到切换成功的视觉反馈
      window.location.reload();
    } catch (err) {
      console.error('Language switch failed:', err);
    }
  }
}
// 设置title
watch(appName, () => {
  // https://github.com/vueuse/head
  useHead({
    title: appName.value,
    meta: [
      {
        name: 'description',
        content: '',
      },
    ],
    link: [
      {
        rel: 'icon',
        type: 'image/svg+xml',
        href: () => '/favicon.png',
      },
    ],
  });
});
watch(
  [() => mainStore.isBusinessReady, () => mainStore.selectedBusinessId[0], () => mainStore.strategyBizId, () => route.fullPath],
  () => {
    void ensureCurrentRoutePermission();
  },
  { immediate: true },
);
onBeforeMount(async () => {
  userStore.getUser();
  // 获取平台配置信息
  await getPlatformInfo();
  // 设置文档标题
  setDocumentTitle(platformConfig.i18n);
  // 设置favicon
  setShortcutIcon(platformConfig.favicon);

  await getBusinessList();
  // 首次加载当前路由模块的 authorized items
  // 注：若 route.meta.mainMenu 未就绪则不兜底加载，等 watch(route.meta.mainMenu) 触发时再加载，避免重复请求
  const currentModule = route.meta?.mainMenu ? String(route.meta.mainMenu) : '';
  if (currentModule) {
    const moduleItems = getModuleAuthorizedItems(currentModule);
    if (moduleItems.length) {
      await authStore.fetchAuthorized(moduleItems, currentModule);
    }
  }
  // 初始化业务选择器（恢复收藏、多选业务、排序、策略默认业务）
  bizSelectorRef.value?.init();
  mainStore.setBusinessReady();
});
onMounted(async () => {
  mainStore.updateWindowInnerHeight(window.innerHeight);
  window.addEventListener(
    'resize',
    debounce(() => {
      mainStore.updateWindowInnerHeight(window.innerHeight);
    }, 300),
  );
  // 初始化收藏管控区域（仅首次使用时设置默认值，避免刷新后覆盖用户的收藏）
  if (!localStorage.getItem('collect_workarea')) {
    localStorage.setItem('collect_workarea', JSON.stringify([0]));
  }
});

// 路由切换时增量加载目标模块的 authorized items
watch(
  () => route.meta?.mainMenu,
  async (newModule) => {
    if (!newModule) return;
    const moduleName = String(newModule);
    // biz 类模块首次切入时同时加载 bizSelector
    if (['nodeManager', 'ruleManager'].includes(moduleName)) {
      const bizItems = getModuleAuthorizedItems('bizSelector');
      if (bizItems.length) {
        await authStore.fetchAuthorized(bizItems, 'bizSelector');
      }
    }
    const moduleItems = getModuleAuthorizedItems(moduleName);
    if (moduleItems.length) {
      await authStore.fetchAuthorized(moduleItems, moduleName);
    }
  },
);

// 菜单页切换时按需加载该页所需的非 view 类 action（operate / manage / create / edit / delete / upload）
watch(
  () => route.name,
  async (newName) => {
    if (!newName || typeof newName !== 'string') return;
    const pageItems = getPageAuthorizedItems(newName);
    if (!pageItems.length) return;
    // 以 `page:<routeName>` 作为 key，避免与模块级 items 重复触发
    await authStore.fetchAuthorized(pageItems, `page:${newName}`);
  },
  { immediate: true },
);
</script>
<style lang="postcss" scoped>
.dropdown-item {
  &:hover {
    background-color: #eaf3ff;
    color: #3a84ff;
  }
}
</style>
<style lang="postcss">
body {
  min-width: 1280px;
  overflow-y: auto;
}
.bk-navigation .navigation-container {
  max-width: none !important;
}

.filterTable {
  min-height: 300px;
  .vxe-table--empty-content {
    height: 200px;
    line-height: 200px;
  }
}
.bk-vxe-table .vxe-table--filter-body {
  min-height: 80px;
}

/* ====== 表格设置弹窗 - 旧版蓝鲸样式覆盖 ====== */
.tippy-box[data-theme~='bk-vxe-table-setting-column-theme'] {
  min-width: 390px !important;
  max-width: 390px !important;
  .tippy-content {
    padding: 0 !important;
  }
  /* 设置菜单容器 - 添加标题 */
  .bk-vxe-table-setting-menu {
    padding-bottom: 8px;
    &::before {
      content: '表格设置';
      display: block;
      font-size: 18px;
      font-weight: 700;
      color: #313238;
      line-height: 24px;
      padding: 25px 24px 0;
    }
  }
  /* 隐藏 Tab 栏 */
  .action-tab-wrapper {
    display: none !important;
  }
  /* 字段列表区 - 双列 flex 布局 */
  .field-list-wrapper {
    max-height: 400px !important;
    padding: 0 24px !important;
    margin: 16px 0 0 !important;
    position: relative;
    /* "字段显示设置"小标题 */
    &::before {
      content: '字段显示设置';
      display: block;
      font-size: 14px;
      font-weight: normal;
      color: #63656e;
      margin-bottom: 8px;
      line-height: 20px;
      width: 100%;
    }
    /* 全选 checkbox - 移到右上角（与小标题同行） */
    > span:first-child {
      position: absolute;
      top: 0;
      right: 24px;
      font-size: 14px;
    }
    /* checkbox group 双列 flex-wrap 布局，固定宽度 390px */
    .bk-checkbox-group {
      display: flex !important;
      flex-wrap: wrap;
      width: 100%;
    }
    .field-list-item {
      height: 36px;
      box-sizing: border-box;
      &:nth-child(odd) {
        width: 55%;
      }
      &:nth-child(even) {
        width: 45%;
      }
    }
  }
}
.bk-loading-mask, .bk-loading-indicator {
  z-index: 10 !important;
}
::-webkit-scrollbar {
  width: 14px;
  height: 14px;
}
::-webkit-scrollbar-thumb {
  border-radius: 7px;
  border: 3px solid transparent;
  -webkit-box-shadow: inset 0 0 8px 8px #c4c6cc;
  box-shadow: inset 0 0 8px 8px #c4c6cc;
}
.notice-show {
  .bk-navigation {
    height: calc(100vh - 40px);
  }
}
/* use-auth-lock hook 创建的跟随鼠标锁图标 */
.auth-lock-cursor {
  position: fixed;
  z-index: 999999;
  pointer-events: none;
}
/* 无权限元素置灰文字 */
.unauthorized-text {
  color: #c4c6cc !important;
}
/* 无权限下拉项：置灰但可 hover 带锁 */
.auth-lock-dropdown-item {
  color: #c4c6cc !important;
  cursor: pointer !important;
  pointer-events: auto !important;

  &:hover {
    color: #c4c6cc !important;
    background-color: #f5f7fa !important;
  }
}
/* 管控单元下拉无权限选项：置灰但可 hover 带锁 */
.unauthorized-unit-row {
  color: #c4c6cc !important;
  cursor: pointer !important;
  pointer-events: auto !important;

  &:hover {
    color: #c4c6cc !important;
    background-color: #f5f7fa !important;
  }
}
</style>
