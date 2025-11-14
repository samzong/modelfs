import { expect, test } from "vitest";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function walk(dir: string, out: string[] = []): string[] {
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const e of entries) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) {
      // skip build artifacts if any
      if (!["dist", "node_modules"].includes(e.name)) {
        walk(p, out);
      }
    } else {
      out.push(p);
    }
  }
  return out;
}

const root = path.join(__dirname, "..");
const allowExts = new Set([".ts", ".tsx", ".js", ".jsx"]);
const files = walk(root).filter((f) => allowExts.has(path.extname(f)));

test("no Chinese characters in UI source", () => {
  const offenders: string[] = [];
  const re = /[\u4e00-\u9fff]/;
  for (const f of files) {
    const txt = fs.readFileSync(f, "utf8");
    if (re.test(txt)) offenders.push(path.relative(root, f));
  }
  expect(offenders, `Files containing Chinese characters:\n${offenders.join("\n")}`).toHaveLength(0);
});

