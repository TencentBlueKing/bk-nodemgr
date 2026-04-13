// gen-api.js 自动生成，请勿手动修改
// AuthResource describes a single resource instance for authorization checks.
export interface AuthResource {
  // system_id is the IAM system that owns this resource (e.g. "bk_cmdb" or
  // "bk_nodemgr").
  system_id: string;
  // type is the resource type identifier (e.g. "biz", "networkarea").
  type: string;
  // id is the resource instance ID.
  id: string;
}

// AuthVerifyItem describes a single action-resource pair for permission
// verification.
export interface AuthVerifyItem {
  // action is the IAM action identifier (e.g. "agent_view",
  // "networkarea_create").
  action: string;
  // resources lists the resources to check for this action. Empty for
  // action-level checks.
  resources: AuthResource[];
}

// AuthVerifyReq describes the HTTP request body for auth verification.
export interface AuthVerifyReq {
  // items lists the action-resource pairs to verify permissions for.
  items: AuthVerifyItem[];
}

// AuthVerifyResult holds the authorization result for a single action.
export interface AuthVerifyResult {
  // action is the IAM action that was checked.
  action: string;
  // authorized indicates whether the user has this permission.
  authorized: boolean;
}

// AuthVerifyResp describes the HTTP response body for auth verification.
export interface AuthVerifyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: AuthVerifyRespData;
}

export interface AuthVerifyRespData {
  results: AuthVerifyResult[];
}

// AuthorizedItem describes a single action-resource type pair for querying
// authorized scope.
export interface AuthorizedItem {
  // action is the IAM action identifier (e.g. "agent_view",
  // "networkarea_create").
  action: string;
  // resource_type is the resource type to query authorized scope for (e.g.
  // "biz", "networkarea").
  resource_type: string;
}

// AuthorizedReq describes the HTTP request body for querying authorized
// resource scope.
export interface AuthorizedReq {
  // items lists the action-resource type pairs to query authorized scope
  // for.
  items: AuthorizedItem[];
}

// AuthorizedResult holds the authorized scope for a single action-resource
// type pair.
export interface AuthorizedResult {
  // action is the IAM action that was queried.
  action: string;
  // resource_type is the resource type that was queried.
  resource_type: string;
  // is_any indicates whether the user has unrestricted access to all
  // resources of this type.
  is_any: boolean;
  // resources lists the authorized resource instances. Empty when is_any is
  // true.
  resources: AuthResource[];
}

// AuthorizedResp describes the HTTP response body for authorized scope
// query.
export interface AuthorizedResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: AuthorizedRespData;
}

export interface AuthorizedRespData {
  results: AuthorizedResult[];
}

