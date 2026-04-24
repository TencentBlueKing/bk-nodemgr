/**
 * @blueking/bkui-form 类型声明
 * 基于 bkui-vue 3.0 和 JSON Schema 的表单渲染器
 */

declare module '@blueking/bkui-form' {
  import type { Component } from 'vue';

  interface CreateFormOptions {
    components?: Record<string, Component>;
  }

  function createForm(options?: CreateFormOptions): Component;
  export default createForm;
}

declare module '@blueking/bkui-form/dist/bkui-form.css' {
  const content: string;
  export default content;
}
