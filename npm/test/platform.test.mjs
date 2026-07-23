import assert from "node:assert/strict";
import { createRequire } from "node:module";
import test from "node:test";

const require = createRequire(import.meta.url);
const { TARGETS, resolveBinary, targetFor } = require(
  "../packages/tloc/lib/platform.js",
);

test("targetFor maps every supported Node platform and architecture", () => {
  const expected = new Map([
    ["darwin-arm64", "@shaunobi/tloc-darwin-arm64"],
    ["darwin-x64", "@shaunobi/tloc-darwin-x64"],
    ["linux-arm64", "@shaunobi/tloc-linux-arm64"],
    ["linux-x64", "@shaunobi/tloc-linux-x64"],
    ["win32-arm64", "@shaunobi/tloc-win32-arm64"],
    ["win32-x64", "@shaunobi/tloc-win32-x64"],
  ]);

  assert.deepEqual(Object.keys(TARGETS), [...expected.keys()]);
  for (const [key, packageName] of expected) {
    const [platform, arch] = key.split("-");
    assert.equal(targetFor(platform, arch).packageName, packageName);
  }
  assert.equal(targetFor("freebsd", "x64"), null);
  assert.equal(targetFor("linux", "arm"), null);
});

test("resolveBinary resolves the binary from the selected optional package", () => {
  const requests = [];
  const binary = resolveBinary({
    platform: "win32",
    arch: "arm64",
    requireResolve(request) {
      requests.push(request);
      return "C:\\tloc\\tloc.exe";
    },
  });

  assert.equal(binary, "C:\\tloc\\tloc.exe");
  assert.deepEqual(requests, ["@shaunobi/tloc-win32-arm64/bin/tloc.exe"]);
});

test("resolveBinary explains unsupported targets", () => {
  assert.throws(
    () => resolveBinary({ platform: "freebsd", arch: "x64" }),
    (error) => {
      assert.equal(error.code, "TLOC_UNSUPPORTED_PLATFORM");
      assert.match(error.message, /freebsd\/x64/);
      return true;
    },
  );
});

test("resolveBinary explains omitted optional dependencies", () => {
  assert.throws(
    () =>
      resolveBinary({
        platform: "linux",
        arch: "x64",
        requireResolve() {
          throw new Error("not found");
        },
      }),
    (error) => {
      assert.equal(error.code, "TLOC_PLATFORM_PACKAGE_MISSING");
      assert.match(error.message, /--omit=optional/);
      assert.match(error.message, /@shaunobi\/tloc-linux-x64/);
      assert.match(error.message, /package-lock\.json/);
      assert.match(error.message, /node_modules/);
      assert.match(error.message, /Remove both/);
      return true;
    },
  );
});
