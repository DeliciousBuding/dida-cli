#!/usr/bin/env node

const fs = require("fs");
const os = require("os");
const path = require("path");
const crypto = require("crypto");
const https = require("https");
const { execFileSync } = require("child_process");

const repo = process.env.DIDA_REPO || "DeliciousBuding/dida-cli";
const version = process.env.DIDA_VERSION || "";
const binDir = path.join(__dirname, "..", "bin");

function platformName() {
  if (process.platform === "win32") return "windows";
  if (process.platform === "linux") return "linux";
  if (process.platform === "darwin") return "darwin";
  throw new Error(`unsupported platform: ${process.platform}`);
}

function archName() {
  if (process.arch === "x64") return "amd64";
  if (process.arch === "arm64") return "arm64";
  throw new Error(`unsupported architecture: ${process.arch}`);
}

function requestBuffer(url) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { "User-Agent": "dida-cli-npm-installer" } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        requestBuffer(res.headers.location).then(resolve, reject);
        return;
      }
      if (res.statusCode !== 200) {
        reject(new Error(`download failed ${res.statusCode}: ${url}`));
        return;
      }
      const chunks = [];
      res.on("data", (chunk) => chunks.push(chunk));
      res.on("end", () => resolve(Buffer.concat(chunks)));
    }).on("error", reject);
  });
}

async function latestVersion() {
  const data = await requestBuffer(`https://api.github.com/repos/${repo}/releases/latest`);
  return JSON.parse(data.toString("utf8")).tag_name;
}

function sha256(buffer) {
  return crypto.createHash("sha256").update(buffer).digest("hex");
}

async function main() {
  const resolvedVersion = version || await latestVersion();
  const osName = platformName();
  const arch = archName();
  const ext = osName === "windows" ? "zip" : "tar.gz";
  const exe = osName === "windows" ? "dida.exe" : "dida";
  const asset = `dida_${resolvedVersion}_${osName}_${arch}.${ext}`;
  const base = `https://github.com/${repo}/releases/download/${resolvedVersion}`;

  const [archive, checksums] = await Promise.all([
    requestBuffer(`${base}/${asset}`),
    requestBuffer(`${base}/checksums.txt`)
  ]);
  const line = checksums.toString("utf8").split(/\r?\n/).find((item) => item.endsWith(`  ${asset}`));
  if (!line) throw new Error(`checksum not found for ${asset}`);
  const expected = line.split(/\s+/)[0].toLowerCase();
  const actual = sha256(archive);
  if (actual !== expected) throw new Error(`checksum mismatch for ${asset}`);

  const temp = fs.mkdtempSync(path.join(os.tmpdir(), "dida-npm-"));
  try {
    const archivePath = path.join(temp, asset);
    fs.writeFileSync(archivePath, archive);
    if (ext === "zip") {
      execFileSync("powershell", ["-NoProfile", "-Command", `Expand-Archive -LiteralPath '${archivePath.replace(/'/g, "''")}' -DestinationPath '${temp.replace(/'/g, "''")}' -Force`], { stdio: "inherit" });
    } else {
      execFileSync("tar", ["-xzf", archivePath, "-C", temp], { stdio: "inherit" });
    }
    const found = findFile(temp, exe);
    if (!found) throw new Error("binary not found in archive");
    fs.mkdirSync(binDir, { recursive: true });
    const target = path.join(binDir, exe);
    fs.copyFileSync(found, target);
    if (osName !== "windows") fs.chmodSync(target, 0o755);
  } finally {
    fs.rmSync(temp, { recursive: true, force: true });
  }
}

function findFile(dir, fileName) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      const hit = findFile(full, fileName);
      if (hit) return hit;
    } else if (entry.name === fileName) {
      return full;
    }
  }
  return null;
}

main().catch((error) => {
  console.error(`dida-cli install failed: ${error.message}`);
  process.exit(1);
});
