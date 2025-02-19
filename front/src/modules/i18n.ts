import type { Locale } from 'vue-i18n';
import { createI18n } from 'vue-i18n';

import type { UserModule } from '@/types.ts';

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

// Setup i18n
const install: UserModule = async ({ app }) => {
  app.use(i18n);
  // 加载默认语言
  await loadLanguageAsync('zh-CN');
};

export {
  i18n,
  availableLocales,
  loadLanguageAsync,
  install,
};
