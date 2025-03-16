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
                    <div class="form-item-content">   
                        <div class="normal type">
                            <div class="prefix">
                                <i class="nodeman-icon nc-monitor"></i>
                            </div>
                            <div class="text">
                                <p>普通远程安装</p>
                                <p>线上表单填写，需要提供登录信息</p>
                            </div>
                        </div>
                        <div class="excel_import type">
                            <div class="prefix">
                                <i class="nodeman-icon nc-excel"></i>
                            </div>
                            <div class="text">
                                <p>Excel 导入远程安装</p>
                                <p>Excel 导入填写， 需要提供登录信息</p>
                            </div>
                        </div>
                        <div class="manual type">
                            <div class="prefix">
                                <i class="nodeman-icon nc-manual"></i>
                            </div>
                            <div class="text">
                                <p>手动安装</p>
                                <p>无需提供登录信息，自行在服务器上执行给定命令完成安装</p>
                            </div>
                        </div>
                    </div>
                </Form.FormItem>
                <Form.FormItem :label="$t('platform.nodeMan.installAgentPage.info')" required>
                    <installTable></installTable>
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