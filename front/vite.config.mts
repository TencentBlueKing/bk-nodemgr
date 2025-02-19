import path from 'node:path';
import Unocss from 'unocss/vite';
import Components from 'unplugin-vue-components/vite';
import { defineConfig, loadEnv } from 'vite';
import VueDevTools from 'vite-plugin-vue-devtools';
import Layouts from 'vite-plugin-vue-layouts';

import VueI18n from '@intlify/unplugin-vue-i18n/vite';
import Vue from '@vitejs/plugin-vue';
import VueJsx from '@vitejs/plugin-vue-jsx';

export default ({ mode }: { mode: string }) => {
  const env = loadEnv(mode, process.cwd(), 'BK_');

  return defineConfig({
    base: env.BK_STATIC_URL,

    server: {
      host: env.BK_APP_HOST,
      port: Number(env.BK_APP_PORT),
      // https: {},
      open: true,
      proxy: {
        '/api': {
          target: env.BK_API_BASE_URL,
          changeOrigin: true,
          secure: true,
          toProxy: true,
          headers: {
            referer: env.BK_API_BASE_URL,
          },
        },
      },
    },

    // 环境变量前缀
    envPrefix: 'BK_',

    // 路径别名
    resolve: {
      alias: {
        '@/': `${path.resolve(__dirname, 'src')}/`,
      },
    },

    plugins: [
      // https://github.com/vitejs/vite-plugin-vue/tree/main/packages/plugin-vue
      Vue(),

      // https://github.com/vitejs/vite-plugin-vue/tree/main/packages/plugin-vue-jsx
      VueJsx(),

      // https://github.com/JohnCampionJr/vite-plugin-vue-layouts
      Layouts(),

      // https://github.com/antfu/unplugin-vue-components
      Components({
        // allow auto load markdown components under `./src/components/`
        extensions: ['vue', 'md'],
        // allow auto import and register components used in markdown
        include: [/\.vue$/, /\.vue\?vue/, /\.md$/],
        dts: 'src/components.d.ts',
      }),

      // https://github.com/antfu/unocss
      // see uno.config.ts for config
      Unocss(),

      // https://github.com/intlify/bundle-tools/tree/main/packages/unplugin-vue-i18n
      VueI18n({
        runtimeOnly: true,
        compositionOnly: true,
        fullInstall: true,
        include: [path.resolve(__dirname, 'locales/**')],
      }),

      // https://github.com/webfansplz/vite-plugin-vue-devtools
      VueDevTools(),
    ],

    // https://github.com/vitest-dev/vitest
    test: {
      include: ['test/**/*.test.ts'],
      environment: 'jsdom',
    },

    ssr: {
      // TODO: workaround until they support native ESM
      noExternal: ['workbox-window', /vue-i18n/],
    },
  });
};
