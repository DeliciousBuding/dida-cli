const assert = require("assert/strict");
const { EventEmitter } = require("events");
const test = require("node:test");
const packageJson = require("../package.json");

function loadInstaller() {
  const https = require("https");
  const originalGet = https.get;
  const originalExit = process.exit;
  https.get = () => {
    const req = new EventEmitter();
    req.setTimeout = () => req;
    req.destroy = () => {};
    return req;
  };
  process.exit = (code) => {
    throw new Error(`unexpected process.exit(${code}) while loading installer`);
  };
  try {
    delete require.cache[require.resolve("./install.js")];
    return require("./install.js");
  } finally {
    https.get = originalGet;
    process.exit = originalExit;
  }
}

function fakeGet(steps) {
  const seen = [];
  const get = (url, _options, callback) => {
    seen.push(url);
    const req = new EventEmitter();
    req.setTimeout = () => req;
    req.destroy = () => {};

    queueMicrotask(() => {
      const step = steps[seen.length - 1];
      if (!step) {
        req.emit("error", new Error(`unexpected request: ${url}`));
        return;
      }
      if (step.error) {
        req.emit("error", step.error);
        return;
      }
      const res = new EventEmitter();
      res.statusCode = step.statusCode;
      res.headers = step.headers || {};
      res.resume = () => {};
      callback(res);
      const bodyChunks = Array.isArray(step.body) ? step.body : [step.body];
      for (const body of bodyChunks) {
        if (body) res.emit("data", Buffer.from(body));
      }
      res.emit("end");
    });

    return req;
  };
  get.seen = seen;
  return get;
}

const installer = loadInstaller();

test("uses package.json version as the default release version", () => {
  const plan = installer.createInstallPlan({
    env: {},
    platform: "linux",
    arch: "x64",
    repo: "owner/repo",
  });

  assert.equal(plan.version, `v${packageJson.version}`);
  assert.equal(plan.base, `https://github.com/owner/repo/releases/download/v${packageJson.version}`);
  assert.equal(plan.osName, "linux");
  assert.equal(plan.arch, "amd64");
});

test("follows relative redirects while downloading", async () => {
  const get = fakeGet([
    { statusCode: 302, headers: { location: "../v1.2.3/checksums.txt" } },
    { statusCode: 200, body: "ok" },
  ]);

  const body = await installer.requestBuffer(
    "https://example.test/releases/latest/download/checksums.txt",
    { get },
  );

  assert.equal(body.toString("utf8"), "ok");
  assert.deepEqual(get.seen, [
    "https://example.test/releases/latest/download/checksums.txt",
    "https://example.test/releases/latest/v1.2.3/checksums.txt",
  ]);
});

test("rejects redirects beyond the configured limit", async () => {
  const get = fakeGet(Array.from({ length: 7 }, () => ({
    statusCode: 302,
    headers: { location: "next" },
  })));

  await assert.rejects(
    installer.requestBuffer("https://example.test/download/checksums.txt", { get }),
    /too many redirects/,
  );
});

test("rejects responses larger than the declared byte limit", async () => {
  const get = fakeGet([
    { statusCode: 200, headers: { "content-length": "11" }, body: "small" },
  ]);

  await assert.rejects(
    installer.requestBuffer("https://example.test/archive.tar.gz", { get, maxBytes: 10 }),
    /response too large for https:\/\/example\.test\/archive\.tar\.gz: 11 bytes exceeds 10 bytes/,
  );
});

test("rejects streamed responses that exceed the byte limit", async () => {
  const get = fakeGet([
    { statusCode: 200, body: ["12345", "67890", "x"] },
  ]);

  await assert.rejects(
    installer.requestBuffer("https://example.test/archive.tar.gz", { get, maxBytes: 10 }),
    /response too large for https:\/\/example\.test\/archive\.tar\.gz: 11 bytes exceeds 10 bytes/,
  );
});

test("retries transient request errors with injected backoff", async () => {
  const get = fakeGet([
    { error: new Error("ECONNRESET") },
    { statusCode: 200, body: "ok" },
  ]);
  const delays = [];

  const body = await installer.requestBuffer("https://example.test/checksums.txt", {
    get,
    retries: 1,
    retryDelayMs: 5,
    sleep: (ms) => {
      delays.push(ms);
      return Promise.resolve();
    },
  });

  assert.equal(body.toString("utf8"), "ok");
  assert.deepEqual(delays, [5]);
  assert.equal(get.seen.length, 2);
});

test("reports final download errors with attempt count", async () => {
  const get = fakeGet([
    { statusCode: 503 },
    { statusCode: 503 },
  ]);

  await assert.rejects(
    installer.requestBuffer("https://example.test/checksums.txt", {
      get,
      retries: 1,
      retryDelayMs: 0,
      sleep: () => Promise.resolve(),
    }),
    /download failed after 2 attempts: download failed 503: https:\/\/example\.test\/checksums\.txt/,
  );
});

test("throws when archive checksum does not match", () => {
  assert.throws(
    () => installer.verifyArchiveChecksum(
      "deadbeef  dida_v1.2.3_linux_amd64.tar.gz\n",
      "dida_v1.2.3_linux_amd64.tar.gz",
      Buffer.from("archive"),
    ),
    /checksum mismatch for dida_v1\.2\.3_linux_amd64\.tar\.gz/,
  );
});

test("throws when checksum entry is missing", () => {
  assert.throws(
    () => installer.verifyArchiveChecksum(
      "deadbeef  dida_v1.2.3_darwin_arm64.tar.gz\n",
      "dida_v1.2.3_linux_amd64.tar.gz",
      Buffer.from("archive"),
    ),
    /checksum not found for dida_v1\.2\.3_linux_amd64\.tar\.gz/,
  );
});

test("rejects unsupported platform and architecture values", () => {
  assert.throws(() => installer.platformName("freebsd"), /unsupported platform: freebsd/);
  assert.throws(() => installer.archName("ia32"), /unsupported architecture: ia32/);
});

test("parses latest asset name from checksums", () => {
  const result = installer.resolveAssetFromChecksums(
    "aaa  dida_v1.2.3_darwin_arm64.tar.gz\nbbb  dida_v1.2.4_linux_amd64.tar.gz\n",
    { version: "", osName: "linux", arch: "amd64", ext: "tar.gz" },
  );

  assert.deepEqual(result, {
    asset: "dida_v1.2.4_linux_amd64.tar.gz",
    resolvedVersion: "v1.2.4",
  });
});
