import { reactive } from 'vue';

export interface TableErrorActions {
  errorMap: Record<number, Record<string, string>>;
  getError: (rowIndex: number, field: string) => string;
  setError: (rowIndex: number, field: string, msg: string) => void;
  clearError: (rowIndex: number, field: string) => void;
  clearAllErrors: () => void;
  shiftErrors: (index: number, offset: number) => void;
}

export default function useTableErrors(): TableErrorActions {
  const errorMap = reactive<Record<number, Record<string, string>>>({});

  const getError = (rowIndex: number, field: string) => errorMap[rowIndex]?.[field] || '';

  const setError = (rowIndex: number, field: string, msg: string) => {
    if (!errorMap[rowIndex]) errorMap[rowIndex] = {};
    errorMap[rowIndex][field] = msg;
  };

  const clearError = (rowIndex: number, field: string) => {
    if (errorMap[rowIndex]) delete errorMap[rowIndex][field];
  };

  const clearAllErrors = () => {
    for (const key in errorMap) delete errorMap[key];
  };

  const shiftErrors = (index: number, offset: number) => {
    const newMap: Record<number, Record<string, string>> = {};
    Object.keys(errorMap).forEach((keyStr) => {
      const k = Number(keyStr);
      if (offset === 1) {
        if (k <= index) newMap[k] = errorMap[k];
        else newMap[k + 1] = errorMap[k];
      } else if (offset === -1) {
        if (k < index) newMap[k] = errorMap[k];
        else if (k > index) newMap[k - 1] = errorMap[k];
      }
    });
    for (const k in errorMap) delete errorMap[k];
    Object.assign(errorMap, newMap);
  };

  return { errorMap, getError, setError, clearError, clearAllErrors, shiftErrors };
}
