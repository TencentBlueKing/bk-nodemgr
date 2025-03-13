<template>
    <div class="setup">
        <div class="tips">
            <i class="nodeman-icon nc-tips"></i>
            <div class="right-text">
                <i18n-t keypath="platform.nodeMan.installAgentPage.tip1" tag="p" class="tip1">
                    <span>{{ $t('platform.nodeMan.installAgentPage.tip1FirstSlotText') }}</span>
                    <span>{{ $t('platform.nodeMan.installAgentPage.tip1SecondSlotText') }}</span>
                </i18n-t>
                <i18n-t keypath="platform.nodeMan.installAgentPage.tip2" tag="p" class="tip2">
                    <span>{{ $t('platform.nodeMan.installAgentPage.tip2FirstSlotText') }}</span>
                    <Button text theme="primary" @click="handleShowPanel">{{
                        $t('platform.nodeMan.installAgentPage.tip2SecondSlotText') }}</Button>
                    <Button text theme="primary" @click="handleShowSetting">{{
                        $t('platform.nodeMan.installAgentPage.tip2ThirdSlotText') }}</Button>
                </i18n-t>
                <p>

                </p>
            </div>
        </div>
        <div class="installForm">
            <Form ref="formRef" :model="formData" :rules="rules">
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.type')" required>
                    <install-type @update:active-type="updateInstallType"></install-type>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.info')" required>
                    <installTable>
                        <Upload :accept="'.xlsx'" :handle-res-code="handleRes" :select-change="handleSelectChange"
                            :url="'https://jsonplaceholder.typicode.com/posts/'" :files="fileList" with-credentials
                            @done="handleDone" @error="handleError" @progress="handleProgress" @success="handleSuccess"
                            :before-upload="handleBeforeUpload">
                            <template #tip>
                                <div class="uploadTip">
                                    <span>{{ $t('仅支持 .xlsx 类型文件，下载') }}</span>
                                    <a :href="url" download="bk_nodeman_info.xlsx">
                                        <Button text theme="primary">
                                            {{ $t('模版文件') }}
                                        </Button>
                                    </a>
                                </div>
                            </template>
                        </Upload>
                    </installTable>
                </Form.FormItem>
            </Form>
        </div>
        <div :class="['footer', { 'atBottom': isAtBottom }]" ref="footerRef">
            <Button theme="primary">{{ $t('去安装') }}</Button>
            <Button>{{ $t('取消') }}</Button>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue';
import { Button, Form, Select, Input, Upload, Message } from 'bkui-vue';
import { Table, TableColumn } from '@blueking/table';
import { AngleDoubleDownLine, Done } from 'bkui-vue/lib/icon';
import { debounce } from 'lodash';
import InstallTable from '@/components/install-table.vue';

const showRightPanel = ref(false);
const formData = ref({
    type: '',
    business: '',
    cloud: '',
    cloud_unit: '',
    info: [],
});
const rules = {};
const isShow = ref(false);
const isAtBottom = ref(false);
// 显示侧边栏安装策略
const handleShowPanel = () => {
    showRightPanel.value = true;
}
// 显示表格设置
const handleShowSetting = () => {

}
const activeInstallType = ref('normal');
const updateInstallType = (type: string) => {
    activeInstallType.value = type;
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
onMounted(() => {
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
<style lang="postcss" scoped>
.setup {
    padding: 24px 0 48px 0;

    .tips {
        margin: 0 24px;
        display: flex;
        height: 56px;
        background: #F0F8FF;
        border: 1px solid #C5DAFF;
        border-radius: 2px;
        padding: 6px 9px;
        gap: 9px;

        .nc-tips {
            padding-top: 2px;
            color: #3A84FF;
        }

        .right-text {
            color: #4d4f56;
            font-size: 12px;
            line-height: 20px;
            text-align: left;

            .tip1 {
                span {
                    font-weight: 700;
                    color: #313238;
                }
            }

            .tip2 {
                span {
                    font-weight: 700;
                    color: #313238;
                }
            }
        }
    }

    .installForm {
        margin: 24px;

        .content-basic {
            width: 568px;
        }

        .bk-upload {
            padding: 24px;

            .uploadTip {
                display: flex;
                align-items: center;
                gap: 3px;
            }
        }

    }

    .footer {
        height: 48px;
        width: 100%;
        display: flex;
        align-items: center;
        gap: 8px;
        padding-left: 174px;
        z-index: 10;

        .bk-button {
            width: 88px;

            &.bk-button-primary {
                width: 100px;
            }
        }

        &.atBottom {
            position: fixed;
            bottom: 0;
            background: #fff;
        }
    }
}
</style>