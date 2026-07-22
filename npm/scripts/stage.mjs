#!/usr/bin/env node

import {
  chmodSync,
  copyFileSync,
  cpSync,
  existsSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { dirname, isAbsolute, parse, relative, resolve, sep } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = resolve(scriptDirectory, "..", "..");

export const TARGETS = Object.freeze([
  Object.freeze({
    goos: "darwin",
    goarch: "arm64",
    directory: "tloc-darwin-arm64",
    binary: "tloc",
  }),
  Object.freeze({
    goos: "darwin",
    goarch: "amd64",
    directory: "tloc-darwin-x64",
    binary: "tloc",
  }),
  Object.freeze({
    goos: "linux",
    goarch: "arm64",
    directory: "tloc-linux-arm64",
    binary: "tloc",
  }),
  Object.freeze({
    goos: "linux",
    goarch: "amd64",
    directory: "tloc-linux-x64",
    binary: "tloc",
  }),
  Object.freeze({
    goos: "windows",
    goarch: "arm64",
    directory: "tloc-win32-arm64",
    binary: "tloc.exe",
  }),
  Object.freeze({
    goos: "windows",
    goarch: "amd64",
    directory: "tloc-win32-x64",
    binary: "tloc.exe",
  }),
]);

export function normalizeVersion(input) {
  const version = String(input || "").trim().replace(/^v/, "");
  const identifier = "[0-9A-Za-z-]+(?:\\.[0-9A-Za-z-]+)*";
  const semver = new RegExp(
    `^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)` +
      `(?:-${identifier})?(?:\\+${identifier})?$`,
  );
  if (!semver.test(version)) {
    throw new Error(`invalid release version: ${JSON.stringify(input)}`);
  }
  return version;
}

function artifactValue(artifact, lower, upper) {
  return artifact[lower] ?? artifact[upper];
}

export function selectBinaryArtifacts(artifacts) {
  if (!Array.isArray(artifacts)) {
    throw new Error("GoReleaser artifacts JSON must contain an array");
  }

  return TARGETS.map((target) => {
    const matches = artifacts.filter((artifact) => {
      const type = String(artifactValue(artifact, "type", "Type") || "");
      const goos = artifactValue(artifact, "goos", "Goos");
      const goarch = artifactValue(artifact, "goarch", "Goarch");
      return (
        type.toLowerCase() === "binary" &&
        goos === target.goos &&
        goarch === target.goarch
      );
    });

    if (matches.length !== 1) {
      throw new Error(
        `expected exactly one GoReleaser binary for ` +
          `${target.goos}/${target.goarch}, found ${matches.length}`,
      );
    }
    return { target, artifact: matches[0] };
  });
}

function resolveArtifactPath(rawPath, repoRoot, artifactsFile) {
  if (!rawPath) {
    throw new Error("GoReleaser binary artifact is missing its path");
  }

  const candidates = isAbsolute(rawPath)
    ? [rawPath]
    : [resolve(repoRoot, rawPath), resolve(dirname(artifactsFile), rawPath)];
  for (const candidate of candidates) {
    if (existsSync(candidate)) {
      return candidate;
    }
  }
  throw new Error(
    `GoReleaser binary does not exist at ${candidates.join(" or ")}`,
  );
}

function containsPath(parent, candidate) {
  const relation = relative(parent, candidate);
  return (
    relation === "" ||
    (relation !== ".." &&
      !relation.startsWith(`..${sep}`) &&
      !isAbsolute(relation))
  );
}

export function assertSafeOutput(outputDir, repoRoot) {
  outputDir = resolve(outputDir);
  repoRoot = resolve(repoRoot);
  const filesystemRoot = parse(outputDir).root;
  const distRoot = resolve(repoRoot, "dist");

  const replacesRepository = containsPath(outputDir, repoRoot);
  const insideRepository = containsPath(repoRoot, outputDir);
  const insideDist = containsPath(distRoot, outputDir);
  if (
    outputDir === filesystemRoot ||
    replacesRepository ||
    (insideRepository && (outputDir === distRoot || !insideDist))
  ) {
    throw new Error(
      `refusing to replace unsafe output directory ${outputDir}; ` +
        `repository output must be a child of ${distRoot}`,
    );
  }
}

function readJSON(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function writeJSON(path, value) {
  writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`, "utf8");
}

function copyTemplate({
  source,
  destination,
  version,
  repoRoot,
  wrapper,
}) {
  cpSync(source, destination, { recursive: true });
  for (const filename of ["README.md", "LICENSE"]) {
    const sourceDocument = resolve(repoRoot, filename);
    if (!existsSync(sourceDocument)) {
      throw new Error(`required package document is missing: ${sourceDocument}`);
    }
    copyFileSync(sourceDocument, resolve(destination, filename));
  }

  const manifestPath = resolve(destination, "package.json");
  const manifest = readJSON(manifestPath);
  delete manifest.private;
  manifest.version = version;
  if (wrapper) {
    for (const dependency of Object.keys(manifest.optionalDependencies || {})) {
      manifest.optionalDependencies[dependency] = version;
    }
  }
  writeJSON(manifestPath, manifest);
}

export function stagePackages({
  version,
  artifactsFile = resolve(defaultRepoRoot, "dist", "artifacts.json"),
  outputDir = resolve(defaultRepoRoot, "dist", "npm"),
  repoRoot = defaultRepoRoot,
  log = console.log,
}) {
  version = normalizeVersion(version);
  repoRoot = resolve(repoRoot);
  artifactsFile = resolve(artifactsFile);
  outputDir = resolve(outputDir);
  const packagesDir = resolve(repoRoot, "npm", "packages");

  assertSafeOutput(outputDir, repoRoot);

  const document = readJSON(artifactsFile);
  const artifacts = Array.isArray(document) ? document : document.artifacts;
  const selections = selectBinaryArtifacts(artifacts).map(
    ({ target, artifact }) => ({
      target,
      artifact,
      binaryPath: resolveArtifactPath(
        artifactValue(artifact, "path", "Path"),
        repoRoot,
        artifactsFile,
      ),
    }),
  );

  rmSync(outputDir, { recursive: true, force: true });
  mkdirSync(outputDir, { recursive: true });

  const wrapperDestination = resolve(outputDir, "tloc");
  copyTemplate({
    source: resolve(packagesDir, "tloc"),
    destination: wrapperDestination,
    version,
    repoRoot,
    wrapper: true,
  });
  // npm preserves the source mode in the tarball. The release runs on Linux,
  // so make the JavaScript bin entry executable before packing it.
  chmodSync(resolve(wrapperDestination, "bin", "tloc.js"), 0o755);

  const staged = [];
  for (const selection of selections) {
    const { target, binaryPath } = selection;
    const destination = resolve(outputDir, target.directory);
    copyTemplate({
      source: resolve(packagesDir, target.directory),
      destination,
      version,
      repoRoot,
      wrapper: false,
    });

    const binaryDestination = resolve(destination, "bin", target.binary);
    mkdirSync(dirname(binaryDestination), { recursive: true });
    copyFileSync(binaryPath, binaryDestination);
    chmodSync(binaryDestination, 0o755);
    staged.push({ ...target, binaryPath: binaryDestination });
  }

  log(`Staged @shaunobi/tloc ${version} and ${staged.length} platform packages.`);
  return { version, outputDir, wrapperDestination, staged };
}

function parseArguments(argv) {
  const options = {
    artifactsFile: resolve(defaultRepoRoot, "dist", "artifacts.json"),
    outputDir: resolve(defaultRepoRoot, "dist", "npm"),
  };

  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--version") {
      options.version = argv[++index];
    } else if (argument === "--artifacts") {
      options.artifactsFile = resolve(argv[++index] || "");
    } else if (argument === "--output") {
      options.outputDir = resolve(argv[++index] || "");
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
    stagePackages(parseArguments(process.argv.slice(2)));
  } catch (error) {
    console.error(`stage npm packages: ${error.message}`);
    process.exitCode = 1;
  }
}
