<template>
  <Navigation navigation-type="top-bottom" :need-menu="!!subMenuData?.length" @toggle="handleNavToggle">
    <template #side-header>
      <img src="/nodeman.png" class="w-[28px] h-[28px] mr-[12px]" @click="handleGotoHome" />
      <span class="text-[16px] text-[#FAFBFD] cursor-pointer" @click="handleGotoHome">{{ t('蓝鲸节点管理') }}</span>
    </template>
    <template #header>
      <FlexRow class="w-full text-[#96A2B9] text-[14px]">
        <template #left>
          <span class="flex items-center text-[14px]">
            <RouterLink v-for="item in navData" :key="item.routeName"
              :to="{ name: item.routeName, params: item.params }"
              :class="[
                'px-[16px] text-[#96A2B9]',
                { 'text-[#fff]': item.routeName === route.meta.mainMenu }
              ]"
            >
              {{ $t(item.title) }}
            </RouterLink>
          </span>
        </template>
        <template #right>
          <bk-popover theme="light" :arrow="false" placement="bottom-start" trigger="click">
            <div class="flex items-center gap-[5px] cursor-pointer">
              <span>{{ userStore.user?.username }}</span>
              <angle-up-fill />
            </div>
            <template #content>
              <Button @click="logout">退出登录</Button>
            </template>
          </bk-popover>
        </template>
      </FlexRow>
    </template>
    <template #menu>
      <div class="nm-menu-biz mb-[10px]" v-if="isNeedBizSelect">
        <div
          v-show="!navToggle"
          class="w-[30px] h-[30px] text-[12px] bg-[#F0F1F5] m-auto border-r-[2px] cursor-pointer flex items-center justify-center"
        >{{ navBizShrinkText }}</div>
        <Select
          v-show="navToggle"
          class="mx-[12px]"
          v-model="business"
          @change="changeCurBusiness"
          multiple
          filterable
          placeholder="全部业务"
          :popoverOptions="{ boundary: 'document.body', width: '235px' }">
          <Select.Option v-for="item in businessList" :key="item.bk_biz_id" :name="item.bk_biz_name" :id="item.bk_biz_id">
            [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
          </Select.Option>
        </Select>
      </div>
      <Menu :active-key="String(route.name)">
        <Menu.Group v-for="item in subMenuData" :key="item.title" :name="$t(item.title)">
          <Menu.Item v-for="subItem in item.children" :key="subItem.routeName" :need-icon="true"
            @click="handleChangeSubMenu(subItem)">
            <template #icon>
              <i v-if="subItem.icon" :class="[subItem.icon, route.name === subItem.routeName ? 'text-[#3A84FF]' : 'text-[#979BA5]']" />
            </template>
            {{ $t(subItem.title) }}
          </Menu.Item>
        </Menu.Group>
      </Menu>
    </template>
    <RouterView />
  </Navigation>
</template>

<script setup lang="ts">
import { AngleUpFill } from 'bkui-vue/lib/icon';
import { Menu, Navigation, Select } from 'bkui-vue';
import { computed, onBeforeMount, onMounted, watch, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { useHead } from '@vueuse/head';

import type { NavItem } from '@/composables/use-menu';
import useMenu from '@/composables/use-menu';
import usePlatform from '@/composables/use-platform';
import useUserStore from '@/stores/user';
import { debounce } from 'lodash';
import { useMainStore } from '@/stores/main';
import { TopoService } from '@/api/modules/topo';
import { logout } from '@/common/auth';

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
const business = ref([]);
const businessList = computed(() => mainStore.businessList);
const loading = ref(false);
const getBusinessList = async () => {
  loading.value = true;
  const res = await TopoService.BusinessList({
    page: {
      limit: 0
    },
  });
  mainStore.updateBusinessList(res.items);
  loading.value = false;
};
const changeCurBusiness = (val: Business) => {
  mainStore.updateCurBusiness(business.value);
}
// 收起左侧菜单展示的业务的文案
const navBizShrinkText = computed(() => {
  if (!mainStore.selectedBusinessName.length) {
    return '全';
  }
  const len = mainStore.selectedBusinessName.length;
  const text = len > 1 ? len : mainStore.selectedBusinessName[0]?.[0];
  return text;
});
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

onBeforeMount(() => {
  userStore.getUser();
  getPlatformInfo();
});
onMounted(async () => {
  mainStore.updateWindowInnerHeight(window.innerHeight);
  window.addEventListener('resize', debounce(() => {
    mainStore.updateWindowInnerHeight(window.innerHeight)
  }, 300));
  await getBusinessList();
});
</script>
<style lang="postcss" scoped>
.nm-menu-biz {
  :deep(.bk-select-trigger) {
    .bk-input {
      border: none;
    }
    .bk-input--text {
      background: #F0F1F5 !important;
    }
  }
}
</style>