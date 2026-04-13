export interface IVersionLog {
  version: string;
  date: string;
  is_current?: boolean;
  content?: string;
}

export interface IVersionLogsListResponse {
  data?: {
    version_logs?: IVersionLog[];
  };
}

export interface ILog {
  title: string;
  date: string;
  detail: string;
  isCurrent: boolean;
}

export const mapVersionLogs = (response: IVersionLogsListResponse): ILog[] => {
  const versionLogs = response.data?.version_logs ?? [];

  return versionLogs.map(item => ({
    title: item.version,
    date: item.date,
    detail: item.content || '',
    isCurrent: item.is_current === true,
  }));
};

export const resolveCurrentIndex = (logs: ILog[]): number => {
  const currentIndex = logs.findIndex(item => item.isCurrent);
  return currentIndex >= 0 ? currentIndex : 0;
};
