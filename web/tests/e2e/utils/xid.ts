import { exec } from "node:child_process";
import { promisify } from "node:util";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
export async function generateXid(): Promise<string> {
  const generator = join(__dirname, "../../../../tools/xid-generator/main.go");
  const { stdout } = await execAsync(`go run ${generator}`);
  return stdout.trim();
}
