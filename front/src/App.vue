<template>
  <Navigation navigation-type="top-bottom" :need-menu="!!subMenuData?.length">
    <template #side-header>
      <img src="/nodeman.png" class="w-[28px] h-[28px] mr-[12px]" @click="handleGotoHome" />
      <span
        class="text-[16px] text-[#FAFBFD] cursor-pointer"
        @click="handleGotoHome">{{ t('蓝鲸节点管理') }}</span>
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
                { 'text-[#fff]': item.routeName === route.meta.mainMenu }
              ]">
              {{ $t(item.title) }}
            </RouterLink>
          </span>
        </template>
        <template #right>
          <bk-popover
            theme="light"
            :arrow="false"
            offset="-20, 10"
            placement="bottom-start"
            trigger="click">
            <div>
              {{ userStore.user?.username }}
            </div>
            <template #content>
            </template>
          </bk-popover>
        </template>
      </FlexRow>
    </template>
    <template #menu>
      <Menu :active-key="String(route.name)">
        <Menu.Group v-for="item in subMenuData" :key="item.title" :name="$t(item.title)">
          <Menu.Item
            v-for="subItem in item.children"
            :key="subItem.routeName"
            :need-icon="true"
            @click="handleChangeSubMenu(subItem)">
            <template #icon>
              <i v-if="subItem.icon" :class="subItem.icon" />
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
import { Menu, Navigation } from 'bkui-vue';
import { computed, onBeforeMount, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { useHead } from '@vueuse/head';

import type { NavItem } from '@/composables/use-menu';
import useMenu from '@/composables/use-menu';
import usePlatform from '@/composables/use-platform';
import useUserStore from '@/stores/user';
import { debounce } from 'lodash';
import { useMainStore } from '@/stores/main';

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
onMounted(() => {
  mainStore.updateWindowInnerHeight(window.innerHeight);
  window.addEventListener('resize', debounce(() => {
    mainStore.updateWindowInnerHeight(window.innerHeight)
  },300));
});
</script>