// pina 持久化存储
import { get, has, merge, set } from 'lodash';
import type { PiniaPluginContext, Store } from 'pinia';
import { createPinia } from 'pinia';

import { STORAGE_KEY, STORAGE_VERSION } from '@/common/const';
import type { UserModule } from '@/types.ts';

// 缓存接口定义
interface Storage {
  getItem: (key: string) => any;
  setItem: (key: string, value: any) => void;
  removeItem: (key: string) => void;
  clear: () => void;
}
// 本地缓存配置信息
interface Options {
  key: string;
  storage: Storage;
  paths: string[];
  ids: string[];
  version: string;
  reducer: (state: Store, paths: string[]) => object;
  getState: (key: string, storage: Storage) => any;
  setState: (key: string, state: object, storage: Storage) => void;
  assertStorage?: (storage: Storage) => void | Error;
}

function reducer(state: Store, paths: string[]) {
  return Array.isArray(paths)
    ? paths.reduce((substate, path) => set(substate, path, get(state, path)), {})
    : state;
}
// 获取缓存数据
function getState(key = '_pina_', storage: Storage = window.localStorage) {
  const value = storage.getItem(key);

  try {
    return !!value ? JSON.parse(value) : value;
  } catch (err) {}

  return undefined;
}
// 设置缓存状态
function setState(key = '_pina_', state: object, storage: Storage) {
  return storage.setItem(key, JSON.stringify(state));
}
// 校验缓存是否可用
function assertStorage(storage: Storage = window.localStorage) {
  storage.setItem('@@', 1);
  storage.removeItem('@@');
}

function installPiniaStorage(opt: Partial<Options> = {}) {
  const options: Options = merge({
    key: '_pina_',
    version: '',
    overwrite: false,
    storage: window.localStorage,
    paths: [],
    ids: [],
    reducer,
    getState,
    setState,
    assertStorage,
  }, opt);

  assertStorage(options.storage);
  const savedState = options.getState(options.key, options.storage);

  return ({ store }: PiniaPluginContext) => {
    if (!options.ids?.includes(store.$id)) return;
    // 还原localstorage里面的值
    if (typeof savedState === 'object' && savedState !== null) {
      Object.keys(savedState).forEach((key) => {
        if (has(store, key)) {
          set(store, key, savedState[key]);
        }
      });
    }

    // store 变化
    store.$subscribe(() => {
      options.setState(
        options.key,
        options.reducer(store, options.paths),
        options.storage,
      );
    });
  };
};

// Setup Pinia
// https://pinia.vuejs.org/
export const install: UserModule = ({ app }) => {
  const pinia = createPinia().use(installPiniaStorage({
    key: STORAGE_KEY,
    version: STORAGE_VERSION,
    ids: ['user'],
  }));
  app.use(pinia);
};
