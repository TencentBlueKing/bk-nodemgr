// 在 <script setup> 顶部添加：
import { Message } from 'bkui-vue';
import { customRef } from 'vue';

// 实现带最小长度限制的 customRef
export default function useMinLengthRef<T>(initialValue: T[], message: string, minLength = 1) {
  let value = initialValue;
  let timer: ReturnType<typeof setTimeout>;

  return customRef<T[]>((track, trigger) => ({
    get() {
      track();
      return value;
    },
    set(newVal) {
      if (newVal.length >= minLength) {
        value = newVal;
        trigger();
      } else {
        // 保留最后一次有效值
        const lastValidValue = [...value];

        // 显示提示信息
        Message({
          theme: 'warning',
          message,
        });

        // 在下个事件循环恢复值
        timer && clearTimeout(timer);
        timer = setTimeout(() => {
          value = lastValidValue;
          trigger();
        }, 0);
      }
    },
  }));
}
