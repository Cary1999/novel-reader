import { describe, expect, it } from "vitest";
import { MAX_COVER_BYTES, MAX_UPLOAD_BYTES, validateCoverUploadFile, validateTxtUploadFile } from "../uploadLimits";

describe("validateTxtUploadFile", () => {
  it("allows txt files up to 50MB", () => {
    const file = new File(["x"], "novel.txt", { type: "text/plain" });
    Object.defineProperty(file, "size", { value: MAX_UPLOAD_BYTES });

    expect(validateTxtUploadFile(file)).toBe("");
  });

  it("rejects files larger than 50MB", () => {
    const file = new File(["x"], "novel.txt", { type: "text/plain" });
    Object.defineProperty(file, "size", { value: MAX_UPLOAD_BYTES + 1 });

    expect(validateTxtUploadFile(file)).toBe("文件不能超过 50MB");
  });
});

describe("validateCoverUploadFile", () => {
  it("allows jpg/png/webp files up to 10MB", () => {
    const file = new File(["x"], "cover.webp", { type: "image/webp" });
    Object.defineProperty(file, "size", { value: MAX_COVER_BYTES });

    expect(validateCoverUploadFile(file)).toBe("");
  });

  it("rejects unsupported cover formats", () => {
    const file = new File(["x"], "cover.gif", { type: "image/gif" });

    expect(validateCoverUploadFile(file)).toBe("仅支持 JPG、PNG、WebP 封面图片");
  });

  it("rejects cover files larger than 10MB", () => {
    const file = new File(["x"], "cover.png", { type: "image/png" });
    Object.defineProperty(file, "size", { value: MAX_COVER_BYTES + 1 });

    expect(validateCoverUploadFile(file)).toBe("封面图片不能超过 10MB");
  });
});
