import { describe, expect, it } from "vitest";
import { MAX_UPLOAD_BYTES, validateTxtUploadFile } from "../uploadLimits";

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

