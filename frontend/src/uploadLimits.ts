export const MAX_UPLOAD_BYTES = 50 * 1024 * 1024;
export const MAX_UPLOAD_LABEL = "50MB";
export const MAX_COVER_BYTES = 10 * 1024 * 1024;
export const MAX_COVER_LABEL = "10MB";

export function validateTxtUploadFile(candidate: File | null) {
  if (!candidate) return "请选择 txt 文件";
  if (!candidate.name.toLowerCase().endsWith(".txt")) return "仅支持 .txt 文件";
  if (candidate.size <= 0) return "文件不能为空";
  if (candidate.size > MAX_UPLOAD_BYTES) return `文件不能超过 ${MAX_UPLOAD_LABEL}`;
  return "";
}

export function validateCoverUploadFile(candidate: File | null) {
  if (!candidate) return "请选择封面图片";
  const lowerName = candidate.name.toLowerCase();
  const allowedTypes = ["image/jpeg", "image/png", "image/webp"];
  const isAllowedExt = [".jpg", ".jpeg", ".png", ".webp"].some((ext) => lowerName.endsWith(ext));
  if (!allowedTypes.includes(candidate.type) && !isAllowedExt) return "仅支持 JPG、PNG、WebP 封面图片";
  if (candidate.size <= 0) return "封面图片不能为空";
  if (candidate.size > MAX_COVER_BYTES) return `封面图片不能超过 ${MAX_COVER_LABEL}`;
  return "";
}
