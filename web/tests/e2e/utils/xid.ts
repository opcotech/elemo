import { randomBytes } from "node:crypto";

// Compatible with github.com/rs/xid string encoding (base32hex, no padding).
const ENCODING = "0123456789abcdefghijklmnopqrstuv";

const machineId = randomBytes(3);
const processId = process.pid & 0xffff;
let counter = randomBytes(3).readUIntBE(0, 3);

function encode(id: Buffer): string {
  const dst = Array.from({ length: 20 });
  dst[19] = ENCODING[(id[11] << 4) & 0x1f];
  dst[18] = ENCODING[(id[11] >> 1) & 0x1f];
  dst[17] = ENCODING[((id[11] >> 6) | (id[10] << 2)) & 0x1f];
  dst[16] = ENCODING[(id[10] >> 3) & 0x1f];
  dst[15] = ENCODING[id[9] & 0x1f];
  dst[14] = ENCODING[((id[9] >> 5) | (id[8] << 3)) & 0x1f];
  dst[13] = ENCODING[(id[8] >> 2) & 0x1f];
  dst[12] = ENCODING[((id[8] >> 7) | (id[7] << 1)) & 0x1f];
  dst[11] = ENCODING[((id[7] >> 4) | (id[6] << 4)) & 0x1f];
  dst[10] = ENCODING[(id[6] >> 1) & 0x1f];
  dst[9] = ENCODING[((id[6] >> 6) | (id[5] << 2)) & 0x1f];
  dst[8] = ENCODING[(id[5] >> 3) & 0x1f];
  dst[7] = ENCODING[id[4] & 0x1f];
  dst[6] = ENCODING[((id[4] >> 5) | (id[3] << 3)) & 0x1f];
  dst[5] = ENCODING[(id[3] >> 2) & 0x1f];
  dst[4] = ENCODING[((id[3] >> 7) | (id[2] << 1)) & 0x1f];
  dst[3] = ENCODING[((id[2] >> 4) | (id[1] << 4)) & 0x1f];
  dst[2] = ENCODING[(id[1] >> 1) & 0x1f];
  dst[1] = ENCODING[((id[1] >> 6) | (id[0] << 2)) & 0x1f];
  dst[0] = ENCODING[(id[0] >> 3) & 0x1f];
  return dst.join("");
}

/**
 * Generate an XID-compatible identifier without spawning a Go subprocess.
 * Used by E2E DB helpers that previously called `go run tools/xid-generator`.
 */
export function generateXid(): string {
  const id = Buffer.alloc(12);
  const now = Math.floor(Date.now() / 1000);

  id.writeUInt32BE(now, 0);
  machineId.copy(id, 4);
  id[7] = (processId >> 8) & 0xff;
  id[8] = processId & 0xff;

  counter = (counter + 1) & 0xffffff;
  id[9] = (counter >> 16) & 0xff;
  id[10] = (counter >> 8) & 0xff;
  id[11] = counter & 0xff;

  return encode(id);
}
