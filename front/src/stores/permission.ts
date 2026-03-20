import { defineStore } from 'pinia';

export interface PermissionAction {
  id: string;
  name: string;
  related_resource_types: Array<{
    system_id: string;
    type: string;
    type_name: string;
  }>;
}

export interface PermissionData {
  system: string;
  system_name: string;
  apply_url: string;
  actions: PermissionAction[];
}

export const usePermissionStore = defineStore('permission', {
  state: (): {
    visible: boolean;
    data: PermissionData | null;
  } => ({
    visible: false,
    data: null,
  }),
  actions: {
    showDialog(data: PermissionData) {
      if (this.visible) return;
      this.data = data;
      this.visible = true;
    },
    hideDialog() {
      this.visible = false;
      this.data = null;
    },
  },
});
