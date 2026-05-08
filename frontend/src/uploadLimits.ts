export const MAX_UPLOAD_BYTES = 50 * 1024 * 1024;
export const MAX_UPLOAD_LABEL = "50MB";

export function validateTxtUploadFile(candidate: File | null) {
  if (!candidate) return "请选择 txt 文件";
  if (!candidate.name.toLowerCase().endsWith(".txt")) return "仅支持 .txt 文件";
  if (candidate.size <= 0) return "文件不能为空";
  if (candidate.size > MAX_UPLOAD_BYTES) return `文件不能超过 ${MAX_UPLOAD_LABEL}`;
  return "";
}

