<template>
    <div class="form-item-content">
        <div 
            v-for="item in installTypeList"
            :key="item.type"
            :class="['type', {'active': activeType === item.type}]"
            @click="handleClick(item.type)"
        >
            <div class="prefix">
                <i :class="['nodeman-icon', item.icon]"></i>
            </div>
            <div class="text">
                <p>{{ item.name }}</p>
                <p>{{ item.desc }}</p>
            </div>
            <div class="checked" v-show="activeType === item.type">
                <i class="nodeman-icon nc-check-small"></i>
            </div>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { ref, computed, defineProps, defineEmits } from 'vue';

const props = defineProps({
    needTypeList: {
        type: Array,
        default: () => ['normal', 'excel_import', 'manual']
    }
});
const emit = defineEmits(['update:activeType']);
const installTypeConfig = [
    {
        type: 'normal',
        name: '普通远程安装',
        icon: 'nc-monitor',
        desc: '线上表单填写，需要提供登录信息'
    },
    {
        type: 'excel_import',
        name: 'Excel 导入远程安装',
        icon: 'nc-excel',
        desc: 'Excel 导入填写， 需要提供登录信息'
    },
    {
        type: 'manual',
        name: '手动安装',
        icon: 'nc-manual',
        desc: '无需提供登录信息，自行在服务器上执行给定命令完成安装'
    }
];
const installTypeList = computed(() => installTypeConfig.filter(item => props.needTypeList.includes(item.type)));
const activeType = ref('normal');
const handleClick = (type: string) => {
    activeType.value = type;
    emit('update:activeType', type);
}
</script>
<style lang="postcss" scoped>
.form-item-content {
    display: flex;
    align-items: center;
    gap: 8px;
    .active {
        &.type {
            border-color: #3A84FF;
            .prefix {
                background: #E1ECFF;
                border-right: 1px solid #3A84FF;
                color: #3A84FF;
            }
        }
    }
    .checked {
        position: absolute;
        top: 0;
        right: 0;
        width: 0;
        height: 0;
        border-style: solid;
        border-width: 0 32px 32px 0;
        border-color: transparent #3A84FF transparent transparent;
        .nc-check-small {
            position: absolute;
            font-size: 20px;
            color: #fff;
            top: 0;
            right: -32px;
        }
    }
    .type {
        position: relative;
        min-width: 280px;
        height: 56px;
        background: #FFFFFF;
        border: 1px solid #C4C6CC;
        border-radius: 2px;
        display: flex;
        align-items: center;
        cursor: pointer;

        .prefix {
            width: 48px;
            height: 100%;
            border-right: 1px solid #C4C6CC;
            background: #F5F7FA;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 21px;
            color: #979BA5;
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
</style>