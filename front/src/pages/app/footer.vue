<template>
  <footer class="w-full h-[70px] text-center text-[12px] leading-[12px] flex flex-col justify-center absolute bottom-0">
    <div v-if="platformConfig.i18n.footerInfoHTML" class="py-[19px]">
      <p class="mb-[8px] text-[#3A84FF]" v-html="platformConfig.i18n.footerInfoHTML"></p>
      <p class="text-[#979ba5]">
        {{ platformConfig.footerCopyrightContent }}
      </p>
    </div>
    <div v-else>
      <ul class="flex items-center justify-center mb-[8px]">
        <li v-for="item, index in links" :key="index">
          <a :href="item.href" :target="item.target" class="text-[#3A84FF]">
            {{ item.name }}
          </a>
          <span v-if="index < 1" class="mx-[5px]">|</span>
        </li>
      </ul>
      <p class="text-[#979ba5]">{{ Copyright }}</p>
    </div>
  </footer>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import usePlatform from '@/composables/use-platform';

const { t } = useI18n();
const { platformConfig } = usePlatform();

const links = [
  {
    name: t('footer.linksBk'),
    href: 'https://bk.tencent.com/s-mart/community',
    target: '_blank',
  },
  {
    name: t('footer.desktop'),
    href: 'https://bk.tencent.com/s-mart/desktop',
    target: '_blank',
  },
];
const version = ref('V1.1.1');
const Copyright = computed(() => `Copyright © 2012 Tencent BlueKing. All Rights Reserved. ${version.value}`);
</script>
