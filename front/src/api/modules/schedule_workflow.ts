// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { ScheduleWorkflowListReq, ScheduleWorkflowListResp, ScheduleWorkflowOperationInstanceListReq, ScheduleWorkflowOperationInstanceListResp, ScheduleWorkflowOperationInstanceLogGetReq, ScheduleWorkflowOperationInstanceLogGetResp, ScheduleWorkflowOperationListReq, ScheduleWorkflowOperationListResp, ScheduleWorkflowUpdateIntervalReq, ScheduleWorkflowUpdateIntervalResp } from '@/@types/schedule_workflow';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const ScheduleWorkflowService = {
  // 修改定时工作流的周期配置，由现有调度同步机制异步生效；不重置上次触发时间或中止正在执行的实例。
  ScheduleWorkflowUpdateInterval: async <Request = ScheduleWorkflowUpdateIntervalReq, ResponseData = ScheduleWorkflowUpdateIntervalResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/schedule/workflow/interval/update')(params, config),
  // 查询当前租户的周期工作流调度配置。
  ScheduleWorkflowList: async <Request = ScheduleWorkflowListReq, ResponseData = ScheduleWorkflowListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/schedule/workflow/list')(params, config),
  // 查询周期工作流根操作，或按父实例下钻直接子操作。
  ScheduleWorkflowOperationList: async <Request = ScheduleWorkflowOperationListReq, ResponseData = ScheduleWorkflowOperationListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/schedule/workflow/operation/list')(params, config),
  // 按操作 ID 或实例 ID 查询周期工作流执行实例及历史状态。
  ScheduleWorkflowOperationInstanceList: async <Request = ScheduleWorkflowOperationInstanceListReq, ResponseData = ScheduleWorkflowOperationInstanceListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/schedule/workflow/operation/instance/list')(params, config),
  // 按实例 ID 查询周期工作流步骤状态、双语日志及子工作流引用。
  ScheduleWorkflowOperationInstanceLogGet: async <Request = ScheduleWorkflowOperationInstanceLogGetReq, ResponseData = ScheduleWorkflowOperationInstanceLogGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/schedule/workflow/operation/instance/log/get')(params, config),
};
