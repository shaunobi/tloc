import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..", "..");

test("release inspection ignores only an explicit GitHub 404", () => {
  const workflow = readFileSync(
    resolve(repoRoot, ".github", "workflows", "release.yml"),
    "utf8",
  );
  const inspection = workflow
    .split("- name: Inspect existing release", 2)[1]
    .split("- name: Build and update GitHub draft", 1)[0];

  assert.match(inspection, /gh api/);
  assert.match(inspection, /HTTP 404/);
  assert.match(inspection, /exit "\$status"/);
  assert.doesNotMatch(inspection, /\|\| true/);
  assert.doesNotMatch(inspection, /jq -e/);
});

test("CI uses pull requests for branch coverage and pushes for main", () => {
  const workflow = readFileSync(
    resolve(repoRoot, ".github", "workflows", "ci.yml"),
    "utf8",
  );

  assert.match(workflow, /push:\s*\n\s*branches:\s*\n\s*- main/);
  assert.match(workflow, /\n\s*pull_request:\s*\n/);
  assert.match(workflow, /\n\s*workflow_call:\s*\n/);
  assert.doesNotMatch(workflow, /branches:\s*\n\s*- ["']?\*\*/);
});
