import { merge } from 'lodash';

import { addQueue, removeQueue } from '@/api/request-queue';

export type Config = {
  prefix?: string
  needRes?: boolean
  responseType?: 'json' | 'text' | 'blod'
  irrevocable?: boolean // 请求不能取消
  id?: string // 唯一ID
  originalResponse?: boolean // 返回原始res对象
  validateCode?: boolean // 校验code是否正确，默认true
  interceptorErr?: boolean // 是否自动拦截http异常弹出message
} & RequestInit;

export type ResCallback= (res: Response, config: Partial<Config>) => any;

export type FetchReturnType<T, P extends Config> = P['originalResponse'] extends true ? Response : T;

const interceptorsReq: Function[] = [];
const interceptorsRes: Array<ResCallback> = [];
const interceptorsResError: Function[] = [];

const OriginFetch = window.fetch;

let seq = 0;
/** 生成 UUID v4 格式的唯一请求 ID（时间戳+序号+随机数），确保不重复 */
function genUUID(): string {
  try {
    const ts = Date.now().toString(16).padStart(12, '0');
    const sn = (++seq % 0x10000).toString(16).padStart(4, '0');
    const rd = crypto.getRandomValues(new Uint8Array(4));
    const r = Array.from(rd).map(b => b.toString(16).padStart(2, '0')).join('');
    return `${ts.slice(0, 8)}-${ts.slice(8)}-4${sn.slice(1)}-a${r.slice(0, 3)}-${sn.slice(0, 1)}${r.slice(3)}`;
  } catch {
    const ts = Date.now().toString(16).padStart(12, '0');
    const sn = (++seq % 0x10000).toString(16).padStart(4, '0');
    const rd = Math.random().toString(16).slice(2, 10).padEnd(8, '0');
    return `${ts.slice(0, 8)}-${ts.slice(8)}-4${sn.slice(1)}-8${rd.slice(0, 3)}-${sn.slice(0, 1)}${rd.slice(3, 7)}`;
  }
}

// fetch
function fetch<T, C extends Config>(input: RequestInfo | URL, init: Partial<C>) {
  interceptorsReq.forEach((fn) => {
    init = fn(init);
  });

  return new Promise<FetchReturnType<T, C>>((resolve, reject) => {
    const controller = new AbortController();
    const BK_REQUEST_ID_HEADER_KEY = window.PROJECT_CONFIG.BK_REQUEST_ID_HEADER_KEY.includes('BK_REQUEST_ID_HEADER_KEY')
      ? 'X-Bkapi-Request-Id'
      : window.PROJECT_CONFIG.BK_REQUEST_ID_HEADER_KEY;
    const requestID = genUUID();
    const defaultHeaders: Record<string, string> = {
      'cess-Control-Allow-Origin': '*',
    };
    defaultHeaders[BK_REQUEST_ID_HEADER_KEY] = requestID;
    const request = OriginFetch(
      input,
      merge(
        {
          headers: defaultHeaders,
          signal: controller.signal,
        },
        init,
      ),
    )
      .then((res) => {
        interceptorsRes.forEach((fn) => {
          res = fn(res, init);
        });
        resolve(res as FetchReturnType<T, C>);
        removeQueue(init?.id || requestID);
      })
      .catch((err) => {
        // 请求被取消（AbortError）时静默处理，不触发错误拦截器和 reject
        if (err?.name === 'AbortError') {
          removeQueue(init?.id || requestID);
          return;
        }
        interceptorsResError.forEach((fn) => {
          err = fn(err, init);
        });
        reject(err);
        removeQueue(init?.id || requestID);
      });

    // const route = useRoute();
    addQueue({
      id: init?.id || requestID,
      controller,
      request,
      config: init,
      // routeName: route?.name,
    });
  });
}

const interceptors = {
  request: {
    use(callback: Function) {
      interceptorsReq.push(callback);
    },
  },
  response: {
    use(callback: ResCallback, errorCallback?: Function) {
      interceptorsRes.push(callback);
      errorCallback && interceptorsResError.push(errorCallback);
    },
  },
};

export {
  interceptors,
  fetch,
};
