<template>
  <div ref="configRef" class="p-[24px] w-1/2 min-w-[750px]">
    <Form ref="formRef" :model="formData" :rules="rules">
      <Form.FormItem :label="'配置名称'" property="configpolicy_name" required>
        <Input v-model="formData.configpolicy_name"></Input>
      </Form.FormItem>
      <Form.FormItem :label="'业务'" property="biz_id" required>
        <Select
          v-model="formData.biz_id"
          :disabled="isEdit"
          auto-focus
          filterable
          multiple
          placeholder="选择业务"
        >
          <Select.Option
            v-for="item in businessList"
            :key="item.bk_biz_id"
            :name="item.bk_biz_name"
            :id="item.bk_biz_id"
          >
            [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
          </Select.Option>
        </Select>
      </Form.FormItem>
      <Form.FormItem
        :label="'是否启用'"
        property="enabled"
        required
        v-if="isEdit"
      >
        <Switcher v-model="formData.enabled" theme="primary"></Switcher>
      </Form.FormItem>
      <Form.FormItem :label="'备注'" property="biz_id">
        <Input type="textarea" v-model="formData.remark"></Input>
      </Form.FormItem>
      <Form.FormItem :label="'作用范围'" property="scopes">
        <div
          v-for="(item, index) in formData.scopes"
          :key="index"
          class="bg-[#F0F1F5] p-[16px] flex items-center gap-[12px]"
        >
          <div class="flex-1 flex flex-col gap-[8px]">
            <Select
              v-model="item.bk_networkarea_id"
              prefix="管控区域"
              auto-focus
              filterable
            >
              <Select.Option
                v-for="option in networkAreaList"
                :key="option.bk_networkarea_id"
                :id="option.bk_networkarea_id"
                :name="option.bk_networkarea_name"
              >
                {{ option.bk_networkarea_name }}
              </Select.Option>
            </Select>
            <Select
              v-model="item.bk_networkunit_id"
              prefix="管控区域"
              auto-focus
              filterable
            >
              <Select.Option
                v-for="option in networkUnitList"
                :key="option.bk_networkarea_id"
                :id="option.bk_networkunit_id"
                :name="option.bk_networkunit_name"
              >
                {{ option.bk_networkunit_name }}
              </Select.Option>
            </Select>
            <Select
              v-model="item.os_type"
              prefix="操作系统"
              :list="osTypeList"
              auto-focus
              filterable
            >
            </Select>
            <Select
              v-model="item.cpu_arch"
              prefix="架构"
              :list="cpuArchList"
              auto-focus
              filterable
            >
            </Select>
          </div>
          <div class="w-[20px] cursor-pointer" @click="handleDelete(index)">
            <i class="nodeman-icon nc-delete-3"></i>
          </div>
        </div>
        <Button theme="primary" text @click="handleAdd">
          <i class="nodeman-icon nc-plus-line text-[11px] mr-[8px]"></i>
          <span class="text-[14px]">添加范围</span>
        </Button>
      </Form.FormItem>
      <Form.FormItem :label="'配置'" property="configs">
        <div class="flex items-center text-[#979BA5]">
          <i class="nodeman-icon nc-tips"></i>
          <span class="ml-[9px] text-[12px]"
            >如需修改默认配置，需打开开关后修改</span
          >
        </div>
        <config-template @update-config="updateConfig"></config-template>
        <div class="text-[#E71818] text-[12px] flex items-center" v-if="isEdit">
          <i class="nodeman-icon nc-remind-fill text-[14px]"></i>
          <span class="mr-[3px] ml-[9px]"
            >编辑器内容有改动，保存该配置版本将会由</span
          >
          <Tag theme="warning">V1</Tag>
          <span class="mx-[3px]">升级为</span>
          <Tag theme="success">V2</Tag>
        </div>
      </Form.FormItem>
    </Form>
  </div>
  <div
    class="flex items-center pl-[174px] gap-[12px] h-[50px] fixed bottom-0 w-full bg-[#fff] z-100"
  >
    <Button theme="primary" @click="handleSubmit" class="mr-[8px]">{{
      isEdit ? "保存" : "提交"
    }}</Button>
    <Button @click="handleCancel">取消</Button>
  </div>
</template>
<script lang="ts" setup>
import { ref, reactive, computed, onMounted, watch } from "vue";
import { Form, Select, Input, Button, Switcher, Tag } from "bkui-vue";
import type { ConfigPolicyCreateReq } from "@/@types/configpolicy";
import { cloneDeep } from "lodash";
import { useMainStore } from "@/stores/main";
import { useRoute, useRouter } from "vue-router";
import { TopoService } from "@/api/modules/topo";
import { ConfigPolicyAPIService } from "@/api/modules/configpolicy";
import useUserStore from '@/stores/user';

const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const userStore = useUserStore();
const nodeRole = computed(() => route.params.node_role);
const isEdit = computed(() => route.name === "editConfig");
const businessList = computed(() => mainStore.businessList);
const initData = {
  configpolicy_name: '',
  node_role: nodeRole.value,
  biz_id: [],
  remark: '',
  scopes: [
    {
      bk_networkarea_id: -1,
      bk_networkunit_id: -1,
      os_type: "-1",
      cpu_arch: "-1",
    },
  ],
  configs: [] as ConfigPolicyConfigBlock[],
  operator: '',
  enabled: false,
};
const formData = reactive(cloneDeep(initData));
const configpolicyId = ref();
const rules = {};
const handleAdd = () => {
  formData.scopes.push({
    bk_networkarea_id: -1,
    bk_networkunit_id: -1,
    os_type: "-1",
    cpu_arch: "-1",
  });
};
const handleDelete = (index: number) => {
  formData.scopes = formData.scopes.filter(
    (_: any, ind: number) => ind !== index
  );
};
const updateConfig = (configs: ConfigPolicyConfigBlock[]) => {
  formData.configs = configs
}
const handleSubmit = async () => {
  let res;
  if(isEdit.value) {
    res = await ConfigPolicyAPIService.ConfigPolicyUpdate({
      configpolicy_id: configpolicyId.value,
      ...formData,
      operator: userStore.user?.username
    }).catch(() => false);
  } else {
    res = await ConfigPolicyAPIService.ConfigPolicyCreate({
      ...formData,
      operator: userStore.user?.username
    }).catch(() => false);
  }
  if (res && nodeRole.value) {
    router.replace({name: `${nodeRole.value}Strategy`});
  }
};
const handleCancel = () => {
  router.replace({ name: "agentStrategy" });
};
const networkAreaList = ref<NetworkArea[]>([]);
// 管控全域
const getNetworkAreaList = async () => {
  const res = await TopoService.NetworkAreaList({}).catch(() => ({
    total: 0,
    items: [],
  }));
  networkAreaList.value = [
    {
      bk_networkarea_id: -1,
      bk_networkarea_name: "不限",
      tenant_id: "single",
      cloud_vendor: "",
    },
    ...res.items,
  ];
};
// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitList({}).catch(() => ({
    total: 0,
    items: [],
  }));
  networkUnitList.value = [
    {
      bk_networkunit_id: -1,
      bk_networkunit_name: "不限",
      tenant_id: "single",
    } as NetworkUnit,
    ...res.items,
  ];
};
// 操作系统和架构
const osTypeList = ref<{ value: string; label: string }[]>([
  {
    value: "-1",
    label: "不限",
  },
]);
const cpuArchList = ref<{ value: string; label: string }[]>([
  {
    value: "-1",
    label: "不限",
  },
]);
const getPlatform = async () => {
  const res = await ConfigPolicyAPIService.ConfigPolicyListPlatform({
    node_role: nodeRole.value,
  }).catch(() => ({
    os_type: [],
    cpu_arch: [],
  }));
  osTypeList.value.splice(
    1,
    0,
    ...res.os_type.map((item) => ({
      value: item,
      label: item,
    }))
  );
  cpuArchList.value.splice(
    1,
    0,
    ...res.cpu_arch.map((item) => ({
      value: item,
      label: item,
    }))
  );
};
watch(
  () => route.name,
  () => {
    if (route.name === "editConfig" && mainStore.configEditData) {
      configpolicyId.value = mainStore.configEditData.configpolicy_id;
      Object.assign(formData, mainStore.configEditData);
    }
  },
  { immediate: true }
);
onMounted(async () => {
  await getNetworkAreaList();
  await getNetworkUnitList();
  await getPlatform();
});
</script>
