import { type RouteRecordName } from 'vue-router';

import type { Config } from './interceptors';

export type RouteName = RouteRecordName | null | undefined;

export interface IQueue {
  id: string
  controller: AbortController
  request: Promise<any>
  config?: Config
  routeName?: RouteName
}

// 请求队列
const requestQueue: Array<IQueue> = [];
// 移除队列
function removeQueue(id: string) {
  const index = requestQueue.findIndex(q => q.id === id);
  if (index > -1) {
    requestQueue.splice(index, 1);
  }
}
// 添加队列
function addQueue(data: IQueue) {
  const index = requestQueue.findIndex(q => q.id === data?.id);
  if (index === -1) {
    requestQueue.push(data);
  }
}
// 清空队列（不取消请求）
function clearQueue(id?: string|string[]) {
  if (id?.length) {
    const ids = Array.isArray(id) ? id : [id];
    ids.forEach((id) => {
      const index = requestQueue.findIndex(queue => queue.id === id);
      index > -1 && requestQueue.splice(index, 1);
    });
  } else {
    requestQueue.length = 0;
  }
}
// 取消队列请求
async function cancelRequest(id?: string | string[]) {
  let queues = requestQueue.filter(queue => !queue.config?.irrevocable); // 过滤配置了不可取消请求的配置
  if (id?.length) {
    const ids = Array.isArray(id) ? id : [id];
    queues = queues.filter(queue => ids.includes(queue.id));
  }
  queues.forEach(queue => queue.controller?.abort());
  await Promise.all(queues.map(item => item.request)).catch(() => {});
  clearQueue(id);
}

export {
  requestQueue, // 内部数据结构，只读模式
  removeQueue,
  addQueue,
  clearQueue,
  cancelRequest,
};
