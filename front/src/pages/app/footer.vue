<template>
  <footer class="w-full h-[50px] text-center text-[12px] leading-[12px] flex flex-col justify-center absolute bottom-0">
    <!-- <div v-if="platformConfig.i18n.footerInfoHTML">
      <p class="mb-[8px] text-[#3A84FF]" v-html="platformConfig.i18n.footerInfoHTML"></p>
      <p class="text-[#979ba5]">
          {{ platformConfig.footerCopyrightContent }}
      </p>
    </div> -->
    <ul class="flex items-center justify-center mb-[8px]">
      <li v-for="item, index in links" :key="index">
        <a :href="item.href" :target="item.target" class="text-[#3A84FF]">
          {{ item.name }}
        </a>
        <span v-if="index < 1" class="mx-[5px]">|</span>
      </li>
    </ul>
    <p class="text-[#979ba5]">{{ Copyright }}</p>
  </footer>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const links = [
  {
    name: t('联系BK助手'),
    href: 'https://bk.tencent.com/s-mart/community',
    target: '_blank',
  },
  {
    name: t('蓝鲸桌面'),
    href: 'https://bk.tencent.com/s-mart/desktop',
    target: '_blank',
  },
];
const version = ref('V1.1.1');
const Copyright = computed(() => `Copyright © 2012 Tencent BlueKing. All Rights Reserved. ${version.value}`);
</script>
