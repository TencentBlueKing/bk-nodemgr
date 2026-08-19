<template>
  <div class="step-param-config bg-[#fff] p-[24px] rounded-[2px]">
    <div class="mb-[16px] flex items-center justify-between" style="max-width: 780px;">
      <div>
        <h3 class="text-[16px] font-bold text-[#313238] mb-[4px]">{{ $t('pluginOperation.paramConfig.title') }}</h3>
        <p class="text-[12px] text-[#979BA5]">{{ $t('pluginOperation.paramConfig.description') }}</p>
      </div>
      <!-- 批量复制粘贴（图标 + tooltip，hover 变蓝） -->
      <div v-if="platformVersions.length" class="flex items-center">
        <i
          v-bk-tooltips="{ content: $t('pluginOperation.paramConfig.copyAllConfig') }"
          class="nodeman-icon nc-copy copy-icon"
          @click="copyAllConfig"
        ></i>
        <i
          v-bk-tooltips="{ content: $t('pluginOperation.paramConfig.pasteAllConfig') }"
          class="nodeman-icon nc-paste-line copy-icon"
          @click="pasteAllConfig"
        ></i>
      </div>
    </div>

    <!-- 无数据提示 -->
    <div v-if="!platformVersions.length" class="text-[14px] text-[#979BA5] py-[20px] text-center">
      {{ $t('pluginOperation.paramConfig.noPlatform') }}
    </div>

    <!-- 全部加载中 -->
    <div v-else-if="loadingAll" class="flex items-center justify-center" :style="{ minHeight: '200px' }">
      <Loading />
    </div>

    <!-- 遍历每个平台的参数配置表单 -->
    <template v-else>
      <div
        v-for="(item, index) in formItems"
        :key="item.platform"
        class="param-card"
        :class="{ 'is-error': item.error }"
      >
        <!-- 卡片头部：可折叠 -->
        <div class="param-card-header">
          <!-- 左侧：展开/收起 + 平台名 + 版本 -->
          <div class="param-card-header-left" @click="toggleCard(index)">
            <DownShape
              class="arrow-icon"
              :class="{ 'is-collapsed': !item.expanded }"
            />
            <span class="platform-name">{{ item.platform.replace('_', '/') }}</span>
            <span class="platform-version">({{ item.version }})</span>
          </div>

          <!-- 右侧：复制 + 粘贴 + 复用到相同表单 + 恢复默认值（仅在有表单数据时显示） -->
          <div v-if="item.loaded && !item.empty && !item.error" class="param-card-header-right">
            <i
              v-bk-tooltips="{ content: $t('pluginOperation.paramConfig.copyConfig') }"
              class="nodeman-icon nc-copy copy-icon"
              @click.stop="copyConfig(index)"
            ></i>
            <i
              v-bk-tooltips="{ content: $t('pluginOperation.paramConfig.pasteConfig') }"
              class="nodeman-icon nc-paste-line copy-icon"
              @click.stop="pasteConfig(index)"
            ></i>
            <i
              v-if="formItems.length > 1"
              v-bk-tooltips="{ content: $t('pluginOperation.paramConfig.applyAll') }"
              class="nodeman-icon nc-brush-fill copy-icon"
              @click.stop="applyToAllSameForms(index)"
            ></i>
            <i
              v-bk-tooltips="{ content: $t('pluginOperation.paramConfig.resetDefault') }"
              class="nodeman-icon nc-withdraw-fill reset-icon"
              @click.stop="resetFormValues(index)"
            ></i>
          </div>
        </div>

        <!-- 卡片内容：表单区域（展开时显示） -->
        <div v-show="item.expanded && !item.loading" class="param-card-body">
          <!-- 加载失败 / 无变量时提示 -->
          <div v-if="item.empty || item.error" class="py-[16px] px-[24px] text-[14px]" :class="item.error ? 'text-[#EA3636]' : 'text-[#979BA5]'">
            {{ item.error || t('pluginOperation.paramConfig.noParams') }}
          </div>

          <!-- 有变量的表单 -->
          <BkSchemaForm
            v-else
            :ref="(el: any) => setFormRef(index, el)"
            v-model="formValues[item.platform]"
            :schema="item.schema"
            form-type="vertical"
            :label-width="300"
          />
        </div>

        <!-- 折叠状态下的加载动画 -->
        <div v-if="item.loading && !item.expanded" class="param-card-loading">
          <Loading size="small" />
        </div>
      </div>
    </template>
  </div>
</template>

<script lang="ts">
// 模块级缓存，确保 createForm() 只执行一次
import type { Component } from 'vue';

import createForm from '@blueking/bkui-form';
import '@blueking/bkui-form/dist/style.css';

let _cachedForm: Component | null = null;
function getOrCreateForm(): Component {
  if (!_cachedForm) {
    _cachedForm = createForm();
  }
  return _cachedForm;
}
</script>

<script lang="ts" setup>
import { DownShape } from 'bkui-vue/lib/icon';
import { Loading, Message } from 'bkui-vue';
import { reactive, ref, triggerRef } from 'vue';
import { useI18n } from 'vue-i18n';

import { PackageService } from '@/api/modules/pkg';
import { PACKAGE_GENERATION } from '@/common/const';

import { usePluginOperationStore } from '@/stores/plugin-operation';

const pluginOpStore = usePluginOperationStore();

const { t, locale } = useI18n();

// 在 setup 上下文中初始化（首次），后续实例复用缓存
const BkSchemaForm = getOrCreateForm();

const props = defineProps({
  /** 插件名称 */
  pluginName: {
    type: String,
    default: '',
  },
  /** 平台版本列表 — 来自步骤1用户选择 [{ os: 'linux_x86_64', version: '10.8.51' }] */
  platformVersions: {
    type: Array as () => { os: string; version: string }[],
    default: () => [],
  },
});

const formValues = defineModel<Record<string, Record<string, any>>>('formValues', { required: true });

/** 单个平台的表单项 */
interface FormItem {
  platform: string;          // 显示用：'linux_x86_64'
  osType: string;            // 接口参数：'linux'
  cpuArch: string;           // 接口参数：'x86_64'
  version: string;
  expanded: boolean;
  loading: boolean;
  loaded: boolean;
  empty: boolean;
  error: string;
  schema: any;
}

/** 所有平台的表单项列表 */
const formItems = reactive<FormItem[]>([]);
/** 是否全部在初始加载 */
const loadingAll = ref(false);
/** BkSchemaForm 实例引用（按索引存储） */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const formRefMap = ref<Map<number, any>>(new Map());
/** 每个平台的表单初始值（用于恢复默认） */
const initialFormValues = ref<Record<string, Record<string, any>>>({});
/** 每个平台的配置模板名称列表（platform → config_template_name[]） */
const configNamesMap = ref<Record<string, string[]>>({});

/**
 * 将后端返回的 ConfigVariables Property 递归转换为 bkui-form JSON Schema 格式
 * 支持任意深度的嵌套对象（object → properties → object → properties → ...）
 */
function convertToSchema(variables: Record<string, any>): any {
  if (!variables || Object.keys(variables).length === 0) return null;

  const properties: Record<string, any> = {};
  const required: string[] = [];

  function convertProp(p: any, key: string): any {
    const isZh = locale.value.startsWith('zh');
    const field: Record<string, any> = {
      title: p.title || key,
      type: (p.type || 'string').toLowerCase(),
      description: isZh ? (p.description || '') : (p.description_en || p.description || ''),
    };

    // protobuf Value 类型解包
    if (p.default !== undefined && p.default !== null) {
      field.default = typeof p.default === 'object'
        ? (p.default.string_value ?? p.default.number_value ?? p.default.bool_value)
        : p.default;
    }

    if (field.type === 'number' || field.type === 'integer') {
      field.type = 'number';
    }
    // 后端用 'bool'，JSON Schema 标准要求是 'boolean'
    if (field.type === 'bool') {
      field.type = 'boolean';
    }

    // 处理数组类型：后端用 properties 描述元素结构，bkui-form 需要 items
    if (field.type === 'array' && p.properties && Object.keys(p.properties).length > 0) {
      // 后端 array 的 properties 是单层描述每个元素的字段
      // 转为 bkui-form 的 items 格式
      const itemEntries = Object.entries(p.properties);
      if (itemEntries.length === 1) {
        // 单字段数组（如 string[]）→ items 直接取该字段的 schema
        const [itemKey, itemVal] = itemEntries[0];
        field.items = convertProp(itemVal, itemKey);
      } else {
        // 多字段数组 → items 为 object，包含所有子字段
        const itemProps: Record<string, any> = {};
        const itemRequired: string[] = [];
        for (const [itemKey, itemVal] of itemEntries) {
          const iv = itemVal as any;
          if (iv.required) itemRequired.push(itemKey);
          itemProps[itemKey] = convertProp(iv, itemKey);
        }
        field.items = {
          type: 'object',
          ...(itemRequired.length > 0 ? { required: itemRequired } : {}),
          properties: itemProps,
        };
      }

      // 数组默认值初始化为空数组
      if (!field.default) field.default = [];
    }

    // 递归处理嵌套对象
    if (field.type === 'object' && p.properties && Object.keys(p.properties).length > 0) {
      const childProps: Record<string, any> = {};
      const childRequired: string[] = [];
      for (const [childKey, childVal] of Object.entries(p.properties)) {
        const cp = childVal as any;
        if (cp.required) childRequired.push(childKey);
        childProps[childKey] = convertProp(cp, childKey);
      }
      field.properties = childProps;
      if (childRequired.length > 0) field.required = childRequired;
      // 让嵌套 object 以带标题和边框的分组形式展示，而不是被打平
      field['ui:group'] = {
        showTitle: true,
        border: true,
        type: 'card',
      };
    }

    return field;
  }

  for (const [key, prop] of Object.entries(variables)) {
    const p = prop as any;
    if (p.required) required.push(key);
    properties[key] = convertProp(p, key);
  }

  return {
    type: 'object',
    required: required.length > 0 ? required : undefined,
    properties,
  };
}

/** 从 schema 中递归提取默认值 */
function extractDefaults(schema: any): Record<string, any> {
  if (!schema?.properties) return {};
  const defaults: Record<string, any> = {};
  for (const [key, prop] of Object.entries(schema.properties)) {
    const p = prop as any;
    if (p.default !== undefined) {
      defaults[key] = p.default;
    } else if (p.type === 'boolean') {
      defaults[key] = false;
    } else if (p.type === 'number' || p.type === 'integer') {
      defaults[key] = 0;
    } else if (p.type === 'array') {
      defaults[key] = [];
    } else if (p.type === 'object' && p.properties) {
      defaults[key] = extractDefaults(p);
    } else {
      defaults[key] = '';
    }
  }
  return defaults;
}

/** 为指定 index 设置 BkSchemaForm 引用 */
function setFormRef(index: number, el: any) {
  if (el) {
    formRefMap.value.set(index, el);
  }
}

/** 切换卡片展开/收起 */
function toggleCard(index: number) {
  formItems[index].expanded = !formItems[index].expanded;
}

/** 恢复指定平台的表单值为初始值 */
function resetFormValues(index: number) {
  const item = formItems[index];
  if (!item || !initialFormValues.value[item.platform]) return;
  formValues.value[item.platform] = { ...initialFormValues.value[item.platform] };
}

/** 将 'linux_x86_64' 解析为 { osType, cpuArch } */
function parsePlatform(platform: string): { osType: string; cpuArch: string } {
  const idx = platform.indexOf('_');
  if (idx === -1) return { osType: platform, cpuArch: '' };
  return { osType: platform.substring(0, idx), cpuArch: platform.substring(idx + 1) };
}

/** 将前端平台标识 'linux_x86_64' 转为后端返回的 key 格式 'linux/x86_64' */
function platformToKey(platform: string): string {
  const { osType, cpuArch } = parsePlatform(platform);
  return cpuArch ? `${osType}/${cpuArch}` : osType;
}

/** 处理单个平台的配置变量数据，生成 schema 并初始化表单值 */
function applyConfigToItem(item: FormItem, configVars: any[]): void {
  // 只保留 is_main_config 为 true 的配置模板
  const mainConfigs = configVars.filter((cv: any) => cv.is_main_config);

  // 收集 main config 的 name（用于提交时传 config_template_name）
  const names = mainConfigs.map((cv: any) => cv.name).filter(Boolean);
  if (names.length > 0) {
    configNamesMap.value[item.platform] = names;
  }

  if (mainConfigs.length > 0) {
    const mainConfig = mainConfigs[0];
    const schema = convertToSchema(mainConfig.variables || {});
    if (schema) {
      item.schema = schema;
      item.empty = false;
      item.expanded = true;
    } else {
      item.empty = true;
    }
  } else {
    item.empty = true;
  }

  if (!formValues.value[item.platform]) {
    const defaults = extractDefaults(item.schema);
    formValues.value[item.platform] = { ...defaults };
    initialFormValues.value[item.platform] = { ...defaults };
  }
}

/** 一次性加载所有平台的配置变量 schema（新版接口支持 platforms 数组批量查询） */
async function loadAllSchemas() {
  // 清空旧数据
  formItems.length = 0;

  if (!props.pluginName || props.platformVersions.length === 0) {
    return;
  }

  loadingAll.value = true;

  // 构建表单项列表（默认收起，有参数的会在加载完成后自动展开）
  for (let i = 0; i < props.platformVersions.length; i++) {
    const pv = props.platformVersions[i];
    const { osType, cpuArch } = parsePlatform(pv.os);
    formItems.push({
      platform: pv.os,
      osType,
      cpuArch,
      version: pv.version,
      expanded: false,
      loading: true,
      loaded: false,
      empty: false,
      error: '',
      schema: null,
    });
  }

  try {
    // 一次调用传入所有 platforms，获取所有平台的配置变量
    const platforms = formItems.map(item => ({
      os_type: item.osType,
      cpu_arch: item.cpuArch,
    }));
    const res = await PackageService.GetConfigVariablesReleasePlugin({
      generation: PACKAGE_GENERATION,
      name: props.pluginName,
      platforms,
      version: formItems[0].version,
    });

    // 从返回的 Record<string, ConfigVariablesList> 中按平台 key 提取
    const configVariablesMap = res?.config_variables || {};
    for (const item of formItems) {
      const key = platformToKey(item.platform);
      const configList = configVariablesMap[key];
      const configVars = configList?.items || [];
      applyConfigToItem(item, configVars);
      item.loaded = true;
      item.loading = false;
    }
  } catch (e: any) {
    console.error('Failed to load config variables:', e);
    for (const item of formItems) {
      item.error = t('pluginOperation.paramConfig.loadFailed', { msg: e?.message || '' });
      item.empty = true;
      item.loaded = true;
      item.loading = false;
    }
  }

  loadingAll.value = false;
}

const validate = async (): Promise<boolean> => {
  let allValid = true;
  for (const [index, formEl] of formRefMap.value.entries()) {
    if (formEl?.validateForm) {
      try {
        const valid = await formEl.validateForm();
        if (!valid) allValid = false;
      } catch {
        allValid = false;
      }
    }
  }
  return allValid;
};

defineExpose({ validate, load: loadAllSchemas, configNamesMap });

/** 复制单个平台的参数到剪贴板 + Pinia */
async function copyConfig(index: number) {
  const item = formItems[index];
  if (!item) return;
  const data = formValues.value[item.platform];
  const json = JSON.stringify({ [item.platform]: data }, null, 2);
  // 双写：剪贴板 + Pinia（粘贴时优先读 Pinia）
  try {
    await navigator.clipboard.writeText(json);
  } catch {
    const textarea = document.createElement('textarea');
    textarea.value = json;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
  }
  pluginOpStore.setCopiedConfig({ [item.platform]: data });
  Message({ theme: 'success', message: t('pluginOperation.paramConfig.copySuccess') });
}

/** 复制全部平台的参数到剪贴板 + Pinia */
async function copyAllConfig() {
  const data: Record<string, any> = {};
  for (const item of formItems) {
    if (formValues.value[item.platform]) {
      data[item.platform] = formValues.value[item.platform];
    }
  }
  const json = JSON.stringify(data, null, 2);
  try {
    await navigator.clipboard.writeText(json);
  } catch {
    const textarea = document.createElement('textarea');
    textarea.value = json;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
  }
  pluginOpStore.setCopiedConfig(data);
  Message({ theme: 'success', message: t('pluginOperation.paramConfig.copySuccess') });
}

/** 从剪贴板读取文本（不可用时返回空） */
async function readClipboardText(): Promise<string> {
  if (!navigator.clipboard?.readText) return '';
  try {
    const text = await navigator.clipboard.readText();
    return text || '';
  } catch {
    return '';
  }
}

/** 从外部剪贴板解析 JSON */
function parseClipboardJson(text: string): Record<string, any> | null {
  if (!text?.trim()) return null;
  try {
    const parsed = JSON.parse(text);
    if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
      return parsed;
    }
  } catch { /* ignore */ }
  return null;
}

/** 粘贴全部参数：剪贴板优先 → Pinia 回退 */
async function pasteAllConfig() {
  // 优先剪贴板
  const clipboardText = await readClipboardText();
  const clipboardData = parseClipboardJson(clipboardText);
  if (clipboardData && Object.keys(clipboardData).length > 0) {
    applyPastedConfig(clipboardData);
    return;
  }

  // 回退 Pinia
  const cached = pluginOpStore.getCopiedConfig();
  if (cached && Object.keys(cached).length > 0) {
    applyPastedConfig(cached);
    return;
  }
  Message({ theme: 'warning', message: t('pluginOperation.paramConfig.pasteEmpty') });
}

/** 粘贴到指定平台：剪贴板优先 → Pinia 回退，按 platform key 匹配 */
async function pasteConfig(index: number) {
  const item = formItems[index];
  if (!item) return;

  // 优先剪贴板
  const clipboardText = await readClipboardText();
  const clipboardData = parseClipboardJson(clipboardText);
  const source = (clipboardData?.[item.platform] ? clipboardData : null)
    ?? pluginOpStore.getCopiedConfig();

  if (!source || Object.keys(source).length === 0) {
    Message({ theme: 'warning', message: t('pluginOperation.paramConfig.pasteEmpty') });
    return;
  }
  const value = source[item.platform];
  if (value && typeof value === 'object') {
    formValues.value[item.platform] = JSON.parse(JSON.stringify(value));
    triggerRef(formValues);
    Message({ theme: 'success', message: t('pluginOperation.paramConfig.pasteSuccess') });
  } else {
    Message({ theme: 'warning', message: t('pluginOperation.paramConfig.pasteFormatError') });
  }
}

/** 将当前平台参数复用给所有 schema 结构相同的其他平台 */
function applyToAllSameForms(index: number) {
  const source = formItems[index];
  if (!source) return;
  const sourceValue = formValues.value[source.platform];
  const sourceKeys = Object.keys(source.schema?.properties || {});

  let reapplied = 0;
  for (let i = 0; i < formItems.length; i++) {
    if (i === index) continue;
    const target = formItems[i];
    if (!target.loaded || target.error) continue;
    const targetKeys = Object.keys(target.schema?.properties || {});
    // schema 结构相同才复用
    if (sourceKeys.length !== targetKeys.length || !sourceKeys.every(k => targetKeys.includes(k))) continue;
    formValues.value[target.platform] = JSON.parse(JSON.stringify(sourceValue));
    reapplied++;
  }

  if (reapplied > 0) {
    triggerRef(formValues);
    Message({ theme: 'success', message: `${t('pluginOperation.paramConfig.applyAll')} (${reapplied})` });
  } else {
    Message({ theme: 'warning', message: t('pluginOperation.paramConfig.pasteFormatError') });
  }
}

/** 写入粘贴的配置到匹配的表单（接受已解析的对象，来自 Pinia store） */
function applyPastedConfig(data: Record<string, any>) {
  let applied = 0;
  for (const [platform, value] of Object.entries(data)) {
    if (formValues.value[platform] !== undefined && value && typeof value === 'object') {
      // 深拷贝避免共享引用导致 BkSchemaForm 不更新
      formValues.value[platform] = JSON.parse(JSON.stringify(value));
      applied++;
    }
  }
  if (applied > 0) {
    triggerRef(formValues);
    Message({ theme: 'success', message: t('pluginOperation.paramConfig.pasteSuccess') });
  } else {
    Message({ theme: 'warning', message: t('pluginOperation.paramConfig.pasteFormatError') });
  }
}
</script>

<style lang="postcss" scoped>
.param-card {
  border: 1px solid #DCDEE5;
  border-radius: 2px;
  margin-bottom: 12px;
  overflow: hidden;
  max-width: 780px;

  &.is-error {
    border-color: #EA3636;
  }
}

.param-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #F5F7FA;
  user-select: none;

  &:hover {
    background: #EBECF0;
  }
}

.param-card-header-left {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.arrow-icon {
  width: 16px;
  height: 16px;
  color: #979BA5;
  margin-right: 6px;
  transition: transform 0.2s ease;
  flex-shrink: 0;

  &.is-collapsed {
    transform: rotate(-90deg);
  }
}

.platform-name {
  font-size: 14px;
  color: #313238;
  line-height: 22px;
}

.platform-version {
  font-size: 14px;
  color: #979BA5;
  margin-left: 4px;
  line-height: 22px;
}

.param-card-header-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.reset-icon {
  font-size: 16px;
  color: #979BA5;
  cursor: pointer;
  padding: 2px;

  &:hover {
    color: #3A84FF;
  }
}

.copy-icon {
  font-size: 16px;
  color: #979BA5;
  cursor: pointer;
  padding: 2px;
  margin-right: 8px;

  &:hover {
    color: #3A84FF;
  }
}

.copy-icon {
  font-size: 16px;
  color: #979BA5;
  cursor: pointer;
  padding: 2px;
  margin-right: 8px;

  &:hover {
    color: #3A84FF;
  }
}

.paste-icon {
  font-size: 16px;
  color: #979BA5;
  cursor: pointer;
  padding: 2px;
  margin-right: 8px;

  &:hover {
    color: #3A84FF;
  }
}

.param-card-body {
  padding: 16px 24px;
  background: #fff;
  min-height: 40px;
}

.param-card-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px 16px;
}
</style>
