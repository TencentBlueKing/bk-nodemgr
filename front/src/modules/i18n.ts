import { ref } from 'vue';
import type { Locale } from 'vue-i18n';
import { createI18n } from 'vue-i18n';

import { parseCookies } from '@/common/util';
import type { UserModule } from '@/types.ts';

// i18n 语言包就绪状态，语言包加载完成前不渲染应用内容
const i18nReady = ref(false);

const i18n = createI18n({
  legacy: false,
  locale: '',
  messages: {},
});

// locales目录下语言map
const localesMap = Object.fromEntries(Object.entries(import.meta.glob('../../locales/*.yml'))
  .map(([path, loadLocale]) => [path.match(/([\w-]*)\.yml$/)?.[1], loadLocale])) as Record<Locale, () => Promise<{ default: Record<string, string> }>>;

const availableLocales = Object.keys(localesMap);
// 已加载语言
const loadedLanguages: string[] = [];

// 设置语言
function setI18nLanguage(lang: Locale) {
  i18n.global.locale.value = lang;
  if (typeof document !== 'undefined') document.querySelector('html')?.setAttribute('lang', lang);
  return lang;
}

// 异步加载语言
async function loadLanguageAsync(lang: string): Promise<Locale> {
  if (i18n.global.locale.value === lang) return setI18nLanguage(lang);

  if (loadedLanguages.includes(lang)) return setI18nLanguage(lang);

  const messages = await localesMap[lang]();
  i18n.global.setLocaleMessage(lang, messages.default);
  loadedLanguages.push(lang);
  return setI18nLanguage(lang);
}

// 获取当前语言设置
function getCurrentLanguage(): string {
  const cookies = parseCookies();
  let currentLang = cookies.blueking_language || 'zh-cn';

  // 标准化语言标识
  if (['zh-CN', 'zh-cn', 'cn', 'zhCN', 'zhcn', 'None', 'none'].indexOf(currentLang) > -1) {
    currentLang = 'zh-CN';
  } else {
    currentLang = 'en-US';
  }

  // 确保语言在可用语言列表中
  return availableLocales.includes(currentLang) ? currentLang : 'zh-CN';
}

// Setup i18n
const install: UserModule = async ({ app }) => {
  app.use(i18n);
  // 根据cookie中的语言设置加载对应语言
  const currentLang = getCurrentLanguage();
  try {
    await loadLanguageAsync(currentLang);
  } finally {
    // 无论成功或失败，均标记就绪，防止永久白屏
    i18nReady.value = true;
  }
};

export {
  i18n,
  i18nReady,
  availableLocales,
  loadLanguageAsync,
  install,
};
