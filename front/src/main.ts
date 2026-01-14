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

const app = createApp(App);
// 安装modules下面所有模块
Object.values(import.meta.glob<{ install: UserModule }>('./modules/*.ts', { eager: true }))
  .forEach(i => i.install?.({ app }));

// 挂载
app.mount('#app');
