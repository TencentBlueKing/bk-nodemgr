import { merge, uniqueId } from 'lodash';

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
    const requestID = uniqueId();
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
        removeQueue(requestID);
      })
      .catch((err) => {
        interceptorsResError.forEach((fn) => {
          err = fn(err, init);
        });
        reject(err);
        removeQueue(requestID);
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
