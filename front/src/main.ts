import { createApp } from 'vue';

import App from './App.vue';
import type { UserModule } from './types.ts';

import '@blueking/table/vue3/vue3.css';
import '@unocss/reset/tailwind.css';
import './styles/main.css';
import 'uno.css';

const app = createApp(App);
// 安装modules下面所有模块
Object.values(import.meta.glob<{ install: UserModule }>('./modules/*.ts', { eager: true }))
  .forEach(i => i.install?.({ app }));

// 挂载
app.mount('#app');
