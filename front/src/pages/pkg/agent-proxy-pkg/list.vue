<template>
  <PageHeader
    class="w-full sticky top-0 z-1"
    :title="t('Agent 包管理')"
    :back="false"
  >
    <Tag radius="14px" class="ml-[20px]">当前版本：{{ curAgentVersion }}</Tag>
  </PageHeader>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button theme="primary">包上传</Button>
      <SearchSelect
        class="ml-[16px] flex-1"
        ref="searchSelect"
        :data="searchSelectData"
        v-model="searchSelectValue"
        :uniqueSelect="true"
        :placeholder="t('版本号、操作系统/架构、标签、上传用户、状态')"
        @update:modelValue="handleSearchSelectChange">
      </SearchSelect>
    </div>
    <div class="flex flex-1">
      <div class="w-[240px] bg-[#fff] rounded-[2px] shadow-[0_2px_4px_#1919290d] h-full">
        <div class="px-[16px] pt-[10px]">{{ t('快捷筛选') }}</div>
      </div>
      <bk-loading title="数据加载中" :loading="loading">
        <!-- <Table
          :data="packageList"
          :empty-text="'暂无数据'"
          :pagination="pagination"
          :column-config="{ resizable: true }"
          show-overflow-tooltip
          :max-height="maxHeight"
          :show-settings="isShowSetting"
          :settings="settings"
          @setting-change="handleSettingChange"
          @checkbox-change="handleSelectChange"
          @checkbox-all="handleSelectAllChange"
          @column-filter="handleFilter"
        >
          <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
          <TableColumn field="bk_host_innerip" :title="t('platform.nodeMan.inner_ip')" width="150" fixed="left"></TableColumn>
          <TableColumn field="bk_host_innerip_v6" :title="t('platform.nodeMan.inner_ipv6')" width="150"></TableColumn>
          <TableColumn field="bk_agent_id" :title="t('platform.nodeMan.agentId')" width="295"></TableColumn>
          <TableColumn
            field="bk_networkarea_name"
            :title="t('platform.nodeMan.bk_cloud_name')"
            :filter="areaFilterOption"
          ></TableColumn>
          <TableColumn field="bk_networkunit_id" :title="t('platform.nodeMan.bk_cloud_unit')" :filter="filterOptionSource.bk_networkunit_id"></TableColumn>
          <TableColumn field="os_type" :title="t('platform.nodeMan.os_type')" :filter="filterOptionSource.os_type"></TableColumn>
          <TableColumn field="node_version" :title="t('platform.nodeMan.agent_version')" :filter="filterOptionSource.node_version"></TableColumn>
        </Table> -->
      </bk-loading>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { Button, SearchSelect, Tag } from 'bkui-vue';
import useTableSetting from '@/composables/use-table-setting';

const { t } = useI18n();
const curAgentVersion = ref('v2.2.6-beta.30');
const loading = ref(false);
const packageList = ref([]);
// 搜索
const searchSelectValue = ref<{id: string, name: string, values: any[]}[]>([]);
const getUniqueChildren = (prop: string) => {
  const uniqueValues = Array.from(new Set(packageList.value.map((item: any) => item[prop]).filter((item: any) => item)));
  return uniqueValues.map(value => ({
    id: value,
    name: String(value),
  }))
}
const searchSelectData = computed(() => [
  // {id: 'bk_host_innerip', name: t('platform.nodeMan.inner_ip'), multiple: true},
  // {id: 'bk_host_innerip_v6', name: t('platform.nodeMan.inner_ipv6'), multiple: true},
  // {id: 'bk_networkarea_id', name: '管控区域ID:IP', children: getUniqueChildren('bk_networkarea_id'), multiple: true},
  // {id: 'bk_agent_id', name: 'Agent ID', multiple: true},
  // // {id: 'bk_networkarea_name', name: '管控区域', children: getUniqueChildren('bk_networkarea_name')},
  // {id: 'bk_networkunit_id', name: '管控单元', children: getUniqueChildren('bk_networkunit_id'), multiple: true},
  // {id: 'os_type', name: '操作系统', children: getUniqueChildren('os_type'), multiple: true},
  // {id: 'node_version', name: 'Agent版本', children: getUniqueChildren('node_version'), multiple: true},
  // {id: 'node_status', name: 'Agent 状态', children: getUniqueChildren('node_status'), multiple: true},
]);

// // 表格
// const { isShowSetting, settings, handleSettingChange } = useTableSetting({
//   checked: [
//     'bk_host_innerip',
//     'bk_host_innerip_v6',
//     'bk_agent_id',
//     'bk_networkarea_name',
//     'bk_networkarea_id',
//     'bk_networkunit_id',
//     'os_type',
//     'node_version',
//     'node_status',
//     'action',
//   ],
//   disabled: ['action'],
// });


const handleSearchSelectChange = () => {

}
</script>