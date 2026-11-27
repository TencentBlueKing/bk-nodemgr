<template>
  <section class="system-config">
    <header class="config-toolbar">
      <h2>{{ t('route.systemConfig') }}</h2>
      <Button :disabled="loading" @click="loadSystemConfig">
        {{ t('platform.systemConfig.refresh') }}
      </Button>
    </header>
    <Alert theme="info" :title="t('platform.systemConfig.snapshotNotice')" class="mb-[16px]" />
    <Loading :loading="loading">
      <div class="config-content" :aria-busy="loading">
        <Exception
          v-if="failed"
          type="500"
          scene="part"
          :title="t('platform.systemConfig.loadFailed')"
          :description="t('platform.systemConfig.loadFailedTips')"
        >
          <Button theme="primary" @click="loadSystemConfig">
            {{ t('platform.systemConfig.retry') }}
          </Button>
        </Exception>
        <template v-else-if="config">
          <section class="config-section">
            <h3>{{ t('platform.systemConfig.basicSettings') }}</h3>
            <dl class="config-fields">
              <dt>{{ t('platform.systemConfig.tenantMode') }}</dt>
              <dd>{{ tenantMode }}</dd>
              <dt>{{ t('platform.systemConfig.environment') }}</dt>
              <dd>{{ config.system?.env || '--' }}</dd>
              <dt>{{ t('platform.systemConfig.edition') }}</dt>
              <dd>{{ config.system?.edition || '--' }}</dd>
              <dt>{{ t('platform.systemConfig.iamV3') }}</dt>
              <dd>{{ formatEnabled(config.iam?.v3_enabled) }}</dd>
              <dt>{{ t('platform.systemConfig.iamV4') }}</dt>
              <dd>{{ formatEnabled(config.iam?.v4_enabled) }}</dd>
            </dl>
          </section>
          <section class="config-section">
            <h3>{{ t('platform.systemConfig.nodeEventDataID') }}</h3>
            <p class="config-hint">{{ t('platform.systemConfig.nodeEventDataIDTips') }}</p>
            <dl class="config-fields">
              <dt>{{ t('platform.systemConfig.agentBaseAlarmEventDataID') }}</dt>
              <dd>{{ config.node_event_data_id?.agent_base_alarm_event_data_id ?? '--' }}</dd>
              <dt>{{ t('platform.systemConfig.taskProcEventDataID') }}</dt>
              <dd>{{ config.node_event_data_id?.task_proc_event_data_id ?? '--' }}</dd>
            </dl>
          </section>
          <section class="config-section">
            <h3>{{ t('platform.systemConfig.defaultDirectNetworkUnit') }}</h3>
            <p class="config-hint">{{ t('platform.systemConfig.defaultDirectNetworkUnitTips') }}</p>
            <dl class="config-fields">
              <dt>{{ t('platform.systemConfig.autoCreate') }}</dt>
              <dd>{{ formatEnabled(config.default_direct_network_unit?.enabled) }}</dd>
              <dt>{{ t('platform.systemConfig.networkUnitName') }}</dt>
              <dd>{{ config.default_direct_network_unit?.name || '--' }}</dd>
            </dl>
          </section>
          <section class="config-section">
            <h3>{{ t('platform.systemConfig.gseDeployConfigs') }}</h3>
            <p class="config-hint">{{ t('platform.systemConfig.gseDeployConfigsTips') }}</p>
            <Exception
              v-if="!config.gse_deploy_configs?.length"
              type="empty"
              scene="part"
              :title="t('platform.systemConfig.noGSEDeployConfigs')"
            />
            <section
              v-for="item in config.gse_deploy_configs"
              :key="`${item.generation}-${item.os_type}`"
              class="deploy-config"
            >
              <h4>
                {{ t('platform.systemConfig.deployConfigTitle', { generation: item.generation, os: item.os_type }) }}
              </h4>
              <dl class="config-fields">
                <dt>{{ t('platform.systemConfig.baseWorkDir') }}</dt>
                <dd>{{ item.base_work_dir || '--' }}</dd>
                <dt>{{ t('platform.systemConfig.baseDeployDir') }}</dt>
                <dd>{{ item.base_deploy_dir || '--' }}</dd>
                <dt>{{ t('platform.systemConfig.extraConfigDir') }}</dt>
                <dd>{{ item.extra_config_dir || t('platform.systemConfig.noOverride') }}</dd>
                <dt>{{ t('platform.systemConfig.hostIDPath') }}</dt>
                <dd>{{ item.host_id_path || t('platform.systemConfig.noOverride') }}</dd>
                <dt>{{ t('platform.systemConfig.v2RunDir') }}</dt>
                <dd>{{ item.v2_run_dir || t('platform.systemConfig.noOverride') }}</dd>
                <dt>{{ t('platform.systemConfig.v2DataDir') }}</dt>
                <dd>{{ item.v2_data_dir || t('platform.systemConfig.noOverride') }}</dd>
                <dt>{{ t('platform.systemConfig.zoneID') }}</dt>
                <dd>{{ item.zone_id || t('platform.systemConfig.noOverride') }}</dd>
                <dt>{{ t('platform.systemConfig.cityID') }}</dt>
                <dd>{{ item.city_id || t('platform.systemConfig.noOverride') }}</dd>
                <dt>{{ t('platform.systemConfig.httpProxyMode') }}</dt>
                <dd>{{ formatProxyMode(item.http_proxy_mode) }}</dd>
                <dt>{{ t('platform.systemConfig.firewallFeatures') }}</dt>
                <dd>
                  {{ item.firewall_features?.length
                    ? item.firewall_features.join(', ')
                    : t('platform.systemConfig.firewallUnmanaged') }}
                </dd>
              </dl>
            </section>
          </section>
        </template>
      </div>
    </Loading>
  </section>
</template>

<script setup lang="ts">
import { Alert, Button, Exception, Loading } from 'bkui-vue';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import type { SystemConfig } from '@/@types/system_config';
import { SystemConfigQueryService } from '@/api/modules/system_config';

const { t } = useI18n();
const config = ref<SystemConfig>();
const loading = ref(false);
const failed = ref(false);

const tenantMode = computed(() => {
  switch (config.value?.tenant_mode) {
    case 'single':
      return t('platform.systemConfig.singleTenant');
    case 'multiple':
      return t('platform.systemConfig.multipleTenants');
    default:
      return config.value?.tenant_mode || '--';
  }
});

function formatEnabled(value: boolean | undefined) {
  if (value === undefined) return '--';
  return t(value ? 'platform.systemConfig.enabled' : 'platform.systemConfig.disabled');
}

function formatProxyMode(mode: string) {
  switch (mode) {
    case 'inherit':
      return t('platform.systemConfig.proxyInherit');
    case 'direct':
      return t('platform.systemConfig.proxyDirect');
    case 'specified':
      return t('platform.systemConfig.proxySpecified');
    default:
      return '--';
  }
}

async function loadSystemConfig() {
  if (loading.value) return;
  loading.value = true;
  failed.value = false;
  config.value = undefined;
  try {
    const data = await SystemConfigQueryService.GetSystemConfig({});
    if (!data) {
      failed.value = true;
      return;
    }
    config.value = data;
  } catch {
    // Shared request handling presents API errors and permission application dialogs.
    failed.value = true;
  } finally {
    loading.value = false;
  }
}

onMounted(loadSystemConfig);
</script>

<style scoped>
.system-config {
  height: 100%;
  overflow: auto;
  padding: 24px;
  color: #63656e;
}

.config-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.config-toolbar h2 {
  margin: 0;
  font-size: 16px;
  color: #313238;
}

.config-content {
  min-height: 240px;
}

.config-section {
  padding: 24px;
  margin-bottom: 16px;
  background: #fff;
  border-radius: 2px;
}

.config-section h3 {
  margin: 0 0 16px;
  font-size: 14px;
  color: #313238;
}

.config-hint {
  margin: 0 0 16px;
  color: #979ba5;
  font-size: 12px;
  line-height: 20px;
}

.config-fields {
  display: grid;
  grid-template-columns: minmax(180px, 260px) minmax(0, 1fr);
  gap: 16px 24px;
  margin: 0;
  font-size: 12px;
  line-height: 20px;
}

.config-fields dt {
  color: #979ba5;
}

.config-fields dd {
  margin: 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.deploy-config + .deploy-config {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #dcdee5;
}

.deploy-config h4 {
  margin: 0 0 16px;
  font-size: 13px;
  color: #313238;
}
</style>
