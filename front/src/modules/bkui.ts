import bkui from 'bkui-vue';

import 'bkui-vue/dist/style.css';
import type { UserModule } from '@/types';

// Setup BKUI
export const install: UserModule = ({ app }) => {
  app.use(bkui);
};
