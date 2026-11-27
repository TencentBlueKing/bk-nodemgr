// gen-api.js 自动生成，请勿手动修改
// 全局共享的已存储配置项，包含自定义配置。
export interface GlobalSetting {
  // 配置变量名称。
  setting_name: string;
  // 数据库保存的原始字符串，不补齐默认值，也不表示进程实际生效值。
  value: string;
}

// 按配置变量名称筛选全局共享的已存储配置。
export interface ListGlobalSettingsReq {
  // 分页条件；非计数查询时 limit 必须为 1 到 500。
  page: Page;
  // 是否仅返回符合条件的配置项总数。
  only_count: boolean;
  // 配置变量名称的精确包含条件。
  exact_include_conditions: ListGlobalSettingsReqExactConditions;
  // 配置变量名称的模糊包含条件。
  fuzzy_include_conditions: ListGlobalSettingsReqFuzzyConditions;
  // 配置变量名称的精确排除条件。
  exact_exclude_conditions: ListGlobalSettingsReqExactConditions;
  // 配置变量名称的模糊排除条件。
  fuzzy_exclude_conditions: ListGlobalSettingsReqFuzzyConditions;
}

// 配置变量名称的精确匹配条件。
export interface ListGlobalSettingsReqExactConditions {
  setting_name: string[];
}

// 配置变量名称的模糊匹配条件。
export interface ListGlobalSettingsReqFuzzyConditions {
  setting_name: string[];
}

// 全局共享配置的已存储原始值，不包含服务启动配置。
export interface ListGlobalSettingsResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ListGlobalSettingsRespData;
}

export interface ListGlobalSettingsRespData {
  // 符合筛选条件的配置项总数。
  total: number;
  // 当前页的配置项；仅计数时为空数组。
  items: GlobalSetting[];
}
