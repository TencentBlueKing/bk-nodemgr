<template>
    <Cascader v-model="area" :list="list" :scroll-height="136" trigger="click" @change="handleChange" @toggle="handleToggle">
        <template #trigger>
            <Button>
                <span>复制</span>
                <i class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"></i>
            </Button>
        </template>
    </Cascader>
</template>

<script lang="ts" setup>
import { Button, Cascader, Message } from 'bkui-vue';
import { ref, watch, onMounted, onBeforeUnmount } from 'vue';
import { useClipboard } from '@vueuse/core'

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
    }
});
const list = ref([
    {
        id: 'select',
        name: '勾选IP',
        disabled: true,
        children: [
            {
                id: 'ipv4',
                name: 'IPv4',
            },
            {
                id: 'ipv6',
                name: 'IPv6',
            }, ,
            {
                id: 'area+ipv4',
                name: '管控区域+IPv4',
            }, ,
            {
                id: 'area+ipv6',
                name: '管控区域+IPv6',
            },
        ],
    },
    {
        id: 'all',
        name: '所有IP',
        children: [
            {
                id: 'ipv4',
                name: 'IPv4',
            },
            {
                id: 'ipv6',
                name: 'IPv6',
            }, ,
            {
                id: 'area+ipv4',
                name: '管控区域+IPv4',
            }, ,
            {
                id: 'area+ipv6',
                name: '管控区域+IPv6',
            },
        ],
    },
]);
const area = ref([]);
// 使用 useClipboard 处理剪贴板操作
const { copy, isSupported } = useClipboard({legacy: true})
// 更换选择项
const handleChange = async () => {
    if (area.value.length === 0) return;
    const type = area.value[1];
    const list = area.value[0] === 'all' ? props.data : props.data.filter((item: any) => item.checked);
    if (list.every((item: any) => ['ipv4', 'area+ipv4'].includes(type) ? !item.inner_ip : !item.inner_ipv6)) {
        Message({
            theme: 'primary',
            message: '没有可复制的内容',
        });
        return;
    }
    const copyContent = list.map((item: any) => {
        switch (type) {
            case 'ipv4':
                return item.inner_ip;
            case 'ipv6':
                return item.inner_ipv6;
            case 'area+ipv4':
                return `${item.bk_cloud_id}:${item.inner_ip}`;
            case 'area+ipv6':
                return `${item.bk_cloud_id}:${item.inner_ipv6}`;
        }
    });
    if (isSupported) {
        try {
            await copy(copyContent.join('\n'));
            Message({
                theme: 'success',
                message: '复制成功',
            });
        } catch (error) {
            Message({
                theme: 'error',
                message: '复制失败',
            });
        }
    } else {
        Message({
            theme: 'error',
            message: '当前环境不支持剪贴板操作',
        });
    }
}
const handleToggle = (value: boolean) => {
  if (!value) area.value = []; // 关闭弹出面板时 清空选项
};
watch(() => props.disabled, (val: boolean) => {
    list.value[0].disabled = val;
},{immediate: true});
</script>