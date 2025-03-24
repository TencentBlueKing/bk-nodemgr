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
        <div class="m-[24px]" v-if="activeInstallType === 'import'">
            <Form ref="formRef" :model="formData" :rules="rules">
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.type')" required>
                    <install-type></install-type>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.info')" required>
                    <install-table ref="installTableRef">
                        <Upload class="p-[24px]" :accept="'.xlsx'" :handle-res-code="handleRes" :select-change="handleSelectChange"
                            :url="'https://jsonplaceholder.typicode.com/posts/'" :files="fileList" with-credentials
                            @done="handleDone" @error="handleError" @progress="handleProgress" @success="handleSuccess"
                            :before-upload="handleBeforeUpload">
                            <template #tip>
                                <div class="flex items-center gap-[3px]">
                                    <span>{{ $t('仅支持 .xlsx 类型文件，下载') }}</span>
                                    <a :href="url" download="bk_nodeman_info.xlsx">
                                        <Button text theme="primary">
                                            {{ $t('模版文件') }}
                                        </Button>
                                    </a>
                                </div>
                            </template>
                        </Upload>
                    </install-table>
                </Form.FormItem>
            </Form>
        </div>
        <div class="m-[24px]" v-else>
            <Form ref="formRef" :model="formData" :rules="rules">
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.type')" required>
                    <install-type></install-type>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.business')" property="business" required>
                    <Select class="w-[568px]" v-model="formData.business" auto-focus filterable placeholder="选择业务"
                        @select="handleSelect">
                        <Select.Option v-for="item in businessList" :key="item.bk_biz_id" :name="item.bk_biz_name"
                            :id="item.bk_biz_id">
                            [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
                        </Select.Option>
                    </Select>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.cloud')" property="cloud" required>
                    <Select class="w-[568px]" v-model="formData.cloud" auto-focus filterable :list="networkAreaList"
                        id-key="bk_networkarea_id"
                        display-key="bk_networkarea_name"
                        @select="handleSelect"></Select>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.cloud_unit')" property="cloud_unit"
                    required>
                    <Select class="w-[568px]" v-model="formData.cloud_unit" auto-focus filterable :list="networkUnitList"
                        id-key="bk_networkunit_id"
                        display-key="bk_networkunit_name"
                        @select="handleSelect"></Select>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.info')" required>
                    <install-table ref="installTableRef" :data="tableData"></install-table>
                </Form.FormItem>
                <Form.FormItem>
                    <Button text theme="primary" class="text-[14px]" @click="isShow = !isShow">
                        <span class="mr-[8.5px]">高级选项</span>
                        <angle-double-down-line :class="{ 'transform rotate-180': isShow }" />
                    </Button>
                </Form.FormItem>
                <Form.FormItem :label="$t('Agent 版本')" required v-if="isShow">
                    <div class="w-[568px]">
                        <Table :data="systemData" :border="true" width="568">
                            <TableColumn field="os" :title="$t('操作系统')" width="200"></TableColumn>
                            <TableColumn field="version" :title="$t('包版本')" width="368">
                                <template #header>
                                    <span class="mr-[2px]">{{ $t('包版本') }}</span>
                                    <span class="mr-[10px] w-[14px] text-[#ea3636]">*</span>
                                    <i class="nodeman-icon nc-bulk-edit"></i>
                                </template>
                                <template #default="{ row }">
                                    <Input type="choose" v-model="row.version" :placeholder="$t('请选择')" />
                                </template>
                            </TableColumn>
                        </Table>
                    </div>
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

const route = useRoute();

const initData = {
  ipv4: '',
  ipv6: '',
  os: '',
  login_ip: '',
  authentication: '',
  login_port: '',
  login_user: '',
  password: '',
};
const tableData = ref([cloneDeep(initData)]);
const mainStore = useMainStore();
const showRightPanel = ref(false);
const formData = ref({
    type: '',
    business: '',
    cloud: '',
    cloud_unit: '',
    info: [],
});
const previewData = reactive({
    isShow: false,
});
const systemData = ref([
    {
        os: 'linux_x86_64',
        version: ''
    },
    {
        os: 'linux_aarch64',
        version: ''
    },
    {
        os: 'windows_x86_32',
        version: ''
    },
    {
        os: 'windows_x86_64',
        version: ''
    },
])
const rules = {};
const isShow = ref(false);
const businessList = computed(() => mainStore.businessList);
const isAtBottom = ref(false);
const handleSelect = (value: string) => {
}
// 安装方式
const activeInstallType = computed(() => mainStore.agentSetupType);
const networkAreaList = ref<NetworkArea[]>([]);
// 管控区域下拉列表获取
const getNetworkAreaList = async () => {
    const res = await TopoService.NetworkAreaList({
        page: {
            limit: 0,
        }
    }).catch((err: any) => {
        console.log(err);
        return {
            total: 0,
            items: [],
        }
    });
    console.log("🚀 ~ getNetworkAreaList ~ res:", res.items)
    networkAreaList.value = res.items;
}

// 管控单元下拉列表获取
const networkUnitList = ref<NetworkUnit[]>([]);
const getNetworkUnitList = async () => {
    const bk_networkarea_id = formData.value.cloud ? [formData.value.cloud] : [];
    const res = await TopoService.NetworkUnitList({
        page: {
            limit: 0,
        },
        exact_include_conditions: {
            bk_networkarea_id,
        }
    }).catch((err: any) => {
        console.log(err);
        return {
            total: 0,
            items: [],
        }
    });
    console.log("🚀 ~ getNetworkAreaList ~ res:", res.items)
    networkUnitList.value = res.items;
}
// 显示侧边栏安装策略
const handleShowPanel = () => {
    showRightPanel.value = true;
}
// 显示表格设置
const handleShowSetting = () => {

}

// Excel 导入
let fileList = ref<File[]>([]);
const url = `${window.location.origin}${import.meta.env.BK_SITE_URL}${import.meta.env.BK_API_PREFIX}api/excel/download`;
const handleSuccess = (file: File, fileList: File[]) => {
    console.log(file, fileList, 'handleSuccess');
};
const handleProgress = (event: Event, file: File, fileList: File[]) => {
    console.log(event, file, fileList, 'handleProgress');
};
const handleError = (file: File, fileList: File[], error: { message: string }) => {
    Message({
        theme: 'error',
        message: error.message
    })
};
const handleDone = (curFileList: File[]) => {
    console.log(fileList, 'handleDone');
    fileList.value = [...curFileList]
};
const handleRes = (response: { id: number | string }) => {
    console.log(response, 'handleRes');
    if (response.id) {
        return true;
    }
    return false;
};
const handleBeforeUpload = (file: File, fileList: File[]) => {
    const whiteList = ['xlsx'];
    let AllFiles = fileList.filter((v) => whiteList.includes(v.name.substring(file.name.lastIndexOf('.') + 1)));
    if (AllFiles.length !== fileList.length) {
        fileList.pop();
        Message({
            theme: 'warning',
            message: '仅支持 .xlsx 类型文件'
        });
        return false;
    }
    return true;
};
const handleSelectChange = (event: Event) => {
    console.log(event, 'change');
};

const installTableRef = ref(null);
const handlePreview = async () => {
    if (installTableRef.value) {
        const res = await installTableRef.value?.validate().catch(() => false);
        if (res) {
            previewData.isShow = true;
        }
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
watch(() => activeInstallType.value, (val: string) => {
    installTableRef.value?.clearAllValidate();
});
onMounted(async () => {
    await getNetworkAreaList();
    await getNetworkUnitList();
    if (footerRef.value) {
        window.addEventListener('resize', debouncedCheck);
        checkIfAtBottom();
    }
});
onUnmounted(() => {
    if (footerRef.value) {
        window.removeEventListener('resize', debouncedCheck);
    }
});
</script>