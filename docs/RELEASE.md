# Release Process

This document describes how to create and publish a release of present2pdf.

## Automated Releases via GitHub Actions

Releases are built automatically when you push a version tag to the repository.

### Prerequisites

- Push access to the repository
- All changes committed

### Steps

1. Ensure the code is ready for release and all tests pass:
   ```bash
   make test
   ```

2. Create an annotated tag (use [semver](https://semver.org/) format):
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   ```

3. Push the tag to trigger the release workflow:
   ```bash
   git push origin v1.0.0
   ```

4. GitHub Actions will:
   - Run the [goreleaser workflow](../.github/workflows/release.yml)
   - Build binaries for Linux, Windows, and macOS (amd64 and arm64)
   - Create a GitHub Release with changelog
   - Attach ZIP archives with binaries

5. Check the [Releases](https://github.com/ekalinin/present2pdf/releases) page for the new release.

### Tag Format

- Use semantic versioning: `vMAJOR.MINOR.PATCH` (e.g., `v1.0.0`, `v1.2.3`)
- Tags must match the pattern `v*` to trigger the workflow

## Release Artifacts

GoReleaser produces the following:

| Platform | Architectures | Format |
|----------|---------------|--------|
| Linux    | amd64, arm64  | zip    |
| Windows  | amd64, arm64  | zip    |
| macOS    | amd64, arm64  | zip    |

Archive names follow the pattern: `present2pdf_<OS>_<ARCH>.zip`

## Local Release Build

To build release binaries locally without publishing:

```bash
# Install GoReleaser (if not installed)
# https://goreleaser.com/install/

# Dry run (builds artifacts to dist/ without publishing)
goreleaser release --snapshot --clean

# Or build a specific version
make build VERSION=1.0.0
```

For version embedding details, see [VERSIONING.md](VERSIONING.md).

## Changelog

The release changelog is auto-generated from git commits since the previous tag. Commit messages are grouped by:

- **Features** — commits matching `feat(...):`
- **Bug Fixes** — commits matching `fix(...):`
- **Documentation** — commits matching `docs(...):`
- **Others** — remaining commits

Excluded: `docs:`, `test:`, `chore:` prefixes.

