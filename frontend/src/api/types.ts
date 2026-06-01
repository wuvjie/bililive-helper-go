export interface StreamerInfo {
  name: string;
  files: number;
  size_bytes: number;
  size_gb: number;
}

export interface StatusResponse {
  disk_usage: number;
  streamers: StreamerInfo[];
  total_gb: number;
}

export interface DiskInfo {
  usage_pct: number;
  total_gb: number;
  used_gb: number;
  free_gb: number;
}

export interface PendingInfo {
  original_files: number;
  original_size_gb: number;
}

export interface DetailResponse {
  disk: DiskInfo;
  pending: PendingInfo;
  streamers: StreamerInfo[];
  schedule: ScheduleStatus;
}

export interface DayStats {
  date: string;
  merge_count: number;
  merge_bytes: number;
  clean_count: number;
  clean_bytes: number;
}

export interface StatsResponse {
  today: Omit<DayStats, "date">;
  month: Omit<DayStats, "date">;
  daily: DayStats[];
}

export interface TaskStatus {
  enabled: boolean;
  interval: number;
  last_run: number;
  next_run: number;
  is_running: boolean;
}

export interface ScheduleStatus {
  running: boolean;
  merge_enabled: boolean;
  merge_interval: number;
  clean_enabled: boolean;
  clean_interval: number;
  merge: TaskStatus;
  clean: TaskStatus;
}

export interface ScheduleConfig {
  merge_enabled: boolean;
  clean_enabled: boolean;
  merge_interval: number;
  clean_interval: number;
  BACKUP_START_HOUR?: number;
  BACKUP_START_MINUTE?: number;
  BACKUP_END_HOUR?: number;
  BACKUP_END_MINUTE?: number;
}

export interface HistoryRecord {
  id: string;
  time: string;
  task: string;
  streamer: string;
  status: string;
  files_count: number;
  freed_bytes: number;
  merged_bytes: number;
  duration: number;
  detail: string;
}

export interface HistoryResponse {
  items: HistoryRecord[];
  total: number;
  page: number;
  per_page: number;
  pages: number;
}

export interface LogFile {
  date: string;
  filename: string;
  mtime: string;
  task: string;
}

export interface StreamerFile {
  name: string;
  size: number;
  size_str: string;
  mtime: string;
  is_merged: boolean;
}

export interface SetupCheck {
  ffmpeg_ok: boolean;
  ffprobe_ok: boolean;
  target_dir_exists: boolean;
  target_dir_writable: boolean;
  ffmpeg_path: string;
  ffprobe_path: string;
  streamer_count: number;
  video_count: number;
  total_size_gb: number;
  disk_total_gb: number;
  disk_free_gb: number;
  disk_usage_pct: number;
  process_group_ok: boolean;
}

export interface CleanEstimate {
  file_count: number;
  total_size_gb: number;
}

export interface ConfigRecommend {
  risk_level: string;
  reason: string;
  analysis: Record<string, any>;
  current_usage: number;
  total_gb: number;
  free_gb: number;
  need_to_free_gb: number;
  TRIGGER_THRESHOLD: number;
  TARGET_THRESHOLD: number;
  MIN_KEEP_PER_STREAMER: number;
  SAFE_AGE_MINUTES: number;
  SAFE_MODE: string;
  MERGE_AGE_MINUTES: number;
  MAX_DELETE_PER_RUN: number;
  GAP_MINUTES: number;
}
