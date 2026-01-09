#!/usr/bin/env bun

import { $ } from "bun";
import { parseArgs } from "util";

const { values } = parseArgs({
  args: Bun.argv.slice(2),
  options: {
    major: { type: "boolean", default: false },
    minor: { type: "boolean", default: false },
    yes: { type: "boolean", short: "y", long: "yes", default: false },
    help: { type: "boolean", short: "h", long: "help", default: false },
  },
});

if (values.help) {
  console.log(`Usage: increment_version.ts [--major | --minor] [--yes]`);
  process.exit(0);
}

const part = values.major ? "major" : values.minor ? "minor" : "patch";

// Get all tags, filter semver, sort using Bun.semver
const tags = await $`git tag --list`.text();
const versions = tags
  .trim()
  .split("\n")
  .filter((t) => /^v\d+\.\d+\.\d+$/.test(t))
  .sort(Bun.semver.order);

const current = versions.at(-1) ?? "v0.0.0";
const [major, minor, patch] = current.slice(1).split(".").map(Number);

const newVersion =
  part === "major"
    ? `v${major! + 1}.0.0`
    : part === "minor"
      ? `v${major}.${minor! + 1}.0`
      : `v${major}.${minor}.${patch! + 1}`;

console.log(`Current version: ${current}`);
console.log(`New version:     ${newVersion} (${part})`);

if (!values.yes) {
  const response = prompt("\nCreate this tag? [y/N]");
  if (!["y", "yes"].includes(response?.toLowerCase() ?? "")) {
    console.log("Aborted.");
    process.exit(0);
  }
}

await $`git tag ${newVersion}`;
console.log(`✓ Created tag ${newVersion}`);
