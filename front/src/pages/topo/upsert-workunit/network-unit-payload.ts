import type { TopoNetworkUnitCreateReq, TopoNetworkUnitUpdateReq } from '@/@types/topo';

import { serializeCustomDeployConfigs } from './custom-deploy-config';

interface NetworkUnitPayloadForm {
  tenant_id: string;
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  generation: number | string;
  custom_deploy_configs: Record<string, CustomDeployConfig>;
  accesspoints: AccessPoint[];
  direct_endpoints: Endpoints;
}

interface BuildNetworkUnitPayloadOptions {
  form: NetworkUnitPayloadForm;
  links: Links;
  isCreate: boolean;
  workUnitId: number;
  type: string;
  isDirect: boolean;
}

interface BuildBaseNetworkUnitPayloadOptions {
  form: NetworkUnitPayloadForm;
  links: Links;
  isCreate: boolean;
  type: string;
  isDirect: boolean;
}

const buildBaseNetworkUnitPayload = ({
  form,
  links,
  isCreate,
  type,
  isDirect,
}: BuildBaseNetworkUnitPayloadOptions) => ({
  bk_networkunit_name: form.bk_networkunit_name,
  bk_networkarea_id: form.bk_networkarea_id,
  custom_deploy_config: serializeCustomDeployConfigs(form.custom_deploy_configs),
  accesspoints: form.accesspoints.map((item: AccessPoint) => ({
    ...item,
    accesspoint_id: item.accesspoint_id ?? -1,
  })),
  direct_endpoints: form.direct_endpoints,
  links,
  is_direct: isCreate ? type === 'direct' : isDirect,
});

export const buildNetworkUnitPayload = ({
  form,
  links,
  isCreate,
  workUnitId,
  type,
  isDirect,
}: BuildNetworkUnitPayloadOptions): TopoNetworkUnitCreateReq | TopoNetworkUnitUpdateReq => {
  const basePayload = buildBaseNetworkUnitPayload({
    form,
    links,
    isCreate,
    type,
    isDirect,
  });

  if (isCreate) {
    return {
      ...basePayload,
      generation: Number(form.generation),
    };
  }

  return {
    networkunit: {
      ...basePayload,
      tenant_id: form.tenant_id,
      generation: Number(form.generation),
      bk_networkunit_id: workUnitId,
    },
    fields: {
      bk_networkunit_name: true,
      accesspoints: true,
      links: true,
      direct_endpoints: true,
      custom_deploy_config: true,
    },
  };
};