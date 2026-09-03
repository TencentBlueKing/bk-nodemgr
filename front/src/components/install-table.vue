<template>
  <div ref="contentRef">
    <VxeTable
      ref="xTableRef"
      :data="tableData"
      :size="settings.size"
      :border="true"
      :max-height="maxHeight"
      :scroll-y="{ enabled: true, gt: 20 }"
      :row-config="{ isHover: true, useKey: true }"
      :edit-config="{ trigger: 'click', mode: 'row', showIcon: false }"
      round
    >
      <!-- 业务属性 -->
      <VxeColgroup :title="$t('components.installTable.bizProperty')" align="center" v-if="isReinstall">
        <VxeColumn
          field="bk_biz_id"
          :title="$t('components.installTable.bkBizId')"
          :visible="settings.checked.includes('bk_biz_id')"
          :min-width="150"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.bkBizId') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.installBusinessTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_biz_id')">
              <!-- 归属业务始终不可编辑 -->
              <div :class="['cell-disabled', { 'cell-disabled--error': getError(rowIndex, 'bk_biz_id') }]">
                {{ row.bk_biz_id ? `[${row.bk_biz_id}] ${businessMap[row.bk_biz_id] || row.bk_biz_id}` : '' }}
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_biz_id')">
              <Select
                v-model="row.bk_biz_id"
                auto-focus
                filterable
                transfer
                :disabled="true"
                @change="clearError(rowIndex, 'bk_biz_id')"
                @toggle="(val: any) => !val && handleFieldBlur(rowIndex, 'bk_biz_id', row.bk_biz_id)"
              >
                <Select.Option
                  v-for="item in businessList"
                  :key="item.bk_biz_id"
                  :name="item.bk_biz_name"
                  :id="item.bk_biz_id"
                >
                  [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
                </Select.Option>
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 拓扑属性 -->
      <VxeColgroup :title="$t('components.installTable.topoProperty')" align="center" v-if="isReinstall">
        <VxeColumn
          field="bk_networkarea_name"
          :title="$t('components.installTable.networkArea')"
          :visible="settings.checked.includes('bk_networkarea_name')"
          :min-width="150"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="320"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.networkArea') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.cloudTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkarea_name')">
              <!-- 管控区域始终不可编辑 -->
              <div :class="['cell-disabled', { 'cell-disabled--error': getError(rowIndex, 'bk_networkarea_name') }]">{{ row.bk_networkarea_name }}</div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkarea_name')">
              <Input
                v-model.trim="row.bk_networkarea_name"
                :disabled="true"
                @change="(val: any) => { handleChangeIPv4(val, row, rowIndex); clearError(rowIndex, 'bk_networkarea_name'); }"
                @blur="handleFieldBlur(rowIndex, 'bk_networkarea_name', row.bk_networkarea_name)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="bk_networkunit_id"
          :title="$t('components.installTable.networkUnit')"
          :min-width="150"
          :visible="settings.checked.includes('bk_networkunit_id')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="320"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="mr-[5px] cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.networkUnit') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('platform.nodeMan.installAgentPage.cloudUnitTooltip') }}</p>
                </div>
              </template>
            </Popover>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              v-if="isReinstall"
              :title="$t('components.installTable.batchEditNetworkUnit')"
              type="select"
              :options="networkUnitBatchOptions"
              :disabled="!isSameNetworkArea || networkUnitLoading"
              :disabled-tip="$t('components.installTable.batchEditNetworkUnitDisabledTip')"
              @confirm="(value) => handleBatchEdit('bk_networkunit_id', value)"
            />
            <i
              v-if="releaseType !== 'proxy' && (isReinstall || isUpgrade) && !networkUnitLoading && !autoAssignLoading"
              class="nodeman-icon nc-manual text-[18px] cursor-pointer ml-[5px]"
              v-bk-tooltips="$t('components.installTable.autoAssignTooltip')"
              @click="handleAutoAssign"
            ></i>
            <i
              v-else-if="releaseType !== 'proxy' && (isReinstall || isUpgrade)"
              class="nodeman-icon nc-manual text-[18px] cursor-not-allowed text-[#C4C6CC] ml-[5px]"
              v-bk-tooltips="$t('components.installTable.autoAssignTooltip')"
            ></i>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkunit_id')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'bk_networkunit_id') }">
                <span v-if="getNetworkUnitName(row.bk_networkunit_id)">{{ getNetworkUnitName(row.bk_networkunit_id) }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.selectPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkunit_id')">
              <Select
                v-if="!networkUnitLoading"
                v-model="row.bk_networkunit_id"
                auto-focus
                filterable
                transfer
                @change="(val: any) => { clearError(rowIndex, 'bk_networkunit_id'); handleNetworkUnitChange(val, row, rowIndex); }"
                @toggle="(val: any) => !val && handleFieldBlur(rowIndex, 'bk_networkunit_id', row.bk_networkunit_id)"
              >
                <Select.Option
                  v-for="option in getNetworkUnitsByAreaId(row.bk_networkarea_id)"
                  :key="option.bk_networkunit_id"
                  :id="String(option.bk_networkunit_id)"
                  :name="option.bk_networkunit_name"
                  :disabled="releaseType === 'proxy' && option.is_direct"
                  :class="{ 'unauthorized-unit-row': !isUnitAuthorized(option.bk_networkunit_id) }"
                  @click="handleUnitOptionClick($event, option.bk_networkunit_id)"
                  @mouseenter="handleUnitOptionMouseEnter($event, option.bk_networkunit_id)"
                  @mousemove="handleUnitOptionMouseMove($event, option.bk_networkunit_id)"
                  @mouseleave="handleUnitOptionMouseLeave()"
                  v-bk-tooltips="{
                    content: (releaseType === 'proxy' && option.is_direct)
                      ? $t('topoManager.installProxy.form.tip')
                      : `[${option.bk_networkunit_id}] ${option.bk_networkunit_name}`,
                    disabled: (releaseType === 'proxy' && option.is_direct)
                      ? false
                      : !textOverflowMap[option.bk_networkunit_id],
                    boundary: 'parent',
                    placement: 'right',
                  }"
                >
                  <div class="truncate" @mouseenter="handleTextMouseenter($event, option)">
                    [{{ option.bk_networkunit_id }}] {{ option.bk_networkunit_name }}
                  </div>
                </Select.Option>
              </Select>
              <div v-else class="h-[32px] w-full rounded-[2px] bg-[#F5F7FA]"></div>
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          v-if="false"
          field="install_origin"
          :title="$t('installProxy.installSource')"
          :min-width="150"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #default="{ row }">
            <span v-if="row.install_origin === 'current'">{{ $t('installProxy.currentUnit') }}</span>
            <span v-else-if="row.install_origin === 'upstream'">{{ $t('installProxy.upstreamUnit') }}</span>
            <span v-else-if="row.install_origin && String(row.install_origin).startsWith('custom:')">
              {{ getInstallOriginDisplayName(row.install_origin) }}
            </span>
            <span v-else class="cell-placeholder">{{ $t('components.installTable.selectPlaceholder') }}</span>
          </template>
          <template #edit="{ row, rowIndex }">
            <Cascader
              :model-value="parseInstallOriginValue(row.install_origin)"
              @update:model-value="(val: any) => { handleInstallOriginChange(val, row, rowIndex); }"
              :list="installOriginList"
              trigger="click"
              transfer
            />
          </template>
        </VxeColumn>
      </VxeColgroup>
      <VxeColgroup align="center">
        <template #header>
          <span class="mr-[5px]">{{ $t('components.installTable.hostIp') }}</span>
          <span class="mx-[3px] text-[#FF5656]">*</span>
        </template>
        <VxeColumn
          field="bk_host_innerip"
          :title="$t('components.installTable.innerIPv4')"
          :visible="settings.checked.includes('bk_host_innerip')"
          :min-width="150"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              v-if="releaseType === 'proxy'"
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="300"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.innerIPv4') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.proxyInnerIPTooltip') }}</p>
                </div>
              </template>
            </Popover>
            <span v-else>{{ $t('components.installTable.innerIPv4') }}</span>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_host_innerip')">
              <!-- 重装时内网IPv4不可编辑 -->
              <div :class="[isReinstall ? 'cell-disabled' : '', { 'cell-disabled--error': getError(rowIndex, 'bk_host_innerip') }]">
                <span v-if="isReinstall || row.bk_host_innerip">{{ row.bk_host_innerip }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_host_innerip')">
              <Input
                v-model.trim="row.bk_host_innerip"
                :disabled="isReinstall"
                @change="(val: any) => { handleChangeIPv4(val, row, rowIndex); clearError(rowIndex, 'bk_host_innerip'); }"
                @blur="handleFieldBlur(rowIndex, 'bk_host_innerip', row.bk_host_innerip)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>
        <VxeColumn
          field="bk_host_innerip_v6"
          :title="$t('components.installTable.innerIPv6')"
          :min-width="150"
          :visible="settings.checked.includes('bk_host_innerip_v6')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              v-if="releaseType === 'proxy'"
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="300"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.innerIPv6') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.proxyInnerIPTooltip') }}</p>
                </div>
              </template>
            </Popover>
            <span v-else>{{ $t('components.installTable.innerIPv6') }}</span>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_host_innerip_v6')">
              <div :class="[isReinstall ? 'cell-disabled' : '', { 'cell-disabled--error': getError(rowIndex, 'bk_host_innerip_v6') }]">
                <span v-if="isReinstall || row.bk_host_innerip_v6">{{ row.bk_host_innerip_v6 }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_host_innerip_v6')">
              <Input
                :disabled="isReinstall"
                v-model.trim="row.bk_host_innerip_v6"
                @change="clearError(rowIndex, 'bk_host_innerip_v6')"
                @blur="handleFieldBlur(rowIndex, 'bk_host_innerip_v6', row.bk_host_innerip_v6)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 主机属性 -->
      <VxeColgroup :title="$t('components.installTable.hostAttr')" align="center">
        <VxeColumn
          field="export_ip"
          :min-width="150"
          :visible="settings.checked.includes('export_ip')"
          v-if="releaseType === 'proxy'"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="300"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="mr-[5px] cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.exportIP') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.exportIPTooltip') }}</p>
                </div>
              </template>
            </Popover>
            <span class="mx-[3px] text-[#FF5656]">*</span>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'export_ip')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'export_ip') }">
                <span v-if="row.export_ip">{{ row.export_ip }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'export_ip')">
              <Input
                v-model.trim="row.export_ip"
                @change="clearError(rowIndex, 'export_ip')"
                @blur="handleFieldBlur(rowIndex, 'export_ip', row.export_ip)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="bk_addressing"
          :title="$t('components.installTable.addressingMode')"
          :min-width="120"
          :visible="settings.checked.includes('bk_addressing')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <span class="mr-[5px]">{{ $t('components.installTable.addressingMode') }}</span>
            <BatchEdit
              v-if="!isReinstall && !isUpgrade"
              :title="$t('components.installTable.addressingMode')"
              type="select"
              :options="addressingOptions"
              @confirm="(value) => handleBatchEdit('bk_addressing', value)"
            />
          </template>
          <template #default="{ row, rowIndex }">
            <!-- 重装或升级时寻址方式不可编辑 -->
            <div :class="[(isReinstall || isUpgrade) ? 'cell-disabled' : '', { 'cell-disabled--error': getError(rowIndex, 'bk_addressing') }]">
              <span v-if="(isReinstall || isUpgrade) || row.bk_addressing">{{ addressingOptions.find(item => item.id === row.bk_addressing)?.name || row.bk_addressing }}</span>
              <span v-else class="cell-placeholder">{{ $t('components.installTable.selectPlaceholder') }}</span>
            </div>
          </template>
          <template #edit="{ row }">
            <Select
              v-model="row.bk_addressing"
              auto-focus
              transfer
              :disabled="isReinstall || isUpgrade"
            >
              <Select.Option
                v-for="option in addressingOptions"
                :key="option.id"
                :id="option.id"
                :name="option.name"
              />
            </Select>
          </template>
        </VxeColumn>

        <VxeColumn
          field="os_type"
          :title="$t('components.installTable.osType')"
          :min-width="120"
          :visible="settings.checked.includes('os_type')"
          v-if="releaseType !== 'proxy' || isReinstall || settings.checked.includes('os_type')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <span class="mr-[5px]">{{ $t('components.installTable.osType') }}</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              v-if="!isUpgrade"
              :title="$t('components.installTable.batchEditOsType')"
              type="select"
              :options="datasourceList"
              @confirm="(value) => handleBatchEdit('os_type', value)"
            />
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'os_type')">
              <!-- 升级时操作系统不可编辑 -->
              <div :class="[isUpgrade ? 'cell-disabled' : '', { 'cell-disabled--error': getError(rowIndex, 'os_type') }]">
                <span v-if="isUpgrade || row.os_type">{{ row.os_type }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.selectPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <Input v-if="isUpgrade" v-model="row.os_type" :disabled="true" />
            <ValidateCell v-else :error="getError(rowIndex, 'os_type')">
              <Select
                v-model="row.os_type"
                auto-focus
                transfer
                @change="(val: any) => { handleChangeOsType(val, row, rowIndex); clearError(rowIndex, 'os_type'); }"
                @toggle="(val: any) => !val && handleFieldBlur(rowIndex, 'os_type', row.os_type)"
              >
                <Select.Option
                  v-for="option in datasourceList"
                  :key="option.id"
                  :id="option.id"
                  :name="option.name"
                />
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="cpu_arch"
          :min-width="120"
          :visible="settings.checked.includes('cpu_arch')"
          v-if="(releaseType === 'proxy' && type === 'offline') || isUpgrade"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <span class="mr-[5px]">{{ $t('components.installTable.cpuArch') }}</span>
            <span v-if="!isUpgrade" class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              v-if="!isUpgrade"
              :title="$t('components.installTable.batchEditCpuArch')"
              type="select"
              :options="cpuArchOptions"
              @confirm="(value) => handleBatchEdit('cpu_arch', value)"
            />
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'cpu_arch')">
              <!-- 升级时CPU架构不可编辑 -->
              <div :class="[isUpgrade ? 'cell-disabled' : '', { 'cell-disabled--error': getError(rowIndex, 'cpu_arch') }]">
                <span v-if="isUpgrade || row.cpu_arch">{{ row.cpu_arch }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.selectPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <Input v-if="isUpgrade" v-model="row.cpu_arch" :disabled="true" />
            <ValidateCell v-else :error="getError(rowIndex, 'cpu_arch')">
              <Select
                v-model="row.cpu_arch"
                auto-focus
                transfer
                @change="clearError(rowIndex, 'cpu_arch')"
                @toggle="(val: any) => !val && handleFieldBlur(rowIndex, 'cpu_arch', row.cpu_arch)"
              >
                <Select.Option
                  v-for="option in cpuArchOptions"
                  :key="option.id"
                  :id="option.id"
                  :name="option.name"
                />
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="advertise_ip"
          :min-width="150"
          v-if="releaseType === 'proxy'"
          :visible="settings.checked.includes('advertise_ip')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="300"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="mr-[5px] cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.serviceIP') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.serviceIPTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'advertise_ip')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'advertise_ip') }">
                <span v-if="row.advertise_ip">{{ row.advertise_ip }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'advertise_ip')">
              <Input
                v-model.trim="row.advertise_ip"
                @change="clearError(rowIndex, 'advertise_ip')"
                @blur="handleFieldBlur(rowIndex, 'advertise_ip', row.advertise_ip)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 登录信息 -->
      <VxeColgroup
        :title="$t('components.installTable.loginInfo')"
        align="center"
        v-if="type !== 'manual' && type !== 'offline'"
      >
        <VxeColumn
          field="login_ip"
          :title="$t('components.installTable.loginIP')"
          :min-width="150"
          :visible="settings.checked.includes('login_ip')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="mr-[5px] cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.loginIP') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.loginIPTooltipDesc') }}</p>
                  <p class="mt-[8px]">{{ $t('components.installTable.loginIPTooltipSupport') }}</p>
                </div>
              </template>
            </Popover>
            <span class="mx-[3px] text-[#FF5656]">*</span>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_ip')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'login_ip') }">
                <span v-if="row.login_ip">{{ row.login_ip }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_ip')">
              <Input
                v-model.trim="row.login_ip"
                @change="clearError(rowIndex, 'login_ip')"
                @blur="handleFieldBlur(rowIndex, 'login_ip', row.login_ip)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="login_port"
          :title="$t('components.installTable.port')"
          :min-width="120"
          :visible="settings.checked.includes('login_port')"
          v-if="releaseType !== 'proxy' || isReinstall || settings.checked.includes('login_port')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="240"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="mr-[5px] cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.port') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.portTooltipDesc') }}</p>
                  <div class="mt-[8px]">
                    <p>
                      {{ $t('components.installTable.portTooltipLinuxLabel') }}
                      <span class="text-[#FF9C01]">{{ $t('components.installTable.portTooltipLinuxPort') }}</span>
                    </p>
                    <p class="mt-[2px]">
                      {{ $t('components.installTable.portTooltipWindowsLabel') }}
                      <span class="text-[#FF9C01]">{{ $t('components.installTable.portTooltipWindowsPort') }}</span>
                    </p>
                  </div>
                  <p
                    class="text-[#3A84FF] mt-[8px] cursor-pointer"
                    @click="handlePortBatchApplyDefault"
                  >{{ $t('components.installTable.portBatchApply') }}</p>
                </div>
              </template>
            </Popover>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              :title="$t('components.installTable.batchEditPort')"
              type="input"
              @confirm="(value) => handleBatchEdit('login_port', value)"
            />
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_port')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'login_port') }">
                <span v-if="row.login_port !== '' && row.login_port != null">{{ row.login_port }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_port')">
              <Input
                v-model.trim="row.login_port"
                @change="() => clearError(rowIndex, 'login_port')"
                @blur="handleFieldBlur(rowIndex, 'login_port', row.login_port)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="login_user"
          :title="$t('components.installTable.account')"
          :min-width="150"
          :visible="settings.checked.includes('login_user')"
          v-if="releaseType !== 'proxy' || isReinstall || settings.checked.includes('login_user')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <span class="mr-[5px]">{{ $t('components.installTable.account') }}</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              :title="$t('components.installTable.batchEditAccount')"
              type="input"
              @confirm="(value) => handleBatchEdit('login_user', value)"
            />
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_user')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'login_user') }">
                <span v-if="row.login_user">{{ row.login_user }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_user')">
              <Input
                v-model.trim="row.login_user"
                @change="clearError(rowIndex, 'login_user')"
                @blur="handleFieldBlur(rowIndex, 'login_user', row.login_user)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="login_mode"
          :min-width="120"
          :visible="settings.checked.includes('login_mode')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <span class="mr-[5px]">{{ $t('components.installTable.authMethod') }}</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              :title="$t('components.installTable.batchEditAuthMethod')"
              type="select"
              :options="authenticationTypes"
              @confirm="(value) => handleBatchEdit('login_mode', value)"
            />
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_mode')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'login_mode') }">
                <span v-if="row.login_mode">{{ authenticationTypes.find(item => item.id === row.login_mode)?.name || row.login_mode }}</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.selectPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'login_mode')">
              <Select
                v-model="row.login_mode"
                auto-focus
                transfer
                @change="(val: any) => { handleChangeMode(val, row, rowIndex); clearError(rowIndex, 'login_mode'); }"
                @toggle="(val: any) => !val && handleFieldBlur(rowIndex, 'login_mode', row.login_mode)"
              >
                <Select.Option
                  v-for="option in authenticationTypes"
                  :key="option.id"
                  :id="option.id"
                  :name="option.name"
                />
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="credit"
          :min-width="200"
          :visible="settings.checked.includes('credit')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <div class="flex">
              <span class="mr-[5px]">{{ $t('components.installTable.passwordKey') }}</span>
              <span class="mx-[3px] text-[#FF5656]">*</span>
              <BatchEdit
                :title="$t('components.installTable.batchEditPasswordKey')"
                type="credit"
                @confirm="(value) => handleBatchEdit('credit', value)"
              />
            </div>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'credit')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'credit') }">
                <span v-if="row.login_mode === 'password_vault'">{{ $t('components.installTable.autoGet') }}</span>
                <span v-else-if="row.login_mode === 'keyfile' && row.credit">{{ $t('components.installTable.creditValid') }}</span>
                <span v-else-if="row.login_mode === 'keyfile' && !row.credit" class="cell-placeholder">{{ $t('components.installTable.uploadKeyfile') }}</span>
                <span v-else-if="row.login_credit_valid">{{ $t('components.installTable.creditValid') }}</span>
                <span v-else-if="row.credit">******</span>
                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPassword') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <Input
              v-if="row.login_mode === 'password_vault'"
              :value="$t('components.installTable.autoGet')"
              disabled
            />
            <ValidateCell v-else :error="getError(rowIndex, 'credit')">
              <Upload
                ref="uploader"
                type="formdata"
                v-if="row.login_mode === 'keyfile'"
                :url="url"
                :size="100"
                :multiple="false"
                :limit="1"
                theme="button"
                :before-upload="(val) => handleBeforeUpload(val, row)"
                :custom-request="() => {}"
                @change="clearError(rowIndex, 'credit')"
              />
              <Input
                v-else
                v-model.trim="row.credit"
                :placeholder="row.login_credit_valid
                  ? $t('components.installTable.creditValid')
                  : $t('components.installTable.inputPassword')"
                type="password"
                @change="clearError(rowIndex, 'credit')"
                @blur="handleFieldBlur(rowIndex, 'credit', row.credit)"
              />
            </ValidateCell>
          </template>
        </VxeColumn>

        <VxeColumn
          field="install_method"
          :min-width="120"
          :visible="settings.checked.includes('install_method')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <span>{{ $t('components.installTable.installMethod') }}</span>
          </template>
          <template #default="{ row }">
            <span>{{ installMethodMap[row.install_method] }}</span>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'install_method')">
              <Select
                v-model="row.install_method"
                clearable
                transfer
                @change="(val: any) => handleInstallMethodChange(val, row, rowIndex)"
              >
                <Select.Option
                  v-for="option in getInstallMethodOptions(row.os_type)"
                  :key="option.id"
                  :id="option.id"
                  :name="option.name"
                />
              </Select>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 安装设置 -->
      <VxeColgroup :title="$t('components.installTable.installSettings')" align="center">
        <VxeColumn
          :min-width="120"
          field="install_pre_ordered_plugins"
          :visible="settings.checked.includes('install_pre_ordered_plugins')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.installPreOrderedPlugins') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.installPreOrderedPluginsTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.install_pre_ordered_plugins" size="small" />
          </template>
          <template #edit="{ row }">
            <Switcher theme="primary" v-model="row.install_pre_ordered_plugins" />
          </template>
        </VxeColumn>

        <VxeColumn
          :min-width="150"
          field="re_register"
          :visible="settings.checked.includes('re_register')"
          :edit-render="{ name: 'VxeInput' }"
        >
          <template #header>
            <Popover
              theme="light"
              trigger="hover"
              placement="top"
              :arrow="true"
              :max-width="280"
              :offset="8"
              :popover-delay="[0, 100]"
              :component-event-delay="0"
            >
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                {{ $t('components.installTable.reRegisterAgentId') }}
              </span>
              <template #content>
                <div class="text-[12px] leading-[20px]">
                  <p>{{ $t('components.installTable.reRegisterAgentIdTooltip') }}</p>
                </div>
              </template>
            </Popover>
          </template>
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.re_register" size="small" />
          </template>
          <template #edit="{ row }">
            <Switcher theme="primary" v-model="row.re_register" />
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 开启的服务 -->
      <VxeColgroup
        :title="$t('components.installTable.enabledSerive')"
        align="center"
        v-if="releaseType === 'proxy'"
      >
        <VxeColumn :min-width="90" field="dedicated_installer" :visible="settings.checked.includes('dedicated_installer')">
          <template #header>
            <Popover theme="light" trigger="hover" placement="top" :arrow="true" :max-width="280" :offset="8" :popover-delay="[0, 100]" :component-event-delay="0">
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">{{ $t('components.installTable.installJump') }}</span>
              <template #content><div class="text-[12px] leading-[20px]"><p>{{ $t('components.installTable.installJumpTooltip') }}</p></div></template>
            </Popover>
          </template>
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.dedicated_installer" />
          </template>
        </VxeColumn>

        <VxeColumn :min-width="90" field="cluster_tunnel" :visible="settings.checked.includes('cluster_tunnel')">
          <template #header>
            <Popover theme="light" trigger="hover" placement="top" :arrow="true" :max-width="280" :offset="8" :popover-delay="[0, 100]" :component-event-delay="0">
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">{{ $t('components.installTable.agentControl') }}</span>
              <template #content><div class="text-[12px] leading-[20px]"><p>{{ $t('components.installTable.agentControlTooltip') }}</p></div></template>
            </Popover>
          </template>
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.cluster_tunnel" />
          </template>
        </VxeColumn>

        <VxeColumn :min-width="90" field="file_tunnel" :visible="settings.checked.includes('file_tunnel')">
          <template #header>
            <Popover theme="light" trigger="hover" placement="top" :arrow="true" :max-width="280" :offset="8" :popover-delay="[0, 100]" :component-event-delay="0">
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">{{ $t('components.installTable.fileTransfer') }}</span>
              <template #content><div class="text-[12px] leading-[20px]"><p>{{ $t('components.installTable.fileTransferTooltip') }}</p></div></template>
            </Popover>
          </template>
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.file_tunnel" />
          </template>
        </VxeColumn>

        <VxeColumn :min-width="90" field="data_tunnel" :visible="settings.checked.includes('data_tunnel')">
          <template #header>
            <Popover theme="light" trigger="hover" placement="top" :arrow="true" :max-width="280" :offset="8" :popover-delay="[0, 100]" :component-event-delay="0">
              <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">{{ $t('components.installTable.dataReport') }}</span>
              <template #content><div class="text-[12px] leading-[20px]"><p>{{ $t('components.installTable.dataReportTooltip') }}</p></div></template>
            </Popover>
          </template>
          <template #default="{ row }">
            <Switcher theme="primary" v-model="row.data_tunnel" />
          </template>
        </VxeColumn>
      </VxeColgroup>

      <VxeColgroup>
        <template #header>
          <Button text style="margin-right: 8px">
            <Settings ref="settingRef" :settings="settings" @setting-change="settingChange" />
          </Button>
        </template>
        <VxeColumn :min-width="80" field="action" :title="$t('components.installTable.operate')">
          <template #default="{ rowIndex }">
            <Button
              text
              :disabled="isReinstall"
              @click="handleAddRow(rowIndex)"
            >
              <i class="nodeman-icon nc-plus"></i>
            </Button>
            <Button
              text
              :disabled="isUpgrade || tableData?.length <= 1"
              @click="handleDelRow(rowIndex)"
              style="margin-left: 8px"
            >
              <i class="nodeman-icon nc-minus"></i>
            </Button>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <template #empty>
        <slot></slot>
      </template>
    </VxeTable>
  </div>
</template>

<script lang="ts" setup>
import { Button, Cascader, InfoBox, Input, Message, Popover, Select, Switcher, Upload } from 'bkui-vue';
import { cloneDeep, debounce, groupBy } from 'lodash';
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import ValidateCell from './validateCell.vue';

import type { PackageReleaseDistinctData } from '@/@types/pkg.d';
import { PackageService } from '@/api/modules/pkg';
import { TopoService } from '@/api/modules/topo';
import { PACKAGE_GENERATION, VALIDATE_REGEX } from '@/common/const';
import { getDefaultLoginMode, getFirstIp } from '@/common/util';
import BatchEdit from '@/components/batch-edit.vue';
import useFullScreen from '@/composables/use-fullscreen';
import useTableErrors from '@/composables/use-table-errors';
import useUnitAuth from '@/composables/use-unit-auth';
import { useAuthStore } from '@/stores/auth';
import { useMainStore } from '@/stores/main';

type VxeComponentSizeType = 'small' | 'medium' | 'large';
interface IValidate {
  validator: Function | RegExp | string;
  message: string;
}
type ValidationRules = Record<string, IValidate[]>;

const tableData = defineModel<Array<ReturnType<typeof getInitData>>>('data');
const props = defineProps({
  data: { type: Array, default: () => [] as any[] },
  maxHeight: { type: Number, default: 300 },
  releaseType: { type: String, default: 'agent' },
  isReinstall: { type: Boolean, default: false },
  isUpgrade: { type: Boolean, default: false },
  installOriginList: { type: Array, default: () => [] },
  currentSettings: {
    type: Object,
    default: () => ({
      fields: [
        { title: '内网 IPv4', field: 'bk_host_innerip' },
        { title: '内网 IPv6', field: 'bk_host_innerip_v6' },
        { title: '寻址方式', field: 'bk_addressing' },
        { title: '操作系统', field: 'os_type' },
        { title: '登录 IP', field: 'login_ip' },
        { title: '登录端口', field: 'login_port' },
        { title: '登录账号', field: 'login_user' },
        { title: '认证方式', field: 'login_mode' },
        { title: '密码 / 密钥', field: 'credit' },
        { title: '安装预设插件', field: 'install_pre_ordered_plugins' },
        { title: '重新注册AgentID', field: 're_register' },
        { title: '安装方式', field: 'install_method' },
      ],
      checked: [
        'bk_host_innerip',
        'bk_host_innerip_v6',
        'bk_addressing',
        'os_type',
        'login_port',
        'login_ip',
        'login_user',
        'login_mode',
        'credit',
      ],
      disabled: ['os_type', 'login_port', 'login_user', 'login_mode', 'credit'],
      size: 'medium' as VxeComponentSizeType,
    }),
  },
});

const { t } = useI18n();

// ----------------------------------------------------------------
// 问题一：Select transfer 下拉选择后 vxe 误判点击在行外导致退出编辑
// 在 document 捕获阶段拦截 mousedown：若点击目标在 .bk-select-dropdown 内，
// 调用 stopImmediatePropagation 阻止 VxeTable 的全局 mousedown 处理器
// 关闭行编辑态。click 事件不受影响，Select 选项仍可正常选中。
// ----------------------------------------------------------------
const handleDocMouseDown = (e: MouseEvent) => {
  const target = e.target as HTMLElement;
  if (target?.closest?.('.bk-select-dropdown')) {
    e.stopImmediatePropagation();
  }
};

// 根据操作系统 + 安装方式计算默认端口（Windows 仅 WMI 用 WMI 端口，SSH/自动用 SSH 端口）
const getDefaultPort = (osType: string, installMethod: string) => {
  if (osType === 'windows') {
    return installMethod === 'wmi'
      ? window.PROJECT_CONFIG.WINDOWS_WMI_PORT_DEFAULT
      : window.PROJECT_CONFIG.WINDOWS_SSH_PORT_DEFAULT;
  }
  return window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT;
};

const handlePortBatchApplyDefault = () => {
  tableData.value?.forEach((item: any, index: number) => {
    const port = getDefaultPort(item.os_type, item.install_method);
    item.login_port = port;
    handleFieldBlur(index, 'login_port', port);
  });
};

const initData = {
  bk_host_innerip: '',
  bk_host_innerip_v6: '',
  os_type: '',
  cpu_arch: '',
  login_ip: '',
  login_port: '',
  login_user: '',
  login_mode: getDefaultLoginMode(),
  login_password: '',
  login_key_file: '',
  login_credit_valid: false,
  bk_addressing: 'static',
  bk_networkunit_id: '',
  bk_biz_id: '',
  bk_host_id: '',
  re_register: false,
  install_pre_ordered_plugins: true,
  credit: '',
  export_ip: '',
  advertise_ip: '',
  dedicated_installer: true,
  cluster_tunnel: true,
  file_tunnel: true,
  data_tunnel: true,
  proxy_tags: [],
  install_method: 'ssh',
};

// 安装方式展示映射
const installMethodMap = computed<Record<string, string>>(() => ({
  ssh: t('components.installTable.installMethodSSH'),
  wmi: t('components.installTable.installMethodWMI'),
  auto: t('components.installTable.installMethodAuto'),
}));

// 根据操作系统返回可选的安装方式（Windows 有 WMI，Linux 没有）
const getInstallMethodOptions = (osType: string) => {
  if (osType === 'windows') {
    return [
      { id: 'ssh', name: t('components.installTable.installMethodSSH') },
      { id: 'wmi', name: t('components.installTable.installMethodWMI') },
      { id: 'auto', name: t('components.installTable.installMethodAuto') },
    ];
  }
  return [
    { id: 'ssh', name: t('components.installTable.installMethodSSH') },
    { id: 'auto', name: t('components.installTable.installMethodAuto') },
  ];
};

const rules: ValidationRules = {
  bk_host_innerip: [
    { validator: VALIDATE_REGEX.IPV4, message: t('components.installTable.ipv4ValidMessage') },
  ],
  bk_host_innerip_v6: [
    { validator: VALIDATE_REGEX.IPV6, message: t('components.installTable.ipv6ValidMessage') },
  ],
  os_type: [{ validator: (val: string) => val, message: t('components.installTable.osTypeValidMessage') }],
  login_ip: [
    { validator: VALIDATE_REGEX.IPV4, message: t('components.installTable.loginIPValidMessage') },
  ],
  login_port: [
    { validator: VALIDATE_REGEX.PORT, message: t('components.installTable.portValidMessage') },
  ],
  login_user: [{ validator: (val: string) => val, message: t('components.installTable.accountValidMessage') }],
  login_mode: [{ validator: (val: string) => val, message: t('components.installTable.loginModeValidMessage') }],
  credit: [{ validator: (val: string) => val, message: t('components.installTable.passwordKeyValidMessage') }],
};

const { contentRef } = useFullScreen();
const xTableRef = ref();
const mainStore = useMainStore();
const authStore = useAuthStore();
const textOverflowMap = reactive<Record<number, boolean>>({});
const handleTextMouseenter = (e: MouseEvent, option: any) => {
  const el = e.target as HTMLElement;
  textOverflowMap[option.bk_networkunit_id] = el.scrollWidth > el.clientWidth;
};
const {
  isUnitAuthorized,
  handleOptionMouseEnter: handleUnitOptionMouseEnter,
  handleOptionMouseMove: handleUnitOptionMouseMove,
  handleOptionMouseLeave: handleUnitOptionMouseLeave,
  handleOptionClick: handleUnitOptionClick,
} = useUnitAuth();
const businessList = computed(() => mainStore.businessList);
const businessMap = computed(() => {
  const map: Record<string, string> = {};
  businessList.value.forEach((item: any) => {
    map[item.bk_biz_id] = item.bk_biz_name;
  });
  return map;
});
const type = computed(() => (props.releaseType === 'proxy'
  ? mainStore.proxySetupType
  : mainStore.agentSetupType));

const { getError, setError, clearError, clearAllErrors, shiftErrors } = useTableErrors();

function getInitData() {
  return cloneDeep(initData);
}

// ----------------------------------------------------------------
// 问题二：加号新增行自动进入编辑状态；加号去掉 isReinstall 限制由业务决定
// 减号简化 disabled 条件
// ----------------------------------------------------------------
const handleAddRow = (index: number) => {
  if (!Array.isArray(tableData.value)) return;
  const newRow = cloneDeep(initData);
  if (props.releaseType === 'proxy' && !props.isReinstall) {
    if (!newRow.os_type) newRow.os_type = 'linux';
    if (!newRow.login_port) newRow.login_port = window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT;
    if (!newRow.login_user) newRow.login_user = 'root';
  }
  tableData.value.splice(index + 1, 0, newRow);
  shiftErrors(index, 1);
  // 新增行自动进入编辑状态
  nextTick(() => {
    xTableRef.value?.setEditRow(tableData.value![index + 1]);
  });
};

const doDelRow = (index: number) => {
  if (!Array.isArray(tableData.value)) return;
  if (tableData.value.length === 1) return Message({ theme: 'warning', message: t('components.installTable.minimum') });
  tableData.value.splice(index, 1);
  shiftErrors(index, -1);
};

const handleDelRow = (index: number) => {
  if (props.isReinstall) {
    const row = tableData.value?.[index];
    const ip = row?.bk_host_innerip || row?.bk_host_innerip_v6 || '';
    InfoBox({
      title: t('components.installTable.confirmDeleteRow'),
      subTitle: t('components.installTable.confirmDeleteRowSub', {
        ip,
        action: props.isUpgrade
          ? t('platform.nodeMan.agentStatus.upgrade')
          : t('platform.nodeMan.agentStatus.reinstall'),
      }),
      onConfirm: () => doDelRow(index),
    });
    return;
  }
  doDelRow(index);
};

const settings = reactive(cloneDeep(props.currentSettings));
const settingChange = (data: any) => {
  settings.checked = data.checked;
  settings.size = data.size;
};

const datasourceList = ref<{ id: string; name: string }[]>([]);
const cpuArchOptions = ref<{ id: string; name: string }[]>([]);
const addressingOptions = ref([
  { id: 'static', name: t('components.installTable.addressingStatic') },
  { id: 'dynamic', name: t('components.installTable.addressingDynamic') },
]);
const authenticationTypes = ref([
  { id: 'password', name: t('components.installTable.password') },
  { id: 'keyfile', name: t('components.installTable.keyfile') },
  ...(window.PROJECT_CONFIG.PASSWORD_VAULT_SWITCH === 'true'
    ? [{ id: 'password_vault', name: window.PROJECT_CONFIG.PASSWORD_VAULT_NAME }]
    : []),
]);
const hostDistinct = ref<PackageReleaseDistinctData | null>();

// --- 业务逻辑 ---
const handleChangeMode = (val: string, row: any, rowIndex: number) => {
  row.login_mode = val;
  row.credit = '';
  clearError(rowIndex, 'credit');
};

const handleInstallMethodChange = (val: string, row: any, rowIndex: number) => {
  if (!val) return;
  if (row.os_type === 'windows') {
    // Windows：WMI 用 WMI 端口，SSH/自动用 SSH 端口
    row.login_port = val === 'wmi'
      ? window.PROJECT_CONFIG.WINDOWS_WMI_PORT_DEFAULT
      : window.PROJECT_CONFIG.WINDOWS_SSH_PORT_DEFAULT;
  } else {
    // Linux：SSH/自动都用 SSH 端口
    row.login_port = window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT;
  }
  handleFieldBlur(rowIndex, 'login_port', row.login_port);
};

const handleChangeIPv4 = (val: string, row: any, rowIndex: number) => {
  if (new RegExp(VALIDATE_REGEX.IPV4).test(val)) {
    row.login_ip = val;
    handleFieldBlur(rowIndex, 'login_ip', val);
  }
};

const handleChangeOsType = (val: string, row: any, rowIndex: number) => {
  if (val === 'linux') {
    row.login_port = window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT;
    row.login_user = 'root';
    // Linux 无 WMI 安装方式，若之前选了 WMI 则重置为 SSH
    if (row.install_method === 'wmi') {
      row.install_method = 'ssh';
    }
    handleFieldBlur(rowIndex, 'login_port', window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT);
    handleFieldBlur(rowIndex, 'login_user', 'root');
  }
  if (val === 'windows') {
    const port = getDefaultPort('windows', row.install_method);
    row.login_port = port;
    row.login_user = 'administrator';
    handleFieldBlur(rowIndex, 'login_port', port);
    handleFieldBlur(rowIndex, 'login_user', 'administrator');
  }
};

const getHostDistinct = async () => {
  const distinctParams = {
    generation: PACKAGE_GENERATION,
    exact_include_conditions: { enabled: [true] },
    distinct_field: { os_type: true, cpu_arch: true },
  };
  const service = props.releaseType === 'proxy'
    ? PackageService.DistinctReleaseProxy(distinctParams as any)
    : PackageService.DistinctReleaseAgent(distinctParams as any);
  const res = await service.catch(() => null);
  if (res) {
    hostDistinct.value = res;
    datasourceList.value = res.os_type.map(item => ({ id: item, name: item }));
    if (res.cpu_arch?.length) {
      cpuArchOptions.value = res.cpu_arch.map(arch => ({ id: arch, name: arch }));
    }
  }
};

const handleBatchEdit = (field: string, value: any) => {
  if (field === 'bk_networkunit_id' && isProxyDirectNetworkUnit(value)) {
    Message({ theme: 'warning', message: t('topoManager.installProxy.form.tip') });
    return;
  }

  tableData.value?.forEach((item: any, index: number) => {
    if (field === 'credit') {
      if (item.login_mode === 'password') item.credit = value.password;
      else item.credit = value.key;
    } else {
      item[field] = value;
    }

    handleFieldBlur(index, field, value);

    if (field === 'os_type') {
      if (value === 'linux') {
        item.login_port = window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT;
        item.login_user = 'root';
        handleFieldBlur(index, 'login_port', window.PROJECT_CONFIG.UNIX_SSH_PORT_DEFAULT);
        handleFieldBlur(index, 'login_user', 'root');
      } else if (value === 'windows') {
        const port = getDefaultPort('windows', item.install_method);
        item.login_port = port;
        item.login_user = 'administrator';
        handleFieldBlur(index, 'login_port', port);
        handleFieldBlur(index, 'login_user', 'administrator');
      }
    }

    if (field === 'bk_networkunit_id') {
      const networkUnit = networkUnitList.value.find((unit: any) => String(unit.bk_networkunit_id) === value);
      if (networkUnit) item.bk_networkunit_name = networkUnit.bk_networkunit_name;
      clearError(index, 'bk_networkunit_id');
    }
  });
};

const handleAutoAssign = async () => {
  if (!tableData.value?.length) return;
  autoAssignLoading.value = true;
  try {
    const items = tableData.value.map((row: any) => ({
      bk_networkarea_id: Number(row.bk_networkarea_id),
      ip: getFirstIp(row.bk_host_innerip) || getFirstIp(row.bk_host_innerip_v6),
    }));
    const res = await TopoService.RecommendNetworkUnitByNetworkSegment({ items });
    if (res?.items) {
      res.items.forEach((result: any, index: number) => {
        if (index >= tableData.value!.length) return;
        const row = tableData.value![index];
        row.bk_networkunit_id = result.bk_networkunit_id === -1 ? '' : String(result.bk_networkunit_id);
        clearError(index, 'bk_networkunit_id');
      });
    }
  } catch (error) {
    console.error('Auto assign failed:', error);
  } finally {
    autoAssignLoading.value = false;
  }
};

const url = location.href;
const handleBeforeUpload = (file: File, row: any) => {
  const reader = new FileReader();
  reader.onload = (e) => {
    const res = e.target?.result as string;
    if (res) {
      row.credit = res.startsWith('data:') ? res.split(',')[1] : res;
    }
  };
  reader.readAsDataURL(file);
  return true;
};

// --- 数据校验逻辑 ---
const validateItemData = (value: any, rulesArr: IValidate[]) => {
  if (!rulesArr || !rulesArr.length) return true;
  if (!value && value !== 0) return false;
  for (const rule of rulesArr) {
    const { validator } = rule;
    let isValid = true;
    if (typeof validator === 'function') isValid = validator(value);
    else if (validator instanceof RegExp) isValid = validator.test(value);
    else if (typeof validator === 'string') isValid = new RegExp(validator).test(value);
    if (!isValid) return false;
  }
  return true;
};

const handleFieldBlur = (rowIndex: number, field: string, value: any) => {
  const requiredFields = [
    'bk_host_innerip', 'os_type', 'login_ip', 'login_port',
    'login_user', 'login_mode', 'export_ip', 'cpu_arch', 'bk_networkunit_id',
  ];
  if (props.isReinstall) requiredFields.push('bk_biz_id');

  if (requiredFields.includes(field) && !value && value !== 0) {
    setError(rowIndex, field, t('validate.required'));
    return;
  }

  if (field === 'bk_networkunit_id' && value && isProxyDirectNetworkUnit(value)) {
    setError(rowIndex, field, t('topoManager.installProxy.form.tip'));
    return;
  }

  if (field === 'credit') {
    const row = tableData.value![rowIndex];
    if (row.login_mode !== 'password_vault' && !row.login_credit_valid && !value) {
      setError(rowIndex, field, t('validate.required'));
      return;
    }
    // login_credit_valid 为 true 时密码已有效，跳过规则校验
    if (row.login_credit_valid) {
      clearError(rowIndex, field);
      return;
    }
  }

  const fieldRules = rules[field];
  if (fieldRules) {
    if (!validateItemData(value, fieldRules)) {
      setError(rowIndex, field, fieldRules[0].message);
    } else {
      clearError(rowIndex, field);
    }
  } else {
    clearError(rowIndex, field);
  }
};

// --- 全局校验 ---
const tableValidate = async () => {
  const data = tableData.value;
  if (!Array.isArray(data) || !data.length) return true;

  clearAllErrors();
  let isValid = true;
  let firstErrorRowIndex = -1;

  for (let i = 0; i < data.length; i++) {
    const row = data[i];
    let rowValid = true;

    if (props.isReinstall && settings.checked.includes('bk_biz_id') && !row.bk_biz_id) {
      setError(i, 'bk_biz_id', t('validate.required')); rowValid = false;
    }
    if (props.isReinstall && !row.bk_networkunit_id) {
      setError(i, 'bk_networkunit_id', t('validate.required')); rowValid = false;
    }

    const hasIpv4Config = settings.checked.includes('bk_host_innerip');
    const hasIpv6Config = settings.checked.includes('bk_host_innerip_v6');
    const ipv4Value = row.bk_host_innerip;
    const ipv6Value = row.bk_host_innerip_v6;

    if ((hasIpv4Config || hasIpv6Config) && !ipv4Value && !ipv6Value) {
      if (hasIpv4Config) setError(i, 'bk_host_innerip', t('components.installTable.either'));
      if (hasIpv6Config) setError(i, 'bk_host_innerip_v6', t('components.installTable.either'));
      rowValid = false;
    }
    if (hasIpv4Config && ipv4Value && !validateItemData(ipv4Value, rules.bk_host_innerip)) {
      setError(i, 'bk_host_innerip', rules.bk_host_innerip[0].message); rowValid = false;
    }
    if (hasIpv6Config && ipv6Value && !validateItemData(ipv6Value, rules.bk_host_innerip_v6)) {
      setError(i, 'bk_host_innerip_v6', rules.bk_host_innerip_v6[0].message); rowValid = false;
    }
    if (settings.checked.includes('os_type') && !row.os_type) {
      setError(i, 'os_type', t('validate.required')); rowValid = false;
    }

    if (props.releaseType === 'proxy') {
      if (row.bk_networkunit_id && isProxyDirectNetworkUnit(row.bk_networkunit_id)) {
        setError(i, 'bk_networkunit_id', t('topoManager.installProxy.form.tip')); rowValid = false;
      }
      if (settings.checked.includes('export_ip')) {
        if (!row.export_ip) { setError(i, 'export_ip', t('validate.required')); rowValid = false; }
        else if (!validateItemData(row.export_ip, rules.login_ip)) { setError(i, 'export_ip', rules.login_ip[0].message); rowValid = false; }
      }
      if (settings.checked.includes('advertise_ip') && row.advertise_ip) {
        if (!validateItemData(row.advertise_ip, rules.login_ip)) { setError(i, 'advertise_ip', rules.login_ip[0].message); rowValid = false; }
      }
    }

    if (props.releaseType === 'proxy' && type.value === 'offline' && !row.cpu_arch) {
      setError(i, 'cpu_arch', t('validate.required')); rowValid = false;
    }

    if (type.value !== 'manual' && type.value !== 'offline') {
      if (settings.checked.includes('login_ip')) {
        if (!row.login_ip) { setError(i, 'login_ip', t('validate.required')); rowValid = false; }
        else if (!validateItemData(row.login_ip, rules.login_ip)) { setError(i, 'login_ip', rules.login_ip[0].message); rowValid = false; }
      }
      if (settings.checked.includes('login_port')) {
        if (!row.login_port) { setError(i, 'login_port', t('validate.required')); rowValid = false; }
        else if (!validateItemData(row.login_port, rules.login_port)) { setError(i, 'login_port', rules.login_port[0].message); rowValid = false; }
      }
      if (settings.checked.includes('login_user') && !row.login_user) { setError(i, 'login_user', t('validate.required')); rowValid = false; }
      if (settings.checked.includes('login_mode') && !row.login_mode) { setError(i, 'login_mode', t('validate.required')); rowValid = false; }
      if (settings.checked.includes('credit') && row.login_mode !== 'password_vault') {
        if (!row.login_credit_valid && !row.credit) { setError(i, 'credit', t('validate.required')); rowValid = false; }
      }
    }

    if (row.login_mode === 'password_vault' && window.PROJECT_CONFIG.PASSWORD_VAULT_SWITCH !== 'true') {
      setError(i, 'login_mode', t('components.installTable.passwordVaultDisabled')); rowValid = false;
    }

    if (!rowValid) {
      isValid = false;
      if (firstErrorRowIndex === -1) firstErrorRowIndex = i;
    }
  }

  if (!isValid && firstErrorRowIndex !== -1 && xTableRef.value) {
    await xTableRef.value.scrollToRow(data[firstErrorRowIndex]);
  }

  return isValid;
};

// --- 管控单元 ---
const networkUnitList = ref<any[]>([]);
const networkUnitGroupMap = ref<Record<number, any[]>>({});
const networkUnitLoading = ref(false);
const autoAssignLoading = ref(false);

const getNetworkUnitList = async () => {
  networkUnitLoading.value = true;
  try {
    const areaIds = [...new Set(
      tableData.value?.map((item: any) => Number(item.bk_networkarea_id)) || []
    )];
    const res = await TopoService.NetworkUnitListBrief({
      exact_include_conditions: { bk_networkarea_id: areaIds },
    }).catch((err: any) => {
      console.error(err);
      return { total: 0, items: [] };
    });
    networkUnitList.value = res.items;
    networkUnitGroupMap.value = groupBy(res.items, 'bk_networkarea_id');
  } finally {
    networkUnitLoading.value = false;
  }
};

const getNetworkUnitsByAreaId = (bkNetworkAreaId: number | string) => {
  return networkUnitGroupMap.value[Number(bkNetworkAreaId)] || [];
};

const getNetworkUnitName = (unitId: number | string) => {
  const unit = networkUnitList.value.find((item: any) => String(item.bk_networkunit_id) === String(unitId));
  return unit ? unit.bk_networkunit_name : '';
};

const isProxyDirectNetworkUnit = (networkUnitID: number | string) => {
  if (props.releaseType !== 'proxy') return false;
  return !!networkUnitList.value.find((unit: any) =>
    String(unit.bk_networkunit_id) === String(networkUnitID) && unit.is_direct
  );
};

const isSameNetworkArea = computed(() => {
  if (!tableData.value?.length) return false;
  const firstAreaId = Number(tableData.value[0].bk_networkarea_id);
  return tableData.value.every((item: any) => Number(item.bk_networkarea_id) === firstAreaId);
});

const networkUnitBatchOptions = computed(() => {
  if (!isSameNetworkArea.value || !tableData.value?.length) return [];
  const areaId = Number(tableData.value[0].bk_networkarea_id);
  return getNetworkUnitsByAreaId(areaId).map((unit: any) => ({
    id: String(unit.bk_networkunit_id),
    name: `[${unit.bk_networkunit_id}] ${unit.bk_networkunit_name}`,
    disabled: props.releaseType === 'proxy' && unit.is_direct,
    disabledTip: t('topoManager.installProxy.form.tip'),
  }));
});

// 安装源：行内存储为字符串（'current' | 'upstream' | 'custom:123' | ''），cascader 需要/返回数组
const parseInstallOriginValue = (val: any): string[] => {
  if (!val) return [];
  const s = String(val);
  if (s === 'current' || s === 'upstream') return [s];
  if (s.startsWith('custom:')) return ['custom', s.slice(7)];
  return [];
};
const formatInstallOriginValue = (val: any): string => {
  if (!Array.isArray(val) || val.length === 0) return '';
  if (val[0] === 'custom' && val.length > 1) return `custom:${val[1]}`;
  return val[0] || '';
};
const getInstallOriginDisplayName = (val: string): string => {
  if (val.startsWith('custom:')) {
    const unitId = val.slice(7);
    const unit = networkUnitList.value.find((u: any) => String(u.bk_networkunit_id) === unitId);
    return unit ? `${t('installProxy.custom')} [${unit.bk_networkunit_id}] ${unit.bk_networkunit_name}` : t('installProxy.custom');
  }
  return val;
};

// ---- proxy 重装：安装源表内联动（自包含，不依赖父组件） ----
const unitHasProxyMap = ref<Map<number, boolean>>(new Map());
const unitUpstreamMap = ref<Map<number, number | undefined>>(new Map());

// 正在重装的主机 id 集合（这些 proxy 不计入"已有 proxy"）
const getReinstallHostIds = () => new Set(
  (tableData.value || [])
    .map((item: any) => item.bk_host_id)
    .filter((id: any) => id !== '' && id !== null && id !== undefined)
    .map((id: any) => Number(id)),
);

const checkUnitHasProxy = async (unitId: number): Promise<boolean> => {
  if (unitHasProxyMap.value.has(unitId)) return unitHasProxyMap.value.get(unitId) ?? false;
  let hasProxy = false;
  try {
    const res = await TopoService.HostList({
      page: { offset: 0, limit: 500 },
      only_count: false,
      exact_include_conditions: {
        bk_networkunit_id: [unitId],
        node_role: ['proxy'],
        node_status: ['running'],
      },
      fuzzy_include_conditions: {},
    });
    const reinstallHostIds = getReinstallHostIds();
    console.log('[debug] checkUnitHasProxy', {
      unitId,
      reinstallHostIds: [...reinstallHostIds],
      total: res?.total,
      items: (res?.items ?? []).map((i: any) => ({ hostId: i.bk_host_id, unitId: i.info?.bk_networkunit_id, role: i.node_role ?? i.info?.node_role, status: i.node_status ?? i.info?.node_status })),
    });
    const filtered = (res.items ?? []).filter((item: any) => !reinstallHostIds.has(Number(item.bk_host_id)));
    hasProxy = filtered.length > 0;
    console.log('[debug] checkUnitHasProxy filtered', { unitId, filteredCount: filtered.length, hasProxy });
  } catch {
    hasProxy = false;
  }
  unitHasProxyMap.value.set(unitId, hasProxy);
  return hasProxy;
};

const getUnitUpstream = async (unitId: number): Promise<number | undefined> => {
  if (unitUpstreamMap.value.has(unitId)) return unitUpstreamMap.value.get(unitId);
  let upstreamId: number | undefined;
  try {
    const detail: any = await TopoService.NetworkUnitGet({ bk_networkunit_id: unitId });
    upstreamId = detail?.links?.cluster?.bk_networkunit_id;
  } catch {
    upstreamId = undefined;
  }
  unitUpstreamMap.value.set(unitId, upstreamId);
  return upstreamId;
};

const resolveInstallOrigin = async (unitId: number): Promise<string> => {
  if (!unitId) return '';
  const hasProxy = await checkUnitHasProxy(unitId);
  const upstreamId = await getUnitUpstream(unitId);
  return hasProxy ? 'current' : (upstreamId != null ? 'upstream' : '');
};

const handleNetworkUnitChange = async (val: string, row: any, rowIndex: number) => {
  if (!val) return;
  const networkUnit = networkUnitList.value.find((unit: any) => String(unit.bk_networkunit_id) === val);
  const target: any = tableData.value?.[rowIndex] || row;
  target.bk_networkunit_id = val;
  if (networkUnit) target.bk_networkunit_name = networkUnit.bk_networkunit_name;
  // 表内联动：切换单元后按新单元重新计算该行安装源
  if (props.isReinstall && props.releaseType === 'proxy') {
    const origin = await resolveInstallOrigin(Number(val));
    target.install_origin = origin;
    if (row !== target) row.install_origin = origin;
  }
};

// 安装源列编辑：同步回真实数据（row 可能是行编辑克隆对象）
const handleInstallOriginChange = (val: any, row: any, rowIndex: number) => {
  const value = formatInstallOriginValue(val);
  row.install_origin = value;
  const target: any = tableData.value?.[rowIndex];
  if (target && target !== row) {
    target.install_origin = value;
  }
};

const settingRef = ref();
const showSetting = () => settingRef.value?.showSetting();

const autoFillDefaults = () => {
  if (!tableData.value?.length) return;
  const firstIp = (val: string) => val?.split(',')[0] || '';
  tableData.value.forEach((row: any) => {
    if (!row.login_ip) {
      row.login_ip = props.releaseType === 'proxy'
        ? firstIp(row.export_ip) || firstIp(row.bk_host_innerip) || firstIp(row.bk_host_innerip_v6)
        : firstIp(row.bk_host_innerip) || firstIp(row.bk_host_innerip_v6);
    }
    if (row.os_type && (!row.login_port || Number(row.login_port) === 0 || String(row.login_port).includes('{{'))) {
      const defaultPort = getDefaultPort(row.os_type, row.install_method);
      row.login_port = String(defaultPort).includes('{{') ? '' : defaultPort;
    }
    if (row.os_type && !row.login_user) {
      row.login_user = row.os_type === 'windows' ? 'administrator' : 'root';
    }
  });
};

// 监听数组引用变化（如 CMDB 同步数据整体替换 formData.info），
// 补填 login_port 为 0 的默认端口。仅监听 length 时，替换后长度不变（如单选 1 台）不会触发
watch(() => tableData.value, () => {
  autoFillDefaults();
}, { immediate: true });

onMounted(async () => {
  // 必须在 VxeTable 注册全局 mousedown 之前注册，保证捕获阶段先执行
  document.addEventListener('mousedown', handleDocMouseDown, true);
  await getHostDistinct();
  await authStore.fetchAuthorized(
    [{ action: 'networkunit_view', resource_type: 'networkunit' }],
    'installTable_networkunit_view',
  );
});

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleDocMouseDown, true);
});

const getNetworkUnitListDebounced = debounce(getNetworkUnitList, 500);
watch(
  () => {
    const ids = tableData.value?.map((item: any) => Number(item.bk_networkarea_id)) ?? [];
    return [...new Set(ids)].sort((a, b) => a - b).join(',');
  },
  async (val: string) => {
    if (!props.isReinstall && !props.isUpgrade) return;
    if (!val) {
      networkUnitList.value = [];
      networkUnitGroupMap.value = {};
      return;
    }
    getNetworkUnitListDebounced();
  },
  { immediate: true },
);

watch(() => type.value, () => {
  clearAllErrors();
}, { immediate: true });

defineExpose({ tableValidate, showSetting });
</script>

<style lang="postcss" scoped>
::v-deep(.vxe-header--column) {
  font-weight: normal !important;
  font-size: 12px !important;
}
::v-deep(.vxe-body--column) {
  height: 56px !important;
  font-size: 12px !important;
}
/* 防止校验状态变化导致 VxeTable 列宽重算、右侧出现空白列 */
::v-deep(.vxe-table--body-wrapper),
::v-deep(.vxe-table--header-wrapper) {
  overflow-x: hidden;
}
::v-deep(.vxe-body--column .vxe-cell) {
  overflow: visible;
}

/*
 * 不可编辑单元格公共样式
 * 与 bkui-vue Input disabled 保持视觉一致
 */
.cell-disabled,
.cell-disabled--error {
  display: flex;
  align-items: center;
  width: 100%;
  height: 32px;
  padding: 0 10px;
  border: 1px solid #dcdee5;
  border-radius: 2px;
  box-sizing: border-box;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* 禁用态（无错误） */
.cell-disabled {
  background-color: #f5f7fa;
  color: #c4c6cc;
  cursor: not-allowed;
}

/* 校验失败态 */
.cell-disabled--error {
  background-color: #fff0f0;
  border-color: #ea3636;
  color: #ea3636;
}

/* 空值占位符提示 */
.cell-placeholder {
  color: #c4c6cc;
}
</style>
