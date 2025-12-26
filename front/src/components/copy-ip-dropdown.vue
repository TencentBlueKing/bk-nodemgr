<template>
  <Cascader
    v-model="area"
    :list="copylist"
    :scroll-height="136"
    trigger="click"
    @change="handleChange"
    @toggle="handleToggle"
  >
    <template #trigger>
      <Button :loading="crossPageSelectLoading">
        <span>{{ $t('components.copyIpDropdown.copy') }}</span>
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
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { useClipboard } from '@vueuse/core';

import { TopoService } from '@/api/modules/topo';

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
    default: () => [],
  },
  isCrossPageSelection: {
    type: Boolean,
    default: false,
  },
  crossPageQueryParams: {
    type: Object,
    default: () => ({}),
  },
});

const { t } = useI18n();
const selectList = [
  {
    id: 'select',
    name: t('components.copyIpDropdown.checkIp'),
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
        name: t('components.copyIpDropdown.workareaAndIpv4'),
      },
      {
        id: 'workarea+ipv6',
        name: t('components.copyIpDropdown.workareaAndIpv6'),
      },
    ],
  },
  {
    id: 'all',
    name: t('components.copyIpDropdown.allIps'),
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
        name: t('components.copyIpDropdown.workareaAndIpv4'),
      },
      {
        id: 'workarea+ipv6',
        name: t('components.copyIpDropdown.workareaAndIpv6'),
      },
    ],
  },
];
const copylist = computed(() => {
  const list = props.list.length ? props.list : selectList;
  list.forEach((item: any) => (item.disabled = item.id === 'all' ? false : props.disabled));
  return list;
});
const area = ref([]); // ['all', 'ipv4']

// 跨页全选的数据
const crossPageSelectionData = ref<any[]>([]);
const crossPageSelectLoading = ref(false);
const getCorssPageIps = async (type: string) => {
  let serve = TopoService.HostSelectInnerIP;
  switch (type) {
    case 'ipv4':
      serve = TopoService.HostSelectInnerIP;
      break;
    case 'ipv6':
      serve = TopoService.HostSelectInnerIPV6;
      break;
    case 'workarea+ipv4':
      serve = TopoService.HostSelectNetWorkareaIDAndInnerIP;
      break;
    case 'workarea+ipv6':
      serve = TopoService.HostSelectNetWorkareaIDAndInnerIPV6;
      break;
  }
  try {
    crossPageSelectLoading.value = true;
    const res = await serve(props.crossPageQueryParams);
    crossPageSelectionData.value = res.items.filter(item => !!(type.includes('workarea') ? item.split(':')[1] : item)).map((item: string) => {
      let bk_host_innerip = '';
      let bk_host_innerip_v6 = '';
      let bk_networkarea_id = '';
      if (type === 'ipv4') {
        bk_host_innerip = item;
      } else if (type === 'ipv6') {
        bk_host_innerip_v6 = item;
      } else if (type === 'workarea+ipv4') {
        bk_networkarea_id = item.split(':')[0];
        bk_host_innerip = item.split(':')[1];
      } else if (type === 'workarea+ipv6') {
        bk_networkarea_id = item.split(':')[0];
        bk_host_innerip_v6 = item.split(':')[1];
      }
      return {
        checked: true,
        bk_host_innerip,
        bk_host_innerip_v6,
        bk_networkarea_id,
      };
    });
  } catch (error) {
    console.error('获取跨页全选数据失败:', error);
  } finally {
    crossPageSelectLoading.value = false;
  }
};

// 使用 useClipboard 处理剪贴板操作
const { copy, isSupported } = useClipboard({ legacy: true });
// 更换选择项
const handleChange = async () => {
  if (area.value.length === 0) return;

  const type = area.value[1];
  if (props.isCrossPageSelection) {
    await getCorssPageIps(type);
  }
  const copyData = props.isCrossPageSelection ? crossPageSelectionData.value : props.data;
  const list = area.value[0] === 'all'
    ? copyData
    : copyData.filter((item: any) => item.checked
            && (!props.filterProp || item[props.filterProp] === area.value[0]));
  if (
    list.every((item: any) => (['ipv4', 'workarea+ipv4'].includes(type)
      ? !(item.bk_host_innerip || item.bk_host_inner)
      : !item.bk_host_innerip_v6))
  ) {
    Message({
      theme: 'primary',
      message: t('components.copyIpDropdown.empty'),
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
        message: `${t('components.copyIpDropdown.success')}（${copyContent.length} 个${type?.includes('ipv4') ? 'IPv4' : 'IPv6'}）`,
      });
    } catch (error) {
      Message({
        theme: 'error',
        message: t('components.copyIpDropdown.failed'),
      });
    }
  } else {
    Message({
      theme: 'error',
      message: t('components.copyIpDropdown.notSupport'),
    });
  }
};
const handleToggle = (value: boolean) => {
  if (!value) area.value = []; // 关闭弹出面板时 清空选项
};
</script>
