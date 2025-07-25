import { Message } from 'bkui-vue';
import { isObject, merge } from 'lodash';
import { loginModal } from '@/common/auth';
import { type Config, fetch, interceptors  } from './interceptors';

type HttpMethods = 'GET' | 'POST' | 'PATCH' | 'DELETE' | 'PUT';

interceptors.response.use(async (response: Response, config: Config) => {
  const res = await resultReduction(response.clone(), config).catch(() => ({}));
  let resData;
  if (config.originalResponse) {
    resData = response;
  } else if (config.needRes) {
    resData = res;
  } else {
    resData = res.data;
  }
  if (config.responseType && ['blod', 'text'].includes(config.responseType)) return resData;
  // todo 未认证
  if (response.status === 401) {
    if (res.login_url) {
      // window.location.href = `${res.login_url}?c_url=${window.location.href}`;
      loginModal();
    } else {
      Message({
        theme: 'error',
        message: {
          code: response.status,
          overview: '登录URL为空',
          suggestion: '',
        },
      });
    }
    return;
  }

  const showMessageData = res?.message || res?.datas?.message || '';
  if (response.status >= 200 && response.status < 300) {
    // API状态码不正确
    if (config.validateCode && res.status !== 0 && res.code !== 0) {
      // 优化后的Messagea
      config.interceptorErr && Message({
        theme: 'error',
        message: {
          code: res.status ?? res.code,
          overview: showMessageData,
          suggestion: '',
          type: 'key-value',
        },
      });
      return Promise.reject(resData);
    };


    return resData;
  }

  // 优化后的Message
  config.interceptorErr && Message({
    theme: 'error',
    message: {
      code: response.status,
      overview: showMessageData,
      suggestion: '',
      type: 'key-value',
      details: {
        bizId: response.headers.get('X-Devops-Rid'),
        message: showMessageData,
        url: response.url,
      },
    },
  });

  return Promise.reject(resData);
});

async function resultReduction(response: Response, config: Config) {
  let res: any;
  switch (config.responseType) {
    case 'json':
      res = await response.json();
      break;
    case 'text':
      res = await response.text();
      break;
    case 'blod':
      res = await response.blob();
      break;
    default:
      res = await response.json();
      break;
  }
  return res;
}

export default class ConsoleFetch {
  config: Config;// 全局配置

  constructor(config: Config) {
    this.config = merge({
      mode: 'cors',
      cache: 'default',
      credentials: 'include',
      headers: {
        'X-Requested-With': 'fetch',
        'Content-Type': 'application/json',
      },
      redirect: 'follow',
      referrerPolicy: 'no-referrer-when-downgrade',
      responseType: 'json',
      validateCode: true,
      interceptorErr: true,
    }, config);
  }
  // 替换URL上的变量和删除params上的变量参数
  parseUrlAndParams(url: string, params = {}) {
    const variableData: Record<string, any> = {

    };
    Object.keys(params).forEach((key) => {
      // 自定义url变量
      if (key.indexOf('$') === 0) {
        variableData[key] = params[key as keyof typeof params];
      }
    });
    let newUrl = url;
    Object.keys(variableData).forEach((key) => {
      if (!variableData[key]) {
        // console.warn(`路由变量未配置${key}`, url);
        // 去除后面的路径符号
        newUrl = newUrl.replace(`/${key}`, '');
      }
      newUrl = newUrl.replace(new RegExp(`\\${key}\\b`, 'g'), variableData[key]);
      // 删除URL上的参数
      delete params[key as keyof typeof params];
    });
    return {
      url: newUrl,
      params,
    };
  }
  async request<T, C extends Config>(method: HttpMethods, url: string, params?: any, config?: C) {
    const fetchConfig = merge(
      {},
      this.config,
      {
        headers: {},
      },
      config || {},
    );
    let body;
    const parseData = this.parseUrlAndParams(`${fetchConfig.prefix}${url}`, isObject(params) ? params : {});
    // GET请求参数放URL，其余请求放在body里面
    if (method === 'GET') {
      const query = new URLSearchParams(parseData.params || {}).toString();
      parseData.url += query ? `?${query}` : '';
    } else {
      body = isObject(params) ? JSON.stringify(parseData.params || {}) : params;
    }

    const response = await fetch<T, C>(parseData.url, {
      method,
      ...fetchConfig,
      body,
    } as any);

    return response;
  }

  // P 请求参数类型 T 返回类型
  get<P, T>(url: string) {
    return <C extends Config>(params?: P, config?: C) => this.request<T, C>('GET', url, params, config);
  }

  post<P, T>(url: string) {
    return <C extends Config>(params?: P, config?: C) => this.request<T, C>('POST', url, params, config);
  }

  patch<P, T>(url: string) {
    return <C extends Config>(params?: P, config?: C) => this.request<T, C>('PATCH', url, params, config);
  }

  delete<P, T>(url: string) {
    return <C extends Config>(params?: P, config?: C) => this.request<T, C>('DELETE', url, params, config);
  }

  put<P, T>(url: string) {
    return <C extends Config>(params?: P, config?: C) => this.request<T, C>('PUT', url, params, config);
  }
}
