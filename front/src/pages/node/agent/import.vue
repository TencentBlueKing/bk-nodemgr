<template>
    <div class="setup pt-[24px] pb-[48px]">
        <div class="mx-[24px] flex min-h-[56px] bg-[#F0F8FF] border border-[#C5DAFF] rounded-[2px] py-[6px] px-[9px] gap-[9px]">
            <i class="nodeman-icon nc-tips pt-[2px] text-[#3A84FF]"></i>
            <div class="text-[#4d4f56] text-[12px] leading-[20px] text-left">
                <i18n-t keypath="platform.nodeMan.installAgentPage.tip1" tag="p">
                    <span class="text-[#313238] font-bold">{{ $t('platform.nodeMan.installAgentPage.tip1FirstSlotText') }}</span>
                    <span class="text-[#313238] font-bold">{{ $t('platform.nodeMan.installAgentPage.tip1SecondSlotText') }}</span>
                </i18n-t>
                <i18n-t keypath="platform.nodeMan.installAgentPage.tip2" tag="p">
                    <span class="text-[#313238] font-bold">{{ $t('platform.nodeMan.installAgentPage.tip2FirstSlotText') }}</span>
                    <Button text theme="primary" @click="handleShowPanel">{{
                        $t('platform.nodeMan.installAgentPage.tip2SecondSlotText') }}</Button>
                    <Button text theme="primary" @click="handleShowSetting">{{
                        $t('platform.nodeMan.installAgentPage.tip2ThirdSlotText') }}</Button>
                </i18n-t>
            </div>
        </div>
        <div class="m-[24px]">
            <Form ref="formRef" :model="formData" :rules="rules">
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.type')" required>
                    <install-type :needTypeList="['setup', 'manual']"></install-type>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.info')" required>
                    <install-table ref="installTableRef" v-model:data="formData.info"></install-table>
                </Form.FormItem>
            </Form>
        </div>
        <div :class="['h-[48px] w-full flex items-center pl-[174px]', { 'fixed bottom-[0] bg-[#fff] z-[100]': isAtBottom }]" ref="footerRef">
            <Button class="w-[100px] mr-[8px]" theme="primary" @click="handlePreview">{{ $t('去安装') }}</Button>
            <Button class="w-[88px]">{{ $t('取消') }}</Button>
        </div>
        <preview v-model:is-show="previewData.isShow"></preview>
    </div>
</template>
<script lang="ts" setup>
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue';
import { Button, Form, Select, Input, Upload, Message } from 'bkui-vue';
import { Table, TableColumn } from '@blueking/table';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { debounce } from 'lodash';
import { useMainStore } from '@/stores/main';
import { computed } from 'vue';
import { TopoService } from '@/api/modules/topo';
import { useRoute } from 'vue-router';
import Preview from './preview.vue';
import { cloneDeep } from 'lodash';
import type { AgentInstallInfo } from '@/@types/node_agent.d';
import { useNodeManageStore } from '@/stores/node-manage';

const route = useRoute();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const showRightPanel = ref(false);
const initData = {
    bk_addressing: 'static',
    bk_host_innerip: '',
    bk_host_innerip_v6: '',
    os_type: '',
    login_ip: '',
    login_port: NaN,
    login_user: '',
    login_mode: '',
    login_password: '',
    login_key_file: null,
    bk_networkunit_id: NaN,
    bk_biz_id: NaN,
    target_version: '',
    bk_host_id: NaN,
    re_register: false
};
const formData = reactive({
    type: '',
    business: '',
    cloud: '',
    cloud_unit: '',
    info: [cloneDeep(initData)] as AgentInstallInfo[],
});
const previewData = reactive({
    isShow: false,
});
const rules = {};
const isShow = ref(false);
const isAtBottom = ref(false);
// 安装方式
const activeInstallType = computed(() => mainStore.agentSetupType);

// 显示侧边栏安装策略
const handleShowPanel = () => {
    showRightPanel.value = true;
}
// 显示表格设置
const handleShowSetting = () => {

}

const formRef = ref(null);
const installTableRef = ref(null);
const handlePreview = async () => {
    const formValid = await formRef.value?.validate().catch(() => false);
    const res = await installTableRef.value?.tableValidate().catch(() => false);
    if (formValid && res) {
        previewData.isShow = true;
    }
}


const footerRef = ref<Element | null>(null);
const checkIfAtBottom = () => {
    if (footerRef.value) {
        let bottom = footerRef.value.getBoundingClientRect().bottom;
        if (isShow.value) {
            bottom += 224;
        } else {
            bottom -= 224;
        }
        isAtBottom.value = bottom >= window.innerHeight;
    }
}
const debouncedCheck = debounce(checkIfAtBottom, 100);
watch(() => isShow.value, (val: boolean) => {
    checkIfAtBottom();
});
onMounted(async () => {
    if (footerRef.value) {
        window.addEventListener('resize', debouncedCheck);
        checkIfAtBottom();
    }
    formData.info = nodeManageStore.agentEditParams.tableData.map((item: Host) => ({
        ...item,
        target_version: item.state.node_version
    }));
});
onUnmounted(() => {
    if (footerRef.value) {
        window.removeEventListener('resize', debouncedCheck);
    }
});
</script>