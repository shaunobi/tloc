import assert from "node:assert/strict";
import {
  chmodSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";
import {
  assertSafeOutput,
  normalizeVersion,
  selectBinaryArtifacts,
  stagePackages,
  TARGETS,
} from "../scripts/stage.mjs";
import {
  PACKAGE_DIRECTORIES,
  runNpm,
} from "../scripts/publish.mjs";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..", "..");

test("normalizeVersion accepts release tags and rejects non-semver input", () => {
  assert.equal(normalizeVersion("v1.2.3"), "1.2.3");
  assert.equal(normalizeVersion("1.2.3-rc.1"), "1.2.3-rc.1");
  assert.throws(() => normalizeVersion("latest"), /invalid release version/);
  assert.throws(() => normalizeVersion("1.2"), /invalid release version/);
});

test("selectBinaryArtifacts rejects missing and duplicate targets", () => {
  assert.throws(() => selectBinaryArtifacts([]), /darwin\/arm64.*found 0/);

  const artifacts = TARGETS.map((target) => ({
    type: "Binary",
    goos: target.goos,
    goarch: target.goarch,
    path: target.directory,
  }));
  artifacts.push({ ...artifacts[0] });
  assert.throws(
    () => selectBinaryArtifacts(artifacts),
    /darwin\/arm64.*found 2/,
  );
});

test("assertSafeOutput protects repository source and ancestors", () => {
  assert.doesNotThrow(() =>
    assertSafeOutput(resolve(repoRoot, "dist", "npm"), repoRoot),
  );
  assert.doesNotThrow(() =>
    assertSafeOutput(resolve(tmpdir(), "tloc-external-stage"), repoRoot),
  );

  for (const outputDir of [
    repoRoot,
    dirname(repoRoot),
    resolve(repoRoot, "dist"),
    resolve(repoRoot, ".github"),
    resolve(repoRoot, "internal"),
    resolve(repoRoot, "npm", "packages"),
  ]) {
    assert.throws(
      () => assertSafeOutput(outputDir, repoRoot),
      /refusing to replace unsafe output directory/,
    );
  }
});

test("stagePackages creates a complete version-synchronized package set", (t) => {
  const temporary = mkdtempSync(join(tmpdir(), "tloc-npm-stage-"));
  t.after(() => rmSync(temporary, { recursive: true, force: true }));

  const artifacts = [];
  for (const target of TARGETS) {
    const binary = resolve(temporary, "binaries", target.directory, target.binary);
    mkdirSync(dirname(binary), { recursive: true });
    writeFileSync(binary, `${target.goos}/${target.goarch}\n`, "utf8");
    chmodSync(binary, 0o755);
    artifacts.push({
      type: "Binary",
      goos: target.goos,
      goarch: target.goarch,
      path: binary,
    });
  }
  artifacts.push({
    type: "Archive",
    goos: "linux",
    goarch: "amd64",
    path: resolve(temporary, "ignored.tar.gz"),
  });

  const artifactsFile = resolve(temporary, "artifacts.json");
  const outputDir = resolve(temporary, "staged");
  writeFileSync(artifactsFile, JSON.stringify(artifacts), "utf8");

  const messages = [];
  const result = stagePackages({
    version: "v1.4.0",
    artifactsFile,
    outputDir,
    repoRoot,
    log(message) {
      messages.push(message);
    },
  });

  assert.equal(result.version, "1.4.0");
  assert.equal(result.staged.length, 6);
  assert.match(messages[0], /6 platform packages/);

  const wrapper = JSON.parse(
    readFileSync(resolve(outputDir, "tloc", "package.json"), "utf8"),
  );
  assert.equal(wrapper.version, "1.4.0");
  assert.equal(wrapper.private, undefined);
  assert.deepEqual(new Set(Object.values(wrapper.optionalDependencies)), new Set(["1.4.0"]));
  if (process.platform !== "win32") {
    assert.notEqual(
      statSync(resolve(outputDir, "tloc", "bin", "tloc.js")).mode & 0o111,
      0,
    );
  }

  for (const target of TARGETS) {
    const directory = resolve(outputDir, target.directory);
    const packageManifest = JSON.parse(
      readFileSync(resolve(directory, "package.json"), "utf8"),
    );
    assert.equal(packageManifest.version, "1.4.0");
    assert.equal(packageManifest.private, undefined);
    assert.equal(
      readFileSync(resolve(directory, "bin", target.binary), "utf8"),
      `${target.goos}/${target.goarch}\n`,
    );
    assert.ok(statSync(resolve(directory, "README.md")).isFile());
    assert.ok(statSync(resolve(directory, "LICENSE")).isFile());
    if (process.platform !== "win32") {
      assert.notEqual(
        statSync(resolve(directory, "bin", target.binary)).mode & 0o111,
        0,
      );
    }
  }

  const sourceWrapper = JSON.parse(
    readFileSync(resolve(repoRoot, "npm", "packages", "tloc", "package.json"), "utf8"),
  );
  assert.equal(sourceWrapper.version, "0.0.0");
  assert.equal(sourceWrapper.private, true);

  for (const directory of PACKAGE_DIRECTORIES) {
    const packed = runNpm(
      ["pack", resolve(outputDir, directory), "--dry-run", "--json"],
      {
        capture: true,
        env: {
          ...process.env,
          npm_config_cache: resolve(temporary, "npm-cache"),
        },
      },
    );
    assert.equal(
      packed.status,
      0,
      `npm pack failed for ${directory}: ${packed.stderr || packed.error}`,
    );
    const result = JSON.parse(packed.stdout);
    const files = new Set(result[0].files.map((file) => file.path));
    assert.ok(files.has("package.json"));
    assert.ok(files.has("README.md"));
    assert.ok(files.has("LICENSE"));
    if (directory === "tloc") {
      assert.ok(files.has("bin/tloc.js"));
      assert.ok(files.has("lib/platform.js"));
    } else {
      const target = TARGETS.find((candidate) => candidate.directory === directory);
      assert.ok(files.has(`bin/${target.binary}`));
    }
  }

});
