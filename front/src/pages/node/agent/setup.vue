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
                    <Button text theme="primary" @click="handleShowPanel">{{ $t('platform.nodeMan.installAgentPage.tip2SecondSlotText') }}</Button>
                    <Button text theme="primary" @click="handleShowSetting">{{ $t('platform.nodeMan.installAgentPage.tip2ThirdSlotText') }}</Button>
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
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.business')" property="business" required>
                    <Select class="content-basic" v-model="formData.business" auto-focus filterable :list="datasourceList" @select="handleSelect"></Select>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.cloud')" property="cloud" required>
                    <Select class="content-basic" v-model="formData.cloud" auto-focus filterable :list="datasourceList" @select="handleSelect"></Select>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.cloud_unit')" property="cloud_unit" required>
                    <Select class="content-basic" v-model="formData.cloud" auto-focus filterable :list="datasourceList" @select="handleSelect"></Select>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.info')" required>
                    <installTable></installTable>
                </Form.FormItem>
                <Form.FormItem>
                    <Button text theme="primary" class="advanced-ptions" @click="isShow = !isShow">
                        <span>高级选项</span>
                        <angle-double-down-line :class="{'down': isShow}"/>
                    </Button>
                </Form.FormItem>
                <Form.FormItem :label="$t('Agent 版本')" required v-if="isShow">
                    <div class="content-basic">
                        <Table :data="systemData" :border="true" width="568">
                            <TableColumn field="os" :title="$t('操作系统')" width="200"></TableColumn>
                            <TableColumn field="version" :title="$t('包版本')" width="368">
                                <template #header>
                                    <span class="title">{{ $t('包版本') }}</span>
                                    <i class="nodeman-icon nc-bulk-edit"></i>
                                </template>
                                <template #default="{ row }">
                                    <Input type="choose" v-model="row.version"  :placeholder="$t('请选择')"/>
                                </template>
                            </TableColumn>
                        </Table>
                    </div>
                </Form.FormItem>
            </Form>
        </div>
        <div :class="['footer', {'atBottom': isAtBottom}]" ref="footerRef">
            <Button theme="primary">{{ $t('去安装') }}</Button>
            <Button>{{ $t('取消') }}</Button>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue';
import { Button, Form, Select, Input } from 'bkui-vue';
import { Table, TableColumn } from '@blueking/table';
import { AngleDoubleDownLine } from 'bkui-vue/lib/icon';
import { debounce } from 'lodash';
import installTable from '@/components/install-table.vue';

const showRightPanel = ref(false);
const formData = ref({
    type: '',
    business: '',
    cloud: '',
    cloud_unit: '',
    info: [],
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
const selectedValue = ref('');
const datasourceList = ref([]);
const isAtBottom = ref(false);
const handleSelect = (value: string) => {
}
// 安装方式
const activeInstallType = ref('normal');
const updateInstallType = (type: string) => {
    activeInstallType.value = type;
}
// 显示侧边栏安装策略
const handleShowPanel = () => {
    showRightPanel.value = true;
}
// 显示表格设置
const handleShowSetting = () => {
    
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
        .form-item-content {
            display: flex;
            align-items: center;
            gap: 8px;
            .type {
                min-width: 280px;
                height: 56px;
                background: #FFFFFF;
                border: 1px solid #C4C6CC;
                border-radius: 2px;
                display: flex;
                align-items: center;
                .prefix {
                    width: 48px;
                    height: 100%;
                    border-right: 1px solid #C4C6CC;
                    background: #F5F7FA;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 21px;
                }
                .text {
                    padding: 6px 8px;
                    p:first-child {
                        font-size: 14px;
                        color: #313238;
                        height: 22px;
                        line-height: 22px
                    }
                    p:last-child {
                        font-size: 12px;
                        color: #4D4F56;
                        height: 20px;
                        line-height: 20px;
                    }
                }
            }
        }
        .content-basic {
            width: 568px;
        }
        .advanced-ptions {
            font-size: 14px;
            span {
                margin-right: 8.5px;
            }
            .down {
                transform: rotate(180deg);
            }
        }
        .title {
            position: relative;
            margin-right: 18.5px;
            &::after {
                position: absolute;
                top: -5px;
                width: 14px;
                color: #ea3636;
                text-align: center;
                content: "*";
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