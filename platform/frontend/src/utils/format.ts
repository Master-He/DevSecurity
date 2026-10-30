import dayjs from "dayjs";

export function formatTime(value?: string | null) {
  if (!value || value === "1970-01-01T00:00:00Z") {
    return "—";
  }
  return dayjs(value).format("YYYY-MM-DD HH:mm");
}
