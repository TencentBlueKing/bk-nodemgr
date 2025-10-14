<template>
  <Cascader
    v-model="area"
    :list="copylist"
    :scroll-height="134"
    trigger="click"
    @change="handleChange"
    @toggle="handleToggle"
  >
    <template #trigger>
      <Button>
        <span>复制</span>
        <i
          class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
        ></i>
      </Button>
    </template>
  </Cascader>
</template>

<script lang="ts" setup>
import { Button, Cascader, Message } from 'bkui-vue';
import { cloneDeep } from 'lodash';
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';

import { useClipboard } from '@vueuse/core';

const props = defineProps({
  type: {
    type: String,
    default: 'agent',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  data: {
    type: Array,
    default: () => [],
  },
  filterProp: {
    type: String,
    default: '',
  },
  list: {
    type: Array,
    default: () => [
      {
        id: 'select',
        name: '勾选IP',
        disabled: true,
        children: [
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
            name: '管控区域+IPv4',
          },
          {
            id: 'workarea+ipv6',
            name: '管控区域+IPv6',
          },
        ],
      },
      {
        id: 'all',
        name: '所有IP',
        children: [
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
            name: '管控区域+IPv4',
          },
          {
            id: 'workarea+ipv6',
            name: '管控区域+IPv6',
          },
        ],
      },
    ],
  },
});
const copylist = ref(cloneDeep(props.list));
const area = ref([]); // ['all', 'ipv4']
// 使用 useClipboard 处理剪贴板操作
const { copy, isSupported } = useClipboard({ legacy: true });
// 更换选择项
const handleChange = async () => {
  if (area.value.length === 0) return;
  const type = area.value[1];
  const list = area.value[0] === 'all'
    ? props.data
    : props.data.filter((item: any) => item.checked
            && (!props.filterProp || item[props.filterProp] === area.value[0]));
  if (
    list.every((item: any) => (['ipv4', 'workarea+ipv4'].includes(type)
      ? !(item.bk_host_innerip || item.bk_host_inner)
      : !item.bk_host_innerip_v6))
  ) {
    Message({
      theme: 'primary',
      message: '没有可复制的内容',
    });
    return;
  }
  const copyContent = list.map((item: any) => {
    switch (type) {
      case 'ipv4':
        return item.bk_host_innerip || item.bk_host_inner;
      case 'ipv6':
        return item.bk_host_innerip_v6;
      case 'workarea+ipv4':
        return `${item.bk_networkarea_id}:${item.bk_host_innerip}`;
      case 'workarea+ipv6':
        return `${item.bk_networkarea_id}:${item.bk_host_innerip_v6}`;
    }
  });
  if (isSupported) {
    try {
      await copy(copyContent.join(',\n'));
      Message({
        theme: 'success',
        message: '复制成功',
      });
    } catch (error) {
      Message({
        theme: 'error',
        message: '复制失败',
      });
    }
  } else {
    Message({
      theme: 'error',
      message: '当前环境不支持剪贴板操作',
    });
  }
};
const handleToggle = (value: boolean) => {
  if (!value) area.value = []; // 关闭弹出面板时 清空选项
};
watch(
  () => props.disabled,
  (val: boolean) => {
    copylist.value.forEach((item: any) => (item.disabled = item.id === 'all' ? false : val));
  },
  { immediate: true },
);
</script>
