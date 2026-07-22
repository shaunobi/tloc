import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import test from "node:test";
import {
  distTagForVersion,
  PACKAGE_DIRECTORIES,
  publishPackages,
} from "../scripts/publish.mjs";

function stagedPackages(t, version = "1.2.3") {
  const root = mkdtempSync(join(tmpdir(), "tloc-npm-publish-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  for (const directory of PACKAGE_DIRECTORIES) {
    const packageDirectory = resolve(root, directory);
    mkdirSync(packageDirectory, { recursive: true });
    writeFileSync(
      resolve(packageDirectory, "package.json"),
      JSON.stringify({ name: `@shaunobi/${directory}`, version }),
      "utf8",
    );
  }
  return root;
}

test("distTagForVersion keeps prereleases away from latest", () => {
  assert.equal(distTagForVersion("v1.2.3"), "latest");
  assert.equal(distTagForVersion("1.2.3+build.4"), "latest");
  assert.equal(distTagForVersion("1.2.3+build-4"), "latest");
  assert.equal(distTagForVersion("1.2.3-rc.1"), "next");
  assert.equal(distTagForVersion("v2.0.0-beta.2+build.4"), "next");
});

test("publishPackages skips existing versions and keeps the wrapper last", (t) => {
  const rootDir = stagedPackages(t);
  const calls = [];
  const messages = [];

  const results = publishPackages({
    rootDir,
    version: "1.2.3",
    log(message) {
      messages.push(message);
    },
    run(args) {
      calls.push(args);
      if (args[0] === "view") {
        if (args[1] === "@shaunobi/tloc-linux-x64@1.2.3") {
          return { status: 0, stdout: '"1.2.3"', stderr: "" };
        }
        return { status: 1, stdout: "", stderr: "npm error code E404" };
      }
      return { status: 0, stdout: "", stderr: "" };
    },
  });

  const publishCalls = calls.filter((args) => args[0] === "publish");
  assert.equal(publishCalls.length, 6);
  assert.equal(publishCalls.at(-1)[1], resolve(rootDir, "tloc"));
  assert.ok(publishCalls.every((args) => args.includes("latest")));
  assert.ok(publishCalls.every((args) => !args.includes("--provenance")));
  assert.ok(
    publishCalls.every((args) => !args[1].endsWith("tloc-linux-x64")),
  );
  assert.equal(results.find((result) => result.name.endsWith("linux-x64")).action, "skipped");
  assert.match(messages.at(-1), /Published @shaunobi\/tloc@1.2.3/);
});

test("publishPackages dry-run validates every package without registry access", (t) => {
  const rootDir = stagedPackages(t, "2.0.0-rc.1");
  const calls = [];

  const results = publishPackages({
    rootDir,
    version: "2.0.0-rc.1",
    dryRun: true,
    log() {},
    run(args) {
      calls.push(args);
      return { status: 0, stdout: "[]", stderr: "" };
    },
  });

  assert.equal(results.length, 7);
  assert.ok(results.every((result) => result.action === "validated"));
  assert.ok(results.every((result) => result.distTag === "next"));
  assert.ok(calls.every((args) => args[0] === "publish"));
  assert.ok(calls.every((args) => args.includes("--dry-run")));
  assert.ok(calls.every((args) => args.includes("next")));
  assert.ok(calls.every((args) => !args.includes("--provenance")));
  assert.equal(calls.at(-1)[1], resolve(rootDir, "tloc"));
});
