#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { pathToFileURL } from "node:url";
import { normalizeVersion } from "./stage.mjs";

const registry = "https://registry.npmjs.org/";

export function distTagForVersion(input) {
  const version = normalizeVersion(input);
  const versionWithoutBuildMetadata = version.split("+", 1)[0];
  return versionWithoutBuildMetadata.includes("-") ? "next" : "latest";
}

// The wrapper must remain last: users should not be offered a wrapper version
// until all of its optional platform packages are available.
export const PACKAGE_DIRECTORIES = Object.freeze([
  "tloc-darwin-arm64",
  "tloc-darwin-x64",
  "tloc-linux-arm64",
  "tloc-linux-x64",
  "tloc-win32-arm64",
  "tloc-win32-x64",
  "tloc",
]);

export function runNpm(
  args,
  { capture = false, env = process.env } = {},
) {
  const cliCandidates = [
    process.env.npm_execpath,
    resolve(dirname(process.execPath), "node_modules", "npm", "bin", "npm-cli.js"),
    resolve(
      dirname(process.execPath),
      "..",
      "lib",
      "node_modules",
      "npm",
      "bin",
      "npm-cli.js",
    ),
  ].filter(Boolean);
  const cli = cliCandidates.find((candidate) => existsSync(candidate));
  const command = cli ? process.execPath : process.platform === "win32" ? "npm.cmd" : "npm";
  const commandArguments = cli ? [cli, ...args] : args;
  return spawnSync(command, commandArguments, {
    encoding: "utf8",
    env,
    stdio: capture ? "pipe" : "inherit",
    windowsHide: true,
  });
}

function assertCommand(result, description) {
  if (result.error) {
    throw new Error(`${description}: ${result.error.message}`);
  }
  if (result.status !== 0) {
    const detail = [result.stdout, result.stderr]
      .filter(Boolean)
      .join("\n")
      .trim();
    throw new Error(`${description}${detail ? `: ${detail}` : ""}`);
  }
}

function isRegistryMiss(result) {
  const output = `${result.stdout || ""}\n${result.stderr || ""}`;
  return /(?:E404|404 Not Found|is not in this registry)/i.test(output);
}

function loadManifest(directory) {
  const manifestPath = resolve(directory, "package.json");
  if (!existsSync(manifestPath)) {
    throw new Error(`staged package is missing ${manifestPath}`);
  }
  return JSON.parse(readFileSync(manifestPath, "utf8"));
}

export function publishPackages({
  rootDir,
  version,
  dryRun = false,
  run = runNpm,
  log = console.log,
}) {
  rootDir = resolve(rootDir);
  version = normalizeVersion(version);
  const distTag = distTagForVersion(version);
  const results = [];

  for (const packageDirectory of PACKAGE_DIRECTORIES) {
    const directory = resolve(rootDir, packageDirectory);
    const manifest = loadManifest(directory);
    if (manifest.version !== version) {
      throw new Error(
        `${manifest.name} has version ${manifest.version}, expected ${version}`,
      );
    }

    if (dryRun) {
      const published = run(
        [
          "publish",
          directory,
          "--access",
          "public",
          "--tag",
          distTag,
          "--registry",
          registry,
          "--dry-run",
          "--json",
        ],
        { capture: true },
      );
      assertCommand(published, `validate ${manifest.name}@${version}`);
      log(`Validated ${manifest.name}@${version} with dist-tag ${distTag}.`);
      results.push({ name: manifest.name, action: "validated", distTag });
      continue;
    }

    const specifier = `${manifest.name}@${version}`;
    const viewed = run(
      ["view", specifier, "version", "--json", "--registry", registry],
      { capture: true },
    );
    if (viewed.error) {
      throw new Error(`check ${specifier}: ${viewed.error.message}`);
    }
    if (viewed.status === 0) {
      log(`Skipping ${specifier}; it is already published.`);
      results.push({ name: manifest.name, action: "skipped", distTag });
      continue;
    }
    if (!isRegistryMiss(viewed)) {
      assertCommand(viewed, `check ${specifier}`);
    }

    const published = run([
      "publish",
      directory,
      "--access",
      "public",
      "--tag",
      distTag,
      "--registry",
      registry,
    ]);
    assertCommand(published, `publish ${specifier}`);
    log(`Published ${specifier} with dist-tag ${distTag}.`);
    results.push({ name: manifest.name, action: "published", distTag });
  }

  return results;
}

function parseArguments(argv) {
  const options = { rootDir: resolve("dist", "npm"), dryRun: false };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--root") {
      options.rootDir = resolve(argv[++index] || "");
    } else if (argument === "--version") {
      options.version = argv[++index];
    } else if (argument === "--dry-run") {
      options.dryRun = true;
    } else {
      throw new Error(`unknown argument: ${argument}`);
    }
  }
  if (!options.version) {
    throw new Error("--version is required");
  }
  return options;
}

const invokedAsScript =
  process.argv[1] &&
  import.meta.url === pathToFileURL(resolve(process.argv[1])).href;
if (invokedAsScript) {
  try {
    publishPackages(parseArguments(process.argv.slice(2)));
  } catch (error) {
    console.error(`publish npm packages: ${error.message}`);
    process.exitCode = 1;
  }
}
