import { showLoginModal } from '@blueking/login-modal';
import { InfoBox } from 'bkui-vue';
import { h } from 'vue';

interface ILoginData {
  loginUrl?: string;
}
// 获取登录地址
const getLoginUrl = (url: string, cUrl: string, isFromLogout: boolean) => {
  const loginUrl = new URL(url);
  if (isFromLogout) {
    loginUrl.searchParams.append('is_from_logout', '1');
  }
  loginUrl.searchParams.append('c_url', cUrl);
  return loginUrl.href;
};

// 跳转到登录页
export const login = (data: ILoginData = {}) => {
  location.href = data.loginUrl || getLoginUrl(window.PROJECT_CONFIG.BK_LOGIN_URL, location.origin, false);
};

// 打开登录弹框
export const loginModal = async () => {
  const loginUrl = getLoginUrl(
    `${window.PROJECT_CONFIG.BK_LOGIN_URL}/plain`,
    `${location.origin + window.PROJECT_CONFIG.SITE_URL}/static/login_success.html`,
    false,
  );

  window.loginModal = await showLoginModal({ loginUrl });
  if (!window.loginModal) {
    const target =
      'https://support.google.com/chrome/answer/95472?hl=zh-Hans&co=GENIE.Platform%3DDesktop';
    InfoBox({
      title: 'HTTP 401: Authorization Required',
      content: h(
        'div',
        {
          style: 'color: #000; font-size: 14px; line-height: 1.5; margin-bottom: 10px;',
        },
        [
          '登录弹出失败，请检查当前浏览器是否设置为阻止弹出窗口，详情请参考：',
          h(
            'a',
            {
              href: target,
              target: '_blank',
              style: 'color: #3a84ff; text-decoration: underline;',
            },
            '在 Chrome 中阻止或允许显示弹出式窗口',
          ),
        ],
      ),
      type: 'warning',
    });
  }
};

// 退出登录
export const logout = () => {
  window.location.replace(getLoginUrl(window.PROJECT_CONFIG.BK_LOGIN_URL, location.origin, true));
};
