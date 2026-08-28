import { createApp } from 'vue';

import App from './App.vue';
import type { UserModule } from './types.ts';

import '@/fonts/iconcool';
import '@blueking/table/vue3/vue3.css';
import '@blueking/vxe-table/lib/style.css';
import 'tippy.js/dist/tippy.css';
import 'tippy.js/themes/light.css';
import '@unocss/reset/tailwind.css';
import './styles/main.css';
import 'uno.css';
import './fonts/style.css';

// 忽略 AbortError 报错
window.addEventListener('unhandledrejection', (event) => {
  if (event.reason?.name === 'AbortError') {
    event.preventDefault();
  }
});

// 修复多租户下 blueking_language 同名 cookie 冲突
// 登录系统的语言 cookie 统一挂在第二段开始的子域名下，与当前应用不同域会导致两个同名 cookie 共存，
// ip-selector 库按 cookie 严格 === 'en' 判断，会走英文语言包，与项目中文模式冲突。
// 此处统一把语言 cookie 设置到第二段开始的子域名下，与登录系统保持一致。
(function normalizeBluekingLanguageCookie() {
  const lang = (navigator.language || 'zh-CN').toLowerCase().startsWith('en') ? 'en' : 'zh-cn';
  const host = location.hostname;
  const parent = host.replace(/^[^.]+\./, '');
  const domain = parent !== host ? `;domain=.${parent}` : '';
  document.cookie = `blueking_language=${lang};path=/${domain}`;
})();

const app = createApp(App);

// 过滤第三方库（@blueking/bkui-form、@blueking/ip-selector）的已知兼容性 warning，不影响功能
app.config.warnHandler = (msg, _instance, trace) => {
  // bkui-form 内部 resolveComponent 传入空名，产生 "Failed to resolve component: " 后跟换行
  if (msg.startsWith('Failed to resolve component:') && /^Failed to resolve component:\s*$/.test(msg.split('\n')[0])) return;
  const ignorePatterns = [
    'provide() can only be used inside setup()',
    'Failed to resolve component: IpSelector',
    'Failed to resolve component: bk-ip-selector',
    'Invalid prop: custom validator check failed for prop "min"',  // bkui-form NumberField 内部传了无效 min
  ];
  if (ignorePatterns.some(p => msg.includes(p))) return;
  // eslint-disable-next-line no-console
  console.warn(`[Vue warn]: ${msg}${trace ? `\n${trace}` : ''}`);
};

// 安装modules下面所有模块
Object.values(import.meta.glob<{ install: UserModule }>('./modules/*.ts', { eager: true }))
  .forEach(i => i.install?.({ app }));

// 挂载
app.mount('#app');
