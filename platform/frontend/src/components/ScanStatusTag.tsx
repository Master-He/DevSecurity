import { Tag } from "antd";
import type { ScanStatus } from "../types";

const labels: Record<ScanStatus, string> = {
  QUEUED: "排队中",
  RUNNING: "扫描中",
  SUCCEEDED: "已完成",
  FAILED: "失败",
  CANCELED: "已取消"
};

const colors: Record<ScanStatus, string> = {
  QUEUED: "default",
  RUNNING: "processing",
  SUCCEEDED: "success",
  FAILED: "error",
  CANCELED: "warning"
};

export default function ScanStatusTag({ status }: { status: ScanStatus }) {
  return <Tag color={colors[status]}>{labels[status]}</Tag>;
}
