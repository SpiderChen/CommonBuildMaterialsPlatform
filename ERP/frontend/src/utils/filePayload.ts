export type BrowserFilePayload = {
  fileName: string;
  fileType: string;
  base64: string;
  url: string;
  checksum: string;
  size: number;
};

export async function browserFilePayload(file: File): Promise<BrowserFilePayload> {
  const buffer = await file.arrayBuffer();
  const fileType = file.type || "application/octet-stream";
  const base64 = arrayBufferToBase64(buffer);
  return {
    fileName: file.name,
    fileType,
    base64,
    url: `data:${fileType};base64,${base64}`,
    checksum: await sha256Checksum(buffer),
    size: file.size
  };
}

async function sha256Checksum(buffer: ArrayBuffer) {
  const subtle = globalThis.crypto?.subtle;
  if (!subtle) return "";
  const digest = await subtle.digest("SHA-256", buffer);
  return `sha256:${hex(new Uint8Array(digest))}`;
}

function arrayBufferToBase64(buffer: ArrayBuffer) {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  const chunkSize = 0x8000;
  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(offset, offset + chunkSize));
  }
  return btoa(binary);
}

function hex(bytes: Uint8Array) {
  return Array.from(bytes).map((byte) => byte.toString(16).padStart(2, "0")).join("");
}
