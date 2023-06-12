// Generate a new XID using the xid-generator tool.
export async function generateXid(): Promise<string> {
  const generator = __dirname + '/../../../tools/xid-generator/main.go';
  const { exec } = require('child_process');
  const { promisify } = require('util');
  const execAsync = promisify(exec);
  const { stdout } = await execAsync(`go run ${generator}`);
  return stdout.trim();
}
