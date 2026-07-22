#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const { resolveBinary } = require("../lib/platform.js");

let binary;
try {
  binary = resolveBinary();
} catch (error) {
  console.error(`tloc: ${error.message}`);
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: true,
});

if (result.error) {
  console.error(`tloc: failed to start ${binary}: ${result.error.message}`);
  process.exit(1);
}

if (result.signal) {
  if (process.platform !== "win32") {
    process.kill(process.pid, result.signal);
  }
  process.exit(1);
}

process.exit(typeof result.status === "number" ? result.status : 1);
