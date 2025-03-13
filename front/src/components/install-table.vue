<template>
    <VxeTable :data="tableData" border round>
      <!-- 主机 IP -->
      <VxeColgroup title="主机 IP" align="center">
        <VxeColumn field="ipv4" title="内网 IPv4">
          <template #default="{ row }">
            <Input />
          </template>
        </VxeColumn>
        <VxeColumn field="ipv6" title="内网 IPv6">
          <template #default="{ row }">
            <Input />
          </template>
        </VxeColumn>
      </VxeColgroup>
  
      <!-- 主机属性 -->
      <VxeColgroup title="主机属性" align="center">
        <VxeColumn field="os" title="操作系统">
          <template #default="{ row }">
            <Input />
          </template>
        </VxeColumn>
      </VxeColgroup>
  
      <!-- 登录信息 -->
      <VxeColgroup title="登录信息" align="center">
        <VxeColumn field="loginIp" title="登录 IP">
          <template #default="{ row }">
            <Input />
          </template>
        </VxeColumn>
        <VxeColumn field="loginPort" title="登录端口">
          <template #default="{ row }">
            <Input />
          </template>
        </VxeColumn>
        <VxeColumn field="authMethod" title="认证方式">
          <template #default="{ row }">
            <Input />
          </template>
        </VxeColumn>
        <VxeColumn field="password" title="密码/密钥">
          <template #default="{ row }">
            <Input />
          </template>
        </VxeColumn>
      </VxeColgroup>
  
      <!-- 自定义设置 -->
      <VxeColgroup title="自定义配置" align="center">
        <VxeColumn field="configName" title="配置名称">
          <template #default="{ row }">
            <Input />
          </template>
        </VxeColumn>
      </VxeColgroup>

      <VxeColgroup>
        <template #header>
            <Button text><i class="nodeman-icon nc-setting"></i></Button>
            <Button text style="margin-left: 8px;"><i class="nodeman-icon nc-icon-full-screen"></i></Button>
          </template>
        <VxeColumn field="action" title="操作">
          <template #default="{ row }">
            <Button text><i class="nodeman-icon nc-plus"></i></Button>
            <Button text style="margin-left: 8px;"><i class="nodeman-icon nc-minus"></i></Button>
          </template>
        </VxeColumn>
      </VxeColgroup>
      <template #empty>
        <slot></slot>
      </template>
    </VxeTable>
</template>

<script lang="ts" setup>
import { VxeTable, VxeColumn, VxeColgroup } from '@blueking/vxe-table';
import { Input, Button } from 'bkui-vue';
import { ref } from 'vue';

const props = defineProps({
  data: {
    type: Array,
    default: () => [],
  }
});
const tableData = ref( [
  // { ipv4: '', ipv6: '', os: 'Windows', loginIp: '', loginPort: '', loginUser: '', authMethod: '', password: '', configName: '', action: '' },
]);
</script>
  