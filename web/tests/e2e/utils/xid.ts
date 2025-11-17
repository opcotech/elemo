import { exec } from "node:child_process";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { promisify } from "node:util";

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
export async function generateXid(): Promise<string> {
  const generator = join(__dirname, "../../../../tools/xid-generator/main.go");
  const { stdout } = await execAsync(`go run ${generator}`);
  return stdout.trim();
}
