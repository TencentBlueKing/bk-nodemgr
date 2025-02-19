import { bkTooltips } from 'bkui-vue';

import type { UserModule } from '@/types';

// 注册全局组件或者指令
export const install: UserModule = ({ app }) => {
  app.directive('bk-tooltips', bkTooltips);
};
