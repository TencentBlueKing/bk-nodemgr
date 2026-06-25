import AlibabaCloud from '../../../../public/images/ali-cloud.svg';
import AWS from '../../../../public/images/aws.svg';
import Azure from '../../../../public/images/azure.svg';
import BaiduCloud from '../../../../public/images/baidu-cloud.svg';
import CapitalCloud from '../../../../public/images/capital-cloud.svg';
import defaultCloud from '../../../../public/images/default-cloud.svg';
import ECloud from '../../../../public/images/e-cloud.svg';
import EnterprisePrivateCloud from '../../../../public/images/enterprise-private-cloud.svg';
import GoogleCloud from '../../../../public/images/google-cloud.svg';
import HuaweiCloud from '../../../../public/images/huawei-cloud.svg';
import IBMCloud from '../../../../public/images/ibm-cloud.svg';
import KSyun from '../../../../public/images/ks-yun.svg';
import MOS from '../../../../public/images/mos.svg';
import OracleCloud from '../../../../public/images/oracle-cloud.svg';
import SalesForce from '../../../../public/images/sales-force.svg';
import TencentCloud from '../../../../public/images/tencent-cloud.svg';
import UCloud from '../../../../public/images/u-cloud.svg';

/**
 * 云服务商字段、国际化、icon、class 映射
 */

export type VendorMap = Record<any, any>;

export const vendorMap: VendorMap = {
  AWS: {
    label: 'topoManager.workArea.vendor.AWS',
    icon: AWS,
    class: 'bg-[#DAE9FD]',
  },
  'Tencent Cloud': {
    label: 'topoManager.workArea.vendor.Tencent',
    icon: TencentCloud,
    class: 'bg-[#DAE9FD]',
  },
  'Google Cloud': {
    label: 'topoManager.workArea.vendor.Google',
    icon: GoogleCloud,
    class: 'bg-[#DAF5C8]',
  },
  Azure: {
    label: 'topoManager.workArea.vendor.Azure',
    icon: Azure,
    class: 'bg-[#D8F4F5]',
  },
  'Enterprise Private Cloud': {
    label: 'topoManager.workArea.vendor.Enterprise',
    icon: EnterprisePrivateCloud,
  },
  SalesForce: {
    label: 'topoManager.workArea.vendor.SalesForce',
    icon: SalesForce,
  },
  'Oracle Cloud': {
    label: 'topoManager.workArea.vendor.Oracle',
    icon: OracleCloud,
  },
  'IBM Cloud': {
    label: 'topoManager.workArea.vendor.IBM',
    icon: IBMCloud,
  },
  'Alibaba Cloud': {
    label: 'topoManager.workArea.vendor.Alibaba',
    icon: AlibabaCloud,
    class: 'bg-[#FFE0BF]',
  },
  ECloud: {
    label: 'topoManager.workArea.vendor.ECloud',
    icon: ECloud,
  },
  UCloud: {
    label: 'topoManager.workArea.vendor.UCloud',
    icon: UCloud,
  },
  MOS: {
    label: 'topoManager.workArea.vendor.MOS',
    icon: MOS,
  },
  KSyun: {
    label: 'topoManager.workArea.vendor.KS',
    icon: KSyun,
  },
  'Baidu Cloud': {
    label: 'topoManager.workArea.vendor.Baidu',
    icon: BaiduCloud,
  },
  'Huawei Cloud': {
    label: 'topoManager.workArea.vendor.Huawei',
    icon: HuaweiCloud,
    class: 'bg-[#FFDDDD]',
  },
  'Capital Cloud': {
    label: 'topoManager.workArea.vendor.Capital',
    icon: CapitalCloud,
  },
  'Tencent Private Cloud': {
    label: 'topoManager.workArea.vendor.TencentPrivate',
    icon: defaultCloud,
  },
  Zenlayer: {
    label: 'topoManager.workArea.vendor.Zenlayer',
    icon: defaultCloud,
  },
};
