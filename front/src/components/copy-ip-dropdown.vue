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
  // 任务详情模式：所有IP 下多一级状态（全部/成功/失败/超时），仅任务详情页传 true
  hasStatusLevel: {
    type: Boolean,
    default: false,
  },
});

const { t } = useI18n();

// 4 个 IP 类型项（IPv4 / IPv6 / 管控区域+IPv4 / 管控区域+IPv6）
const buildIpTypeChildren = () => [
  { id: 'ipv4', name: 'IPv4' },
  { id: 'ipv6', name: 'IPv6' },
  { id: 'workarea+ipv4', name: t('components.copyIpDropdown.workareaAndIpv4') },
  { id: 'workarea+ipv6', name: t('components.copyIpDropdown.workareaAndIpv6') },
];

// 4 个状态项（全部 / 成功 / 失败 / 超时），每项下挂 IP 类型
const buildAllStatusChildren = () => [
  { id: 'all', name: t('platform.nodeMan.taskDetail.status.all'), children: buildIpTypeChildren() },
  { id: 'success', name: t('platform.nodeMan.taskDetail.status.success'), children: buildIpTypeChildren() },
  { id: 'failed', name: t('platform.nodeMan.taskDetail.status.failed'), children: buildIpTypeChildren() },
  { id: 'timeout', name: t('platform.nodeMan.taskDetail.status.timeout'), children: buildIpTypeChildren() },
];

// 第一级菜单：勾选IP / 所有IP
const selectList = [
  {
    id: 'select',
    name: t('components.copyIpDropdown.checkIp'),
    children: buildIpTypeChildren(),
  },
  {
    id: 'all',
    name: t('components.copyIpDropdown.allIps'),
    children: props.hasStatusLevel ? buildAllStatusChildren() : buildIpTypeChildren(),
  },
];

const copylist = computed(() => {
  const list = props.list.length ? props.list : selectList;
  list.forEach((item: any) => {
    if (item.id === 'select') item.disabled = props.disabled;
    else if (item.id === 'all') item.disabled = false;
  });
  return list;
});

const area = ref<string[]>([]);

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
    crossPageSelectionData.value = res.items.filter((item: string) => !!(type.includes('workarea') ? item.split(':')[1] : item)).map((item: string) => {
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

// 构造复制内容
const buildCopyContent = (list: any[], type: string) => list
  .map((item: any) => {
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
    return '';
  })
  .filter((v: string) => !!v);

// 检查列表中是否有目标 IP 类型
const hasTargetIpType = (list: any[], type: string) => {
  if (['ipv4', 'workarea+ipv4'].includes(type)) {
    return list.some((item: any) => !!(item.bk_host_innerip || item.bk_host_inner));
  }
  return list.some((item: any) => !!item.bk_host_innerip_v6);
};

// 更换选择项
const handleChange = async () => {
  if (area.value.length === 0) return;
  const scope = area.value[0];

  // 任务详情模式（hasStatusLevel）：
  //   勾选IP: area = ['select', '<ipType>']
  //   所有IP: area = ['all', '<status>', '<ipType>']
  // 默认模式：
  //   勾选IP/所有IP: area = ['<scope>', '<ipType>']
  const type = props.hasStatusLevel
    ? (scope === 'select' ? area.value[1] : area.value[2])
    : area.value[1];
  const status = props.hasStatusLevel && scope === 'all' ? area.value[1] : 'all';

  if (!type) return;
  if (props.hasStatusLevel && scope === 'all' && !status) return;

  // 准备基础数据
  let copyData: any[] = props.isCrossPageSelection ? crossPageSelectionData.value : props.data;

  // 跨页全选场景：按 IP 类型拉数据
  if (props.isCrossPageSelection) {
    await getCorssPageIps(type);
    copyData = crossPageSelectionData.value;
  }

  // 按 scope / status 过滤
  let list: any[];
  if (scope === 'select') {
    list = copyData.filter((item: any) => item.checked);
  } else if (status !== 'all') {
    list = copyData.filter((item: any) => item.state === status);
  } else {
    list = copyData;
  }

  if (list.length === 0 || !hasTargetIpType(list, type)) {
    Message({ theme: 'primary', message: t('components.copyIpDropdown.empty') });
    return;
  }

  const copyContent = buildCopyContent(list, type);

  if (copyContent.length === 0) {
    Message({ theme: 'primary', message: t('components.copyIpDropdown.empty') });
    return;
  }

  if (isSupported) {
    try {
      await copy(copyContent.join('\n'));
      const ipLabel = type.includes('ipv4') ? 'IPv4' : 'IPv6';
      Message({
        theme: 'success',
        message: `${t('components.copyIpDropdown.success')}（${copyContent.length} 个${ipLabel}）`,
      });
    } catch (error) {
      Message({ theme: 'error', message: t('components.copyIpDropdown.failed') });
    }
  } else {
    Message({ theme: 'error', message: t('components.copyIpDropdown.notSupport') });
  }
};

const handleToggle = (value: boolean) => {
  if (!value) area.value = []; // 关闭弹出面板时 清空选项
};
</script>
<style scoped>
.bk-cascader-wrapper {
  width: 86px;
}
</style>
