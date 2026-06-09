#!/usr/bin/env node

const fs = require("fs");
const os = require("os");
const path = require("path");
const crypto = require("crypto");
const https = require("https");
const { execFileSync } = require("child_process");
const pkg = require("../package.json");

const defaultRepo = "DeliciousBuding/dida-cli";
const defaultMaxDownloadBytes = 200 * 1024 * 1024;
const binDir = path.join(__dirname, "..", "bin");

function platformName(platform = process.platform) {
  if (platform === "win32") return "windows";
  if (platform === "linux") return "linux";
  if (platform === "darwin") return "darwin";
  throw new Error(`unsupported platform: ${platform}`);
}

function archName(arch = process.arch) {
  if (arch === "x64") return "amd64";
  if (arch === "arm64") return "arm64";
  throw new Error(`unsupported architecture: ${arch}`);
}

function createInstallPlan(options = {}) {
  const env = options.env || process.env;
  const packageJson = options.packageJson || pkg;
  const repo = options.repo || env.DIDA_REPO || defaultRepo;
  const version = env.DIDA_VERSION || `v${packageJson.version}`;
  const osName = platformName(options.platform);
  const arch = archName(options.arch);
  const ext = osName === "windows" ? "zip" : "tar.gz";
  const exe = osName === "windows" ? "dida.exe" : "dida";
  const installedExe = osName === "windows" ? "dida.exe" : "dida-bin";
  const base = version
    ? `https://github.com/${repo}/releases/download/${version}`
    : `https://github.com/${repo}/releases/latest/download`;

  return { repo, version, osName, arch, ext, exe, installedExe, base };
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function resolveAssetFromChecksums(checksumsText, plan) {
  let resolvedVersion = plan.version;
  let asset = "";

  if (resolvedVersion) {
    asset = `dida_${resolvedVersion}_${plan.osName}_${plan.arch}.${plan.ext}`;
  } else {
    const suffix = `_${plan.osName}_${plan.arch}.${plan.ext}`;
    asset = checksumsText.split(/\r?\n/)
      .map((line) => line.trim().split(/\s+/)[1])
      .find((name) => name && name.startsWith("dida_v") && name.endsWith(suffix)) || "";
    if (asset) {
      const pattern = `^dida_(v[^_]+)_${plan.osName}_${plan.arch}\\.${escapeRegExp(plan.ext)}$`;
      const match = asset.match(new RegExp(pattern));
      if (match) resolvedVersion = match[1];
    }
  }

  if (!resolvedVersion || !asset) {
    throw new Error(`could not resolve latest release asset for ${plan.osName}/${plan.arch}`);
  }
  return { asset, resolvedVersion };
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function responseTooLargeError(url, actualBytes, maxBytes) {
  return new Error(`response too large for ${url}: ${actualBytes} bytes exceeds ${maxBytes} bytes`);
}

function retryableError(message) {
  const error = new Error(message);
  error.retryable = true;
  return error;
}

function shouldRetryDownload(error) {
  return error && error.retryable === true;
}

function requestBuffer(url, options = {}) {
  const requestOptions = typeof options === "number" ? { redirects: options } : options;
  const redirects = requestOptions.redirects || 0;
  const maxRedirects = requestOptions.maxRedirects ?? 5;
  const retries = requestOptions.retries ?? 2;
  const retryDelayMs = requestOptions.retryDelayMs ?? 250;
  const wait = requestOptions.wait || requestOptions.sleep || sleep;
  const get = requestOptions.get || https.get;

  return requestBufferWithRetries(url, { ...requestOptions, redirects, maxRedirects, retries, retryDelayMs, wait, get });
}

async function requestBufferWithRetries(url, requestOptions) {
  let lastError = null;
  for (let attempt = 0; attempt <= requestOptions.retries; attempt += 1) {
    try {
      return await requestBufferOnce(url, requestOptions);
    } catch (error) {
      lastError = error;
      if (attempt >= requestOptions.retries || !shouldRetryDownload(error)) {
        if (attempt > 0) {
          throw new Error(`download failed after ${attempt + 1} attempts: ${error.message}`);
        }
        throw error;
      }
      const delay = requestOptions.retryDelayMs * (2 ** attempt);
      if (delay > 0) await requestOptions.wait(delay);
      else await requestOptions.wait(0);
    }
  }
  throw lastError;
}

function requestBufferOnce(url, requestOptions) {
  const maxBytes = requestOptions.maxBytes ?? defaultMaxDownloadBytes;

  return new Promise((resolve, reject) => {
    let settled = false;
    const finish = (error, value) => {
      if (settled) return;
      settled = true;
      if (error) reject(error);
      else resolve(value);
    };

    const req = requestOptions.get(url, { headers: { "User-Agent": "dida-cli-npm-installer" }, timeout: 60000 }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        if (requestOptions.redirects >= requestOptions.maxRedirects) {
          finish(new Error(`too many redirects: ${url}`));
          return;
        }
        const next = new URL(res.headers.location, url).toString();
        requestBuffer(next, { ...requestOptions, redirects: requestOptions.redirects + 1 })
          .then((value) => finish(null, value), (error) => finish(error));
        return;
      }
      if (res.statusCode !== 200) {
        res.resume();
        const error = res.statusCode >= 500
          ? retryableError(`download failed ${res.statusCode}: ${url}`)
          : new Error(`download failed ${res.statusCode}: ${url}`);
        finish(error);
        return;
      }
      const contentLength = Number(res.headers["content-length"]);
      if (Number.isFinite(contentLength) && contentLength > maxBytes) {
        res.resume();
        finish(responseTooLargeError(url, contentLength, maxBytes));
        return;
      }
      const chunks = [];
      let received = 0;
      res.on("data", (chunk) => {
        received += chunk.length;
        if (received > maxBytes) {
          if (typeof res.destroy === "function") res.destroy();
          finish(responseTooLargeError(url, received, maxBytes));
          return;
        }
        chunks.push(chunk);
      });
      res.on("end", () => finish(null, Buffer.concat(chunks)));
      res.on("error", (error) => {
        error.retryable = true;
        finish(error);
      });
    }).on("error", (error) => {
      error.retryable = true;
      finish(error);
    });
    req.on("timeout", () => {
      req.destroy();
      finish(retryableError(`request timed out: ${url}`));
    });
  });
}

function sha256(buffer) {
  return crypto.createHash("sha256").update(buffer).digest("hex");
}

function verifyArchiveChecksum(checksumsText, asset, archive) {
  const line = checksumsText.split(/\r?\n/).find((item) => item.endsWith(`  ${asset}`));
  if (!line) throw new Error(`checksum not found for ${asset}`);
  const expected = line.split(/\s+/)[0].toLowerCase();
  const actual = sha256(archive);
  if (actual !== expected) throw new Error(`checksum mismatch for ${asset}`);
  return expected;
}

async function main(options = {}) {
  const plan = createInstallPlan(options);
  const download = options.requestBuffer || requestBuffer;
  const checksums = await download(`${plan.base}/checksums.txt`);
  const checksumsText = checksums.toString("utf8");
  const { asset } = resolveAssetFromChecksums(checksumsText, plan);

  const archive = await download(`${plan.base}/${asset}`);
  verifyArchiveChecksum(checksumsText, asset, archive);

  const temp = fs.mkdtempSync(path.join(os.tmpdir(), "dida-npm-"));
  try {
    const archivePath = path.join(temp, asset);
    fs.writeFileSync(archivePath, archive);
    if (plan.ext === "zip") {
      execFileSync("powershell", ["-NoProfile", "-Command", `Expand-Archive -LiteralPath '${archivePath.replace(/'/g, "''")}' -DestinationPath '${temp.replace(/'/g, "''")}' -Force`], { stdio: "inherit" });
    } else {
      execFileSync("tar", ["-xzf", archivePath, "-C", temp], { stdio: "inherit" });
    }
    const found = findFile(temp, plan.exe);
    if (!found) throw new Error("binary not found in archive");
    fs.mkdirSync(binDir, { recursive: true });
    const target = path.join(binDir, plan.installedExe);
    fs.copyFileSync(found, target);
    if (plan.osName !== "windows") fs.chmodSync(target, 0o755);
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

module.exports = {
  archName,
  createInstallPlan,
  findFile,
  main,
  platformName,
  requestBuffer,
  responseTooLargeError,
  resolveAssetFromChecksums,
  sha256,
  shouldRetryDownload,
  verifyArchiveChecksum,
};

if (require.main === module) {
  main().catch((error) => {
    console.error(`dida-cli install failed: ${error.message}`);
    process.exit(1);
  });
}
