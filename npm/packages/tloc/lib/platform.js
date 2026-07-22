"use strict";

const TARGETS = Object.freeze({
  "darwin-arm64": Object.freeze({
    packageName: "@shaunobi/tloc-darwin-arm64",
    binary: "bin/tloc",
  }),
  "darwin-x64": Object.freeze({
    packageName: "@shaunobi/tloc-darwin-x64",
    binary: "bin/tloc",
  }),
  "linux-arm64": Object.freeze({
    packageName: "@shaunobi/tloc-linux-arm64",
    binary: "bin/tloc",
  }),
  "linux-x64": Object.freeze({
    packageName: "@shaunobi/tloc-linux-x64",
    binary: "bin/tloc",
  }),
  "win32-arm64": Object.freeze({
    packageName: "@shaunobi/tloc-win32-arm64",
    binary: "bin/tloc.exe",
  }),
  "win32-x64": Object.freeze({
    packageName: "@shaunobi/tloc-win32-x64",
    binary: "bin/tloc.exe",
  }),
});

function targetFor(platform = process.platform, arch = process.arch) {
  return TARGETS[`${platform}-${arch}`] || null;
}

function resolveBinary({
  platform = process.platform,
  arch = process.arch,
  requireResolve = require.resolve,
} = {}) {
  const target = targetFor(platform, arch);
  if (!target) {
    const error = new Error(
      `no prebuilt binary is available for ${platform}/${arch}. ` +
        `Supported targets: ${Object.keys(TARGETS).join(", ")}.`,
    );
    error.code = "TLOC_UNSUPPORTED_PLATFORM";
    throw error;
  }

  const request = `${target.packageName}/${target.binary}`;
  try {
    return requireResolve(request);
  } catch (cause) {
    const error = new Error(
      `the optional package ${target.packageName} is missing for ` +
        `${platform}/${arch}. Reinstall @shaunobi/tloc without ` +
        `--omit=optional, or install with "go install ` +
        `github.com/shaunobi/tloc@latest".`,
    );
    error.code = "TLOC_PLATFORM_PACKAGE_MISSING";
    error.cause = cause;
    throw error;
  }
}

module.exports = {
  TARGETS,
  resolveBinary,
  targetFor,
};
