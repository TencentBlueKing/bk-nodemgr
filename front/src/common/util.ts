
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

