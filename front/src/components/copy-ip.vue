<template>
  <Cascader
    v-model="data"
    :list="list"
    :scroll-height="136"
    trigger="click"
    @change="handleChange"
    @toggle="handleToggle">
    <template #trigger>
      <Button>
        <span>{{ $t('action.copy') }}</span>
        <i class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"></i>
      </Button>
    </template>
  </Cascader>
</template>

<script lang="ts" setup>
import { Button, Cascader, Message } from 'bkui-vue';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { useClipboard } from '@vueuse/core';
const props = defineProps({
  getSelectData: {
    type: Function,
    required: true,
  },
});
const { t } = useI18n();

const subList = [
  {
    id: 'ipv4',
    name: 'IPv4',
  },
  {
    id: 'ipv6',
    name: 'IPv6',
  },
  {
    id: 'workarea+ipv4',
    name: `${t('topoManager.workArea.copy.workarea')}+IPv4`,
  },
  {
    id: 'workarea+ipv6',
    name: `${t('topoManager.workArea.copy.workarea')}+IPv6`,
  },
];

const list = [
  {
    id: 'select',
    name: t('topoManager.workArea.copy.select'),
    children: subList,
  },
  {
    id: 'all',
    name: t('topoManager.workArea.copy.selectAll'),
    children: subList,
  },
];

const handleChange = (values: string[]) => {
  if (!values[0] && !values[1]) return; // 若清空数据 不触发emit
  const type = values[0];
  const value = values[1];
  const tableData = props.getSelectData(type);

  // 子组件进行(IPv4/IPv6/管控区域+IPv4/管控区域+IPv6)处理并复制
  const copyList = getCopyValueBySubList(tableData, value);
  copyText(copyList.join(',\n'));
};

const data = ref([]);
const handleToggle = (value: boolean) => {
  if (!value) data.value = []; // 关闭弹出面板时 清空data
};

const getCopyValueBySubList = (list: Array<any>, id: string): Array<any> => {
  let result: any[] = [];
  switch (id) {
    case 'ipv4':
      result = list.map(item => item.ipv4);
      break;
    case 'ipv6':
      result = list.map(item => item.ipv6);
      break;
    case 'workarea+ipv4':
      result = list.map(item => `${item.bk_networkarea_name}+${item.ipv4}`);
      break;
    case 'workarea+ipv6':
      result = list.map(item => `${item.bk_networkarea_name}+${item.ipv6}`);
      break;
  }
  return result;
};

const copyText = (value: string) => {
  const { copy } = useClipboard({
    legacy: true, // 使用 execCommand 作为后备处理副本
  });
  try {
    copy(value);
    Message({
      theme: 'success',
      message: t('topoManager.workArea.copy.success'),
    });
  } catch (error) {
    console.error(error);
    Message({
      theme: 'success',
      message: t('topoManager.workArea.copy.failed'),
    });
  }
};

</script>
