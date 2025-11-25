<template>
  <Navigation
    navigation-type="top-bottom"
    :need-menu="!!subMenuData?.length"
    @toggle="handleNavToggle"
  >
    <template #side-header>
      <img
        src="/nodeman.png"
        class="w-[28px] h-[28px] mr-[12px]"
        @click="handleGotoHome"
      />
      <span
        class="text-[16px] text-[#FAFBFD] cursor-pointer"
        @click="handleGotoHome"
      >{{ t("platform.title") }}</span
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
                { 'text-[#fff]': item.routeName === route.meta.mainMenu },
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
                  <li class="dropdown-item" @click="logout">退出登录</li>
                </ul>
              </template>
            </bk-popover>
          </div>
        </template>
      </FlexRow>
    </template>
    <template #menu>
      <div class="nm-menu-biz mb-[10px]" v-if="isNeedBizSelect">
        <div
          v-show="!navToggle"
          class="w-[30px] h-[30px] text-[12px] bg-[#F0F1F5] m-auto cursor-pointer flex items-center justify-center"
        >
          {{ navBizShrinkText }}
        </div>
        <Select
          v-show="navToggle"
          class="mx-[12px]"
          v-model="business"
          :filter-option="filterOption"
          :show-selected-icon="false"
          multiple
          filterable
          placeholder="全部业务"
          :popover-options="{ boundary: 'document.body', width: '235px' }"
          @change="changeCurBusiness"
          @toggle="handleToggle"
        >
          <Select.Option
            v-for="item in businessList"
            :key="item.bk_biz_id"
            :name="item.bk_biz_name"
            :id="item.bk_biz_id"
          >
            <div class="w-full flex items-center justify-between biz-select-option">
              <span>[{{ item.bk_biz_id }}] {{ item.bk_biz_name }}</span>
              <Button
                text
                @click.native.stop="handleCollect(item.bk_biz_id)">
                <i
                  class="nodeman-icon nc-collect text-[#ffb848] text-[18px]"
                  v-if="collectList.includes(item.bk_biz_id)">
                </i>
                <i
                  class="nodeman-icon nc-not-favorited text-[#63656e] text-[18px] hidden"
                  v-else>
                </i>
              </Button>
            </div>
          </Select.Option>
        </Select>
      </div>
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
</template>

<script setup lang="ts">
import { Button, Dropdown, Menu, Navigation, Select } from 'bkui-vue';
import { AngleUpFill } from 'bkui-vue/lib/icon';
import { debounce, isArray } from 'lodash';
import { computed, onBeforeMount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { useHead } from '@vueuse/head';

import { TopoService } from '@/api/modules/topo';
import { logout } from '@/common/auth';
import { parseCookies, setCookie } from '@/common/util';
import type { NavItem } from '@/composables/use-menu';
import useMenu from '@/composables/use-menu';
import usePlatform from '@/composables/use-platform';
import { useMainStore } from '@/stores/main';
import useUserStore from '@/stores/user';

const { t } = useI18n();
const mainStore = useMainStore();
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
const isNeedBizSelect = computed(() => !route.path.includes('topo-manager') && !route.path.includes('pkg-manager'));

// 收藏
const collectList = ref<number[]>([]);

// 跳转首页
function handleGotoHome() {
  router.push({
    name: 'nodeManager',
  });
}

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

// 业务选择
const business = ref<number[]>([]);
const businessList = computed(() => mainStore.businessList);
const currentActive = computed(() => route.meta.parentName || route.name);
const loading = ref(false);
const getBusinessList = async () => {
  loading.value = true;
  const res = await TopoService.BusinessList({
    page: {
      limit: 0,
    },
  });
  mainStore.updateBusinessList(res.items);
  loading.value = false;
};
const changeCurBusiness = (val: number[]) => {
  mainStore.updateCurBusiness(val);
  localStorage.setItem('bk_biz_id', JSON.stringify(val));
};
// 收起左侧菜单展示的业务的文案
const navBizShrinkText = computed(() => {
  if (!mainStore.selectedBusinessName.length) {
    return '全';
  }
  const len = mainStore.selectedBusinessName.length;
  const text = len > 1 ? len : mainStore.selectedBusinessName[0]?.[0];
  return text;
});
const helpList = [
  {
    id: 'DOC',
    name: t('产品文档'),
    href: window.PROJECT_CONFIG.BK_DOCS_CENTER_URL,
  },
  {
    id: 'VERSION',
    name: t('版本日志'),
  },
  {
    id: 'FAQ',
    name: t('问题反馈'),
    href: 'https://bk.tencent.com/s-mart/community',
  },
  {
    id: 'FAQ',
    name: t('开源社区'),
    href: window.PROJECT_CONFIG.BKAPP_NAV_OPEN_SOURCE_URL,
  },
];
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
  return langs.value.find(item => item.id === currentLang) || { id: 'zh-cn', icon: 'nodeman-icon nc-lang-zh-cn' };
});
// 切换语言
function handleChangeLang(item) {
  if (item.id !== curLang.value) {
    const {
      BK_COMPONENT_API_URL: overwriteUrl = '',
      BK_DOMAIN: domain = '',
    } = window.PROJECT_CONFIG;

    const api = `${overwriteUrl}/api/c/compapi/v2/usermanage/fe_update_user_language/?language=${item.id}`;
    const scriptId = 'jsonp-script';
    const prevJsonpScript = document.getElementById(scriptId);
    if (prevJsonpScript) {
      document.body.removeChild(prevJsonpScript);
    }
    const scriptEl = document.createElement('script');
    scriptEl.type = 'text/javascript';
    scriptEl.src = api;
    scriptEl.id = scriptId;
    document.body.appendChild(scriptEl);

    const today = new Date();
    today.setTime(today.getTime() + 1000 * 60 * 60 * 24);
    document.cookie = `blueking_language=${item.id};path=/;domain=${domain};expires=${today.toUTCString()}`;
    location.reload();
  }
}
// 自定义批量搜索方法
const filterOption = (input: any, options: {id: number, name: string}) => (options.name?.includes(input) || input.includes(options.id));
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
        href: () => '/favicon.svg',
      },
    ],
  });
});
// 收藏
const handleToggle = () => {
  businessList.value.sort((a: Business, b: Business) => {
    // 判断a是否在收藏列表中
    const aIsCollected = collectList.value.includes(a.bk_biz_id);
    // 判断b是否在收藏列表中
    const bIsCollected = collectList.value.includes(b.bk_biz_id);

    // 判断是否为选中的业务
    const aIsSelected = business.value.includes(a.bk_biz_id);
    const bIsSelected = business.value.includes(b.bk_biz_id);

    // 优先级1:选中状态（选中的排在前面）
    if (aIsSelected && !bIsSelected) {
      return -1; // a选中, b未选中 → a在前
    }
    if (!aIsSelected && bIsSelected) {
      return 1; // a未选中, b选中 → b在前
    }

    // 优先级2: 选中状态相同则按收藏状态排序（收藏在前）
    if (aIsCollected && !bIsCollected) {
      return -1; // a收藏, b未收藏 → a在前
    }
    if (!aIsCollected && bIsCollected) {
      return 1; // a未收藏, b收藏 → b在前
    }

    // 优先级3: 选中和收藏状态都相同则按ID从小到大排序
    return a.bk_biz_id - b.bk_biz_id;
  });
};

const handleCollect = (val: number) => {
  if (collectList.value.includes(val)) {
    collectList.value = collectList.value.filter(item => item !== val);
  } else {
    collectList.value.push(val);
  }
  localStorage.setItem('collect', JSON.stringify(collectList.value));
};

onBeforeMount(async () => {
  userStore.getUser();
  getPlatformInfo();
  if (isNeedBizSelect.value) {
    await getBusinessList();
  }
  const bizIdsJson = localStorage.getItem('bk_biz_id');
  if (bizIdsJson) {
    const bizIds = JSON.parse(bizIdsJson);
    business.value = bizIds;
    mainStore.updateCurBusiness(bizIds);
  }
  // 获取收藏业务
  const collectsJson = localStorage.getItem('collect');
  if (collectsJson) {
    const collects = JSON.parse(collectsJson);
    collectList.value = collects;
  }
});
onMounted(async () => {
  mainStore.updateWindowInnerHeight(window.innerHeight);
  window.addEventListener(
    'resize',
    debounce(() => {
      mainStore.updateWindowInnerHeight(window.innerHeight);
    }, 300),
  );
});
</script>
<style lang="postcss" scoped>
.nm-menu-biz {
  :deep(.bk-select-trigger) {
    .bk-input {
      border: none;
    }
    .bk-input--text {
      background: #f0f1f5 !important;
    }
  }
}
.dropdown-item {
  &:hover {
    background-color: #eaf3ff;
    color: #3a84ff;
  }
}
.biz-select-option {
  &:hover {
    .nc-not-favorited {
      display: inline;
    }
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
</style>
