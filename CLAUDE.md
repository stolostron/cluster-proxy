# Cluster Proxy Downstream/Upstream Sync Guide

## Repository Relationship

This repo (`stolostron/cluster-proxy`) is a **downstream fork** of the upstream community project
[`open-cluster-management-io/cluster-proxy`](https://github.com/open-cluster-management-io/cluster-proxy).

The upstream repo is cloned separately on disk (for example under a dedicated path for community repos).
The downstream adds Red Hat / ACM-specific files and features on top of the shared upstream codebase.

| Attribute | Downstream (this repo) | Upstream |
|-----------|------------------------|----------|
| GitHub org | `stolostron` | `open-cluster-management-io` |
| Branch | `main` | `main` |

Git remote names are customizable per developer. Throughout this guide, commands discover them
dynamically from `git remote -v` -- look for `stolostron` in the URL to identify the downstream
remote and `open-cluster-management-io` for the upstream remote.

## Downstream-Only Files

These files exist ONLY in the downstream fork and must be preserved during syncs:

- **`.tekton/`** -- Tekton pipeline configs for Konflux CI (MCE component builds)
- **`renovate.json`** -- Renovate bot config extending stolostron/acm-config
- **`sonar-project.properties`** -- SonarQube code analysis config
- **`cmd/Dockerfile.rhtap`** -- RHTAP/Konflux-oriented container build (used by Tekton pipelines)
- **`cmd/pure.Dockerfile`** -- Prow CI container build (used by OpenShift CI prow jobs)
- **`OWNERS`** -- The downstream OWNERS lists the Red Hat ACM team, which is entirely different
  from the upstream community maintainers. Always preserve the stolostron version during syncs.

## Dockerfile Landscape

There are four Dockerfiles in this repo. Understanding which is used by which CI system is
important for syncs:

| Dockerfile | Used by | Builder image | Notes |
|---|---|---|---|
| `cmd/Dockerfile` | Upstream CI; local `make images` (GitHub Actions **disabled** on stolostron) | `golang:<version>` | **Upstream-owned.** Take upstream's version during syncs -- no conflict resolution needed. Upstream downloads vanilla apiserver-network-proxy tarball; stolostron does NOT use this file in active CI. |
| `cmd/Dockerfile.rhtap` | **Konflux/Tekton pipelines** (`.tekton/`) | `brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_<version>` | Downstream-only. Builds apiserver-network-proxy from `third_party/` submodule with `CGO_ENABLED=1`. Update Go version when upstream bumps `go.mod`. |
| `cmd/pure.Dockerfile` | **Prow CI** (config in `openshift/release` repo at `ci-operator/config/stolostron/cluster-proxy/`) | `registry.ci.openshift.org/stolostron/builder:go<version>-linux` | Downstream-only. Builds apiserver-network-proxy from `third_party/` submodule with `CGO_ENABLED=1`. Post-merge, the built image is published to the stolostron registry. Update Go version when upstream bumps `go.mod`. |
| `test/e2e/Dockerfile` | E2e tests (upstream) | `golang:<version>-alpine` | Upstream-owned. Comes through the merge automatically. |

Both `cmd/Dockerfile.rhtap` and `cmd/pure.Dockerfile` build the apiserver-network-proxy proxy-agent
and proxy-server binaries from the `third_party/apiserver-network-proxy` git submodule via
`COPY . .` -- they automatically pick up the new apiserver-network-proxy code when the submodule
is updated.

## Downstream-Only Commits

The downstream fork contains commits that will never exist in upstream:

- **Tekton file additions** -- `.tekton/` pipeline YAML files for Konflux
- **Konflux/RHTAP commits** -- Red Hat CI/CD pipeline integration
- **Renovate config changes** -- Downstream-specific dependency management

## apiserver-network-proxy Submodule

This repo includes a git submodule for apiserver-network-proxy:

```
third_party/apiserver-network-proxy → stolostron/apiserver-network-proxy (branch: v0.36.0-patch)
```

This submodule is a **separate sync concern** -- it tracks a patched fork of the upstream
apiserver-network-proxy. The proxy-agent and proxy-server binaries are built from this submodule
during Docker image creation. Submodule updates are done independently from the main repo sync.

The stolostron fork adds:
- A `replace` directive pointing to `stolostron/grpc-go` for HTTPS forward proxy support
  (upstream grpc-go does not support HTTPS proxies as of v1.81.x)

Upstream (`open-cluster-management-io/cluster-proxy`) does NOT use a submodule -- it downloads
vanilla apiserver-network-proxy tarballs at build time. The submodule is therefore always a
**merge conflict** during syncs. Resolution: always keep the submodule
(`git checkout HEAD -- third_party/apiserver-network-proxy`).

## Expected Merge Conflicts During Syncs

These files conflict on every upstream sync:

| File | Why it conflicts | Resolution |
|---|---|---|
| `OWNERS` | Stolostron team vs upstream community maintainers | `git checkout --ours OWNERS` |
| `third_party/apiserver-network-proxy` | Upstream deleted the submodule; stolostron keeps it | `git checkout HEAD -- third_party/apiserver-network-proxy` |
| `cmd/Dockerfile` | Was previously modified downstream; now taken from upstream | `git checkout --theirs cmd/Dockerfile` |

## Historical Sync Patterns

Two methods have been used to sync:

### 1. GitHub "Sync Fork" Button (Bulk Merge)

Creates commits like `Merge branch 'open-cluster-management-io:main' into main`. Brings in
all upstream commits since the last sync as a single merge commit. This is the primary method.

### 2. Manual Cherry-Pick Sync PRs (Targeted)

When a specific upstream PR is urgently needed before the next bulk sync, someone creates a
branch like `sync-upstream-pr-NNNN` that cherry-picks or recreates the upstream commit.
This creates **different SHAs** for the same content, which can cause merge conflicts later.

## How to Sync (Step-by-Step)

### Prerequisites

Ensure a remote pointing to the upstream community repo is configured. It can point to a local
clone or directly to the GitHub URL:

```bash
# Option A: point to a local clone
git remote add upstream /path/to/upstream_cluster-proxy

# Option B: point to GitHub directly
git remote add upstream git@github.com:open-cluster-management-io/cluster-proxy.git
```

Before running any sync commands, discover the correct remote names from your local config:

```bash
# Find the upstream remote (points to open-cluster-management-io)
UPSTREAM=$(git remote -v | awk '/open-cluster-management-io.*(push)/ {print $1; exit}')

# Find the downstream remote (points to stolostron)
DOWNSTREAM=$(git remote -v | awk '/stolostron.*(push)/ {print $1; exit}')

echo "UPSTREAM=${UPSTREAM}  DOWNSTREAM=${DOWNSTREAM}"
```

Use `${UPSTREAM}` and `${DOWNSTREAM}` in all commands below in place of hardcoded remote names.

### Step 1: Fetch upstream

```bash
git fetch "${UPSTREAM}" main
```

### Step 2: Create a sync branch

```bash
git checkout "${DOWNSTREAM}/main"
git checkout -b sync-upstream-$(date +%Y%m%d)
```

### Step 3: Merge upstream

```bash
git merge "${UPSTREAM}/main" --no-ff
```

Using `--no-ff` ensures a merge commit is created even if a fast-forward were possible,
making the sync point visible in history. Individual upstream commits are preserved with
their original SHAs.

### Step 4: Resolve conflicts

Three files always conflict (see "Expected Merge Conflicts" section above):

```bash
# OWNERS: always keep stolostron version
git checkout --ours OWNERS

# cmd/Dockerfile: always take upstream version
git checkout --theirs cmd/Dockerfile

# third_party/apiserver-network-proxy: keep submodule, restore from HEAD
git checkout HEAD -- third_party/apiserver-network-proxy
```

Then update the submodule to the correct version and branch:

```bash
# Update .gitmodules if the apiserver-network-proxy branch has changed (e.g. v0.34.0-patch -> v0.36.0-patch)
# Edit .gitmodules branch field, then:
git submodule update --init --recursive --remote
```

Also check `go.mod` for conflicts from dependency version differences:

Resolution strategy:
- For dependencies that exist in both repos: **keep the higher version**
- For downstream-only dependencies (rare): keep them
- **`OWNERS`**: always keep the stolostron/downstream version -- use `git checkout --ours OWNERS`

After resolving `go.mod` conflicts, regenerate `go.sum` automatically:

```bash
go mod tidy
```

### Step 5: Verify vendor integrity and build

After resolving all conflicts, run these commands to catch any remaining issues:

```bash
# Stage resolved files
git add OWNERS cmd/Dockerfile .gitmodules third_party/apiserver-network-proxy

# Re-sync vendor directory
go mod vendor

# Confirm the code compiles cleanly and passes linting
make lint
```

If `go mod vendor` produces changes, stage them before committing.
If `make lint` fails, trace the error back to the conflicted file and fix it.

### Step 6: Verify upstream commits are preserved

```bash
# upstream/main must be a reachable ancestor of your branch HEAD
git merge-base --is-ancestor "${UPSTREAM}/main" HEAD && echo "OK - all upstream commits preserved" || echo "ERROR - upstream commits missing!"

# Should print nothing (no commits in upstream not in your branch)
git log --oneline "${UPSTREAM}/main" --not HEAD
```

Run this AFTER the merge commit is created (not before).

### Step 7: Verify downstream-only files

After resolving conflicts, verify these files still exist and are correct:

```bash
ls .tekton/
cat renovate.json
cat sonar-project.properties
test -f cmd/Dockerfile.rhtap && echo "Dockerfile.rhtap OK"
test -f cmd/pure.Dockerfile && echo "pure.Dockerfile OK"
cat OWNERS
```

### Step 8: Commit the merge (signed)

```bash
git commit -s -S -m "Sync with upstream open-cluster-management-io/cluster-proxy $(date +%Y-%m-%d)"
```

Include in the commit message:
- List of upstream commits being brought in
- Conflict resolution summary
- Reminder about merge commit when merging the PR

### Step 9: Update downstream Dockerfiles

`cmd/Dockerfile.rhtap` and `cmd/pure.Dockerfile` are downstream-only files that parallel the
upstream `cmd/Dockerfile`. During a sync, first check whether the upstream `cmd/Dockerfile` was
modified by any of the incoming commits:

```bash
git log --oneline "${UPSTREAM}/main" --not HEAD~ -- cmd/Dockerfile
```

If upstream changed its `cmd/Dockerfile` (new build stages, added dependencies, changed
`COPY`/`RUN` instructions, etc.), evaluate whether `cmd/Dockerfile.rhtap` and `cmd/pure.Dockerfile`
need equivalent updates. The files intentionally differ -- the downstream Dockerfiles use Red Hat
builder images and build apiserver-network-proxy from the `third_party/` submodule -- but structural build logic
changes often need to be mirrored manually.

If the upstream sync includes a **Go version bump** (check `go.mod` for a changed `go` directive),
the downstream-only Dockerfiles must be updated manually -- they are not touched by the merge.

```bash
# Check current Go version in go.mod
go list -m -json go

# Check current versions in downstream Dockerfiles
grep -i "^FROM .* AS builder$" cmd/Dockerfile.rhtap cmd/pure.Dockerfile
```

Update `cmd/Dockerfile.rhtap` (builder image tag pattern is `rhel_<rhel-version>_<goversion>`, e.g.):

```
FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_1.26 AS builder
```

Update `cmd/pure.Dockerfile` (builder image tag pattern is `go<goversion>-linux`):

```
FROM registry.ci.openshift.org/stolostron/builder:go1.26-linux AS builder
```

Note: `cmd/Dockerfile` and `test/e2e/Dockerfile` are upstream-owned and updated automatically
by the merge. The prow CI config (`openshift/release` repo) may also need its `build_root`
tag updated if the builder image changes -- check
`ci-operator/config/stolostron/cluster-proxy/stolostron-cluster-proxy-main.yaml`.
In the future, a `.ci-operator.yaml` file in this repo can replace that external config for
specifying the build root image.

Commit these Dockerfile changes separately (signed):

```bash
git add cmd/Dockerfile.rhtap cmd/pure.Dockerfile
git commit -s -S -m "chore: upgrade downstream Dockerfiles to Go <version>"
```

### Step 10: Push and open PR

```bash
git push "${DOWNSTREAM}" sync-upstream-$(date +%Y%m%d)
```

Then create a PR on GitHub targeting `stolostron/cluster-proxy:main`.

**Important:** When merging the PR, always use **"Create a merge commit"** -- not squash or
rebase. This preserves the original upstream commit SHAs in the downstream history, making
future syncs and divergence checks accurate.

## Handling Manual/Urgent Cherry-Picks

If a specific upstream PR is needed before the next bulk sync:

1. Identify the upstream PR number and commit SHA
2. Create a branch: `git checkout -b sync-upstream-pr-NNNN`
3. Cherry-pick: `git cherry-pick <upstream-sha>`
4. PR title convention: `:seedling: [sync] <original commit message>`
5. PR body should reference: `Synced from upstream PR: open-cluster-management-io/cluster-proxy#NNNN`

**Warning:** Cherry-picks create different SHAs. The next bulk sync merge will likely have
conflicts in go.mod where both versions appear. Resolve by keeping the higher version, then run `go mod tidy`.

## Diagnosing Sync State

To check how far behind downstream is from upstream:

```bash
# Fetch latest upstream
git fetch "${UPSTREAM}" main

# See commits in upstream that aren't in downstream
git log --oneline "${UPSTREAM}/main" --not "${DOWNSTREAM}/main"

# See downstream-only commits
git log --oneline "${DOWNSTREAM}/main" --not "${UPSTREAM}/main"
```
