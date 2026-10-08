import { beforeEach, describe, expect, it, vi } from 'vitest';

import { createPinia, setActivePinia } from 'pinia';
import { defineComponent, h, nextTick } from 'vue';
import { createI18n } from 'vue-i18n';
import { flushPromises, mount } from '@vue/test-utils';
import { createMemoryHistory, createRouter } from 'vue-router';

import Log from '../src/pages/node/history/log.vue';
import { useMainStore } from '../src/stores/main';
import { useNodeManageStore } from '../src/stores/node-manage';

vi.mock('@/api/modules/node_workflow', () => ({
  NodeWorkflowService: {
    NodeWorkflowOperationRetry: vi.fn(),
    NodeWorkflowOperationTerminate: vi.fn(),
    NodeWorkflowOperationList: vi.fn().mockResolvedValue({
      operations: [{
        operation_id: 'op-1',
        node_deployment_info: {
          bk_host_id: 1,
          bk_host_inner_list: ['1.1.1.1'],
          bk_host_innerip_list: ['1.1.1.1'],
          bk_host_innerip_v6_list: [],
        },
        latest_oper_inst_brief_data: {
          life_cycle: {
            state: 'running',
          },
          latest_action_inst_brief_data: {
            tags: [],
          },
        },
      }],
      total_count: 1,
    }),
    NodeWorkflowOperationInstanceList: vi.fn().mockResolvedValue({
      oper_inst_data: [{
        oper_inst_id: 'oper-inst-1',
        action_names: ['install_pre_ordered_plugins'],
      }],
    }),
    NodeWorkflowOperationInstanceLogGet: vi.fn().mockResolvedValue({
      total: 1,
      oper_inst_logs: {
        install_pre_ordered_plugins: {
          life_cycle: {
            state: 'running',
            start_time: Date.now() - 1000,
            end_time: 0,
          },
          message: {
            logs: [],
          },
          display_name_zh: '预装插件',
          display_name_en: 'Install pre ordered plugins',
          sub_workflow_refs: [{
            workflow_id: 'wf:test-sub',
            workflow_domain: 'plugin',
          }],
        },
      },
      extra_execution_logs: {
        logs: [],
      },
    }),
    NodeWorkflowOperationDistinct: vi.fn().mockResolvedValue({
      state: ['running'],
    }),
  },
}));

vi.mock('@/api/modules/plugin_workflow', () => ({
  PluginWorkflowService: {
    PluginWorkflowOperationRetry: vi.fn(),
    PluginWorkflowOperationTerminate: vi.fn(),
    PluginWorkflowOperationList: vi.fn(),
    PluginWorkflowOperationInstanceList: vi.fn(),
    PluginWorkflowOperationInstanceLogGet: vi.fn(),
    PluginWorkflowOperationDistinct: vi.fn(),
  },
}));

vi.mock('@/composables/use-interval', () => ({
  default: () => ({
    start: vi.fn(),
    stop: vi.fn(),
  }),
}));

vi.mock('bkui-vue', () => {
  const Button = defineComponent({
    name: 'Button',
    emits: ['click'],
    setup(_, { attrs, emit, slots }) {
      return () => h('button', {
        ...attrs,
        onClick: (event: MouseEvent) => emit('click', event),
      }, slots.default?.());
    },
  });

  const Checkbox = defineComponent({
    name: 'Checkbox',
    props: {
      modelValue: { type: Boolean, default: false },
    },
    emits: ['change', 'update:modelValue'],
    setup(_, { attrs, slots }) {
      return () => h('label', {
        ...attrs,
        'data-test': 'checkbox',
      }, slots.default?.());
    },
  });

  const Dropdown = defineComponent({
    name: 'Dropdown',
    setup(_, { slots }) {
      return () => h('div', { 'data-test': 'dropdown' }, [
        slots.default?.(),
        slots.content?.(),
      ]);
    },
  });

  const DropdownMenu = defineComponent({
    name: 'DropdownMenu',
    setup(_, { slots }) {
      return () => h('div', slots.default?.());
    },
  });

  const DropdownItem = defineComponent({
    name: 'DropdownItem',
    emits: ['click'],
    setup(_, { attrs, emit, slots }) {
      return () => h('div', {
        ...attrs,
        onClick: (event: MouseEvent) => emit('click', event),
      }, slots.default?.());
    },
  });

  return {
    Button,
    Checkbox,
    Dropdown: Object.assign(Dropdown, { DropdownMenu, DropdownItem }),
    InfoBox: vi.fn(),
    Input: defineComponent({
      name: 'Input',
      setup(_, { attrs }) {
        return () => h('input', attrs);
      },
    }),
    Message: vi.fn(),
    overflowTitle: defineComponent({
      name: 'overflowTitle',
      setup(_, { slots }) {
        return () => h('div', slots.default?.());
      },
    }),
    SearchSelect: defineComponent({
      name: 'SearchSelect',
      setup(_, { slots }) {
        return () => h('div', slots.default?.());
      },
    }),
  };
});

vi.mock('bkui-vue/lib/icon', () => {
  const icon = (name: string) => defineComponent({
    name,
    setup() {
      return () => h('i', { 'data-icon': name });
    },
  });

  return {
    AngleUpFill: icon('AngleUpFill'),
    ArrowsLeft: icon('ArrowsLeft'),
    Close: icon('Close'),
    ExclamationCircleShape: icon('ExclamationCircleShape'),
    RightTurnLine: icon('RightTurnLine'),
    RightShape: icon('RightShape'),
    Spinner: icon('Spinner'),
  };
});

vi.mock('@blueking/table', () => {
  const TableColumn = defineComponent({
    name: 'TableColumn',
    props: {
      field: { type: String, default: '' },
      title: { type: String, default: '' },
    },
    setup() {
      return () => null;
    },
  });

  const Table = defineComponent({
    name: 'Table',
    props: {
      data: { type: Array, default: () => [] },
    },
    setup(props, { slots }) {
      return () => {
        const columns = (slots.default?.() ?? []).filter(vnode => vnode.type && vnode.type !== Symbol.for('v-fgt'));
        return h('table', {},
          (props.data as any[]).map((row, rowIndex) => h('tr', { key: rowIndex },
            columns.map((column: any, columnIndex: number) => {
              const renderCell = typeof column.children?.default === 'function'
                ? column.children.default
                : null;
              return h('td', {
                key: `${rowIndex}-${columnIndex}`,
                'data-field': column.props?.field ?? '',
                'data-title': column.props?.title ?? '',
              }, renderCell ? renderCell({ row }) : row[column.props?.field]);
            }))));
      };
    },
  });

  return { Table, TableColumn };
});

const messages = {
  'zh-CN': {
    table: {
      loading: '加载中',
      empty: '暂无数据',
    },
    taskDetail: {
      table: {
        retry: '重试',
      },
    },
    platform: {
      nodeMan: {
        log: {
          searchNode: '搜索节点',
          searchPlugin: '搜索插件',
          step: '步骤',
          costTime: '耗时',
          executionStatus: '执行情况',
          terminate: '终止',
          viewSubWorkflow: '查看子工作流',
          executionLog: '执行日志',
          executionLogOf: '{inner}{type}执行日志',
          waitOfflineOperation: '等待离线操作',
          offlineGuide: '离线指引',
          waitManualOperation: '等待人工操作',
          operationGuide: '操作指引',
          terminateConfirmTitle: '确认终止',
          terminateConfirmSubTitle: '确认终止当前任务',
          terminateFailedMsg: '终止失败',
        },
        taskHistory: {
          taskType: {
            install_plugin: '插件安装',
          },
        },
      },
    },
  },
};

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/history/detail/:taskId', name: 'taskDetail', component: defineComponent({ template: '<div />' }) },
    { path: '/history/detail/:taskId/log/:hostId', name: 'log', component: Log },
  ],
});

async function mountPage() {
  window.PROJECT_CONFIG = {
    ENABLE_NOTICE: 'false',
  } as any;

  const pinia = createPinia();
  setActivePinia(pinia);

  const mainStore = useMainStore();
  const nodeManageStore = useNodeManageStore();
  mainStore.updateLanguage('zh-CN');
  nodeManageStore.updateCurrentRowData({
    type: 'install_plugin',
  } as any);

  await router.push({
    name: 'log',
    params: {
      taskId: 'wf:test-parent',
      hostId: '1',
    },
  });
  await router.isReady();

  const wrapper = mount(Log, {
    global: {
      plugins: [
        pinia,
        router,
        createI18n({
          legacy: false,
          locale: 'zh-CN',
          messages,
        }),
      ],
      stubs: {
        guide: defineComponent({
          name: 'guide',
          setup() {
            return () => h('div');
          },
        }),
        'page-header': defineComponent({
          name: 'page-header',
          setup(_, { slots }) {
            return () => h('div', slots.default?.());
          },
        }),
        'bk-loading': defineComponent({
          name: 'bk-loading',
          setup(_, { slots }) {
            return () => h('div', slots.default?.());
          },
        }),
        'bk-overflow-title': defineComponent({
          name: 'bk-overflow-title',
          setup(_, { slots }) {
            return () => h('div', slots.default?.());
          },
        }),
      },
      directives: {
        'bk-tooltips': () => undefined,
      },
    },
  });

  await flushPromises();
  await nextTick();
  await flushPromises();

  return wrapper;
}

describe('node history log sub-workflow action', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders sub-workflow entry in execution status column instead of the rightmost action column', async () => {
    const wrapper = await mountPage();

    expect(wrapper.find('td[data-field="state"] [data-test="sub-workflow-entry"]').exists()).toBe(true);
    expect(wrapper.find('td[data-field=""]').text()).not.toContain('查看子工作流');
  });
});
