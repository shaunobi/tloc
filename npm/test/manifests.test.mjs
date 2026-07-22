import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";
import { TARGETS } from "../scripts/stage.mjs";

const npmRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const packagesRoot = resolve(npmRoot, "packages");

function manifest(directory) {
  return JSON.parse(
    readFileSync(resolve(packagesRoot, directory, "package.json"), "utf8"),
  );
}

test("package templates are private and have no installation lifecycle scripts", () => {
  for (const directory of ["tloc", ...TARGETS.map((target) => target.directory)]) {
    const packageManifest = manifest(directory);
    assert.equal(packageManifest.private, true);
    const scripts = packageManifest.scripts || {};
    for (const lifecycle of ["preinstall", "install", "postinstall"]) {
      assert.equal(
        scripts[lifecycle],
        undefined,
        `${packageManifest.name} must not define ${lifecycle}`,
      );
    }
    assert.deepEqual(packageManifest.publishConfig, {
      access: "public",
      registry: "https://registry.npmjs.org/",
    });
  }
});

test("wrapper declares all platform packages as exact optional dependencies", () => {
  const wrapper = manifest("tloc");
  const expected = Object.fromEntries(
    TARGETS.map((target) => [
      manifest(target.directory).name,
      wrapper.version,
    ]),
  );
  assert.deepEqual(wrapper.optionalDependencies, expected);
  assert.deepEqual(wrapper.bin, { tloc: "bin/tloc.js" });
});

test("platform manifests constrain npm to the matching target", () => {
  const archNames = { amd64: "x64", arm64: "arm64" };
  const osNames = { darwin: "darwin", linux: "linux", windows: "win32" };

  for (const target of TARGETS) {
    const packageManifest = manifest(target.directory);
    assert.deepEqual(packageManifest.os, [osNames[target.goos]]);
    assert.deepEqual(packageManifest.cpu, [archNames[target.goarch]]);
    assert.deepEqual(packageManifest.files, [`bin/${target.binary}`]);
  }
});
