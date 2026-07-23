#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import { delimiter, dirname, extname, resolve, sep } from "node:path";
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

function npmCliFromCommandShim(shimPath, { exists, readFile }) {
  const shimDirectory = dirname(shimPath);
  const adjacentCandidates = [
    resolve(shimDirectory, "node_modules", "npm", "bin", "npm-cli.js"),
    resolve(
      shimDirectory,
      "..",
      "lib",
      "node_modules",
      "npm",
      "bin",
      "npm-cli.js",
    ),
  ];
  for (const candidate of adjacentCandidates) {
    if (exists(candidate)) {
      return candidate;
    }
  }

  let source;
  try {
    source = readFile(shimPath, "utf8");
  } catch {
    return null;
  }

  // npm's Windows shim ultimately invokes npm-cli.js. Resolve that file and
  // execute it with Node instead of opting into shell parsing for arbitrary
  // publish paths and arguments.
  for (const match of source.matchAll(/["']([^"'\r\n]*npm-cli\.js)["']/gi)) {
    const expanded = match[1].replace(
      /%(?:~dp0|dp0%)/gi,
      `${shimDirectory}${sep}`,
    );
    if (/%[^%]+%/.test(expanded)) {
      continue;
    }
    const candidate = resolve(shimDirectory, expanded);
    if (exists(candidate)) {
      return candidate;
    }
  }
  return null;
}

function invocationForCandidate(
  candidate,
  { platform, nodeExecutable, exists, readFile },
) {
  if (!candidate || !exists(candidate)) {
    return null;
  }

  const extension = extname(candidate).toLowerCase();
  if (extension === ".js" || extension === ".cjs" || extension === ".mjs") {
    return { command: nodeExecutable, commandArguments: [candidate] };
  }
  if (
    platform === "win32" &&
    (extension === ".cmd" || extension === ".bat")
  ) {
    const cli = npmCliFromCommandShim(candidate, { exists, readFile });
    return cli
      ? { command: nodeExecutable, commandArguments: [cli] }
      : null;
  }
  return { command: candidate, commandArguments: [] };
}

export function resolveNpmInvocation({
  env = process.env,
  platform = process.platform,
  nodeExecutable = process.execPath,
  cliCandidates,
  exists = existsSync,
  readFile = readFileSync,
} = {}) {
  const candidates =
    cliCandidates ??
    [
      env.npm_execpath,
      resolve(
        dirname(nodeExecutable),
        "node_modules",
        "npm",
        "bin",
        "npm-cli.js",
      ),
      resolve(
        dirname(nodeExecutable),
        "..",
        "lib",
        "node_modules",
        "npm",
        "bin",
        "npm-cli.js",
      ),
    ].filter(Boolean);

  for (const candidate of candidates) {
    const invocation = invocationForCandidate(candidate, {
      platform,
      nodeExecutable,
      exists,
      readFile,
    });
    if (invocation) {
      return invocation;
    }
  }

  if (platform !== "win32") {
    return { command: "npm", commandArguments: [] };
  }

  const pathValue = env.Path ?? env.PATH ?? env.path ?? "";
  for (const directory of pathValue.split(delimiter).filter(Boolean)) {
    for (const filename of ["npm.exe", "npm.cmd", "npm.bat"]) {
      const invocation = invocationForCandidate(resolve(directory, filename), {
        platform,
        nodeExecutable,
        exists,
        readFile,
      });
      if (invocation) {
        return invocation;
      }
    }
  }
  return null;
}

export function runNpm(
  args,
  {
    capture = false,
    env = process.env,
    platform = process.platform,
    nodeExecutable = process.execPath,
    cliCandidates,
    exists = existsSync,
    readFile = readFileSync,
  } = {},
) {
  const invocation = resolveNpmInvocation({
    env,
    platform,
    nodeExecutable,
    cliCandidates,
    exists,
    readFile,
  });
  if (!invocation) {
    const error = new Error(
      "could not locate npm-cli.js behind npm.cmd; set npm_execpath or install npm next to Node",
    );
    error.code = "ENOENT";
    return {
      error,
      status: null,
      signal: null,
      stdout: capture ? "" : null,
      stderr: capture ? "" : null,
    };
  }
  return spawnSync(
    invocation.command,
    [...invocation.commandArguments, ...args],
    {
      encoding: "utf8",
      env,
      stdio: capture ? "pipe" : "inherit",
      windowsHide: true,
    },
  );
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
