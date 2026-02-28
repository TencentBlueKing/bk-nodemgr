
import { useI18n } from 'vue-i18n';

import { getPlatformConfig, setDocumentTitle, setShortcutIcon } from '@blueking/platform-config';

import { usePlatformConfigStore } from '@/stores/platform-config';
export default function usePlatform() {
  const platformConfig = usePlatformConfigStore();

  async function getPlatformInfo() {
    const { t } = useI18n();

    // 根据环境设置版本信息
    const getVersionByEnv = () => {
      const env = import.meta.env.BK_NODE_ENV;
      if (env === 'production') {
        return window.PROJECT_CONFIG.APP_VERSION;
      }
      return '';
    };

    const defaults = {
      name: '节点管理',
      nameEn: 'Nodemgr',
      appLogo: '/nodeman.png',
      brandName: '蓝鲸智云',
      brandNameEn: 'Tencent BlueKing',
      productName: '蓝鲸节点管理',
      productNameEn: 'BK Nodemgr',
      favicon: '/favicon.svg',
      helperLink: 'wxwork://message?uin=8444252571319680',
      helperText: t('platform.onCall'),
      footerInfoHTML: '',
      version: getVersionByEnv(),
      i18n: {
        footerInfoHTML: '',
      },
    };
    let data: Record<string, any> = {};
    if (import.meta.env.BK_SHARED_RES_BASE_JS_URL) {
      data = await getPlatformConfig(import.meta.env.BK_SHARED_RES_BASE_JS_URL, defaults);
    } else {
      data = await getPlatformConfig(defaults);
    }
    Object.keys(platformConfig.$state).forEach((key) => {
      platformConfig.$patch({
        [key]: data[key],
      });
    });
    return data;
  };
  return {
    platformConfig,
    getPlatformInfo,
    setDocumentTitle,
    setShortcutIcon,
  };
}
