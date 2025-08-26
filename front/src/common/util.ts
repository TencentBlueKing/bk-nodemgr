import dayjs from "dayjs";

/**
 *  @param {number}  timezone -12 - 12
 */
export function formatTimeByTimezone(date: string, timezone?: number, fmt = 'YYYY-mm-dd HH:MM:SS') {
  let formatTime = '--';
  if (date) {
    try {
      const currentTimezone = new Date().getTimezoneOffset() / -60;
      const offsetTimezone = timezone || timezone === 0 ? currentTimezone - timezone : 0;
      const timeString = new Date(date).getTime();
      formatTime = filterTimeFormat(new Date(timeString - (offsetTimezone * 60 * 60 * 1000)), fmt);
    } catch (err) {}
  }
  return formatTime;
}
// 格式化时间戳 dateFormat(date, "YYYY-mm-dd HH:MM:SS")
export function filterTimeFormat(date: string | Date, fmt = 'YYYY-mm-dd HH:MM:SS'): string {
  if (!date) return '';
  const newDate = new Date(date);
  const fmtArr = ['Y+', 'm+', 'd+', 'H+', 'M+', 'S+'];
  const opt: { [key: string]: string } = {
    'Y+': newDate.getFullYear().toString(),        // 年
    'm+': (newDate.getMonth() + 1).toString(),     // 月
    'd+': newDate.getDate().toString(),            // 日
    'H+': newDate.getHours().toString(),           // 时
    'M+': newDate.getMinutes().toString(),         // 分
    'S+': newDate.getSeconds().toString(),         // 秒
  };
  let res;
  let time = fmt;
  fmtArr.forEach((key) => {
    res = new RegExp(`(${key})`).exec(fmt);
    if (res) {
      time = time.replace(res[1], (res[1].length === 1) ? (opt[key]) : (opt[key].padStart(res[1].length, '0')));
    }
  });
  return time;
}

export function getTimeStamp(date: string | Date) {
  return dayjs(date).unix();
};

// 秒级时间戳转时间格式
export function formatTimestamp(
  timestamp: number,
  format = "YYYY-MM-DD HH:mm:ss"
) {
  // 将时间戳转换为字符串以检查其长度
  const timestampString = timestamp.toString();

  // 如果长度为 10，假定为秒级时间戳；如果为 13，假定为毫秒级时间戳
  const isSecondLevel = timestampString.length === 10;
  const date = isSecondLevel ? dayjs.unix(timestamp) : dayjs(timestamp);

  // 格式化日期时间为 "YYYY-MM-DD HH:mm:ss" 格式
  return date.format(format);
}

// 首字母大写
export function capitalizeFirstLetter(str: string) {
  if (!str) return str; // 处理空字符串的情况
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export function compareVersions(a: string, b: string) {
  if (a === b) return 0;
  if (!a || !b) return a ? 1 : -1;
  // 解析版本号字符串，返回数字数组
  const parseVersion = (version: string) => version?.match(/\d+/g)?.map(Number);

  // 获取版本号的数字数组
  const versionA = parseVersion(a) || [];
  const versionB = parseVersion(b) || [];

  // 比较主版本号、次版本号、补丁号和附加编号
  for (let i = 0; i < versionA.length; i++) {
    const diff = versionA[i] - versionB[i];
    if (diff !== 0) return diff;
  }

  return 0; // 全部相同
}

export function bytesToMegabytes(bytes: number) {
  const megabytes = bytes / (1024 * 1024);
  return megabytes.toFixed(2);
}