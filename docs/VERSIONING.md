# Versioning

This document describes how versioning works in the present2pdf project.

## How it works

The version is embedded into the binary at build time using Go's `-ldflags` flag. The version string is injected into the `version` variable in `main.go`.

## Checking version

To check the version of a built binary:

```bash
./present2pdf -version
```

Example output:
```
present2pdf version 1.0.0
```

## Building with version

### Automatic version from git

By default, the version is automatically determined from git tags:

```bash
make build
```

This uses `git describe --tags --always --dirty` to generate a version string:
- If you're on a tagged commit: `v1.0.0`
- If you're after a tag: `v1.0.0-3-g1234567` (3 commits after v1.0.0)
- If the working directory has uncommitted changes: adds `-dirty` suffix
- If no tags exist: uses commit hash like `2322264`

### Manual version

You can override the version manually:

```bash
make build VERSION=1.0.0
```

Or when installing:

```bash
make install VERSION=1.0.0
```

### Direct go build

If you're building without make:

```bash
go build -ldflags "-X main.version=1.0.0" -o present2pdf ./cmd/present2pdf
```

## Development version

When building without any version information, the default version is `dev`:

```bash
go build -o present2pdf ./cmd/present2pdf
./present2pdf -version
# Output: present2pdf version dev
```

## Release process

When creating a release:

1. Tag the commit with a version:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. Build the release binary:
   ```bash
   make build VERSION=1.0.0
   ```

   Or let it auto-detect from git:
   ```bash
   make build
   ```

3. The binary will include the version information.

## Implementation details

The version system works through:

1. **Variable in main.go**:
   ```go
   var version = "dev"
   ```

2. **Build-time injection in Makefile**:
   ```makefile
   VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
   LDFLAGS := -X main.version=$(VERSION)
   ```

3. **CLI flag in main.go**:
   ```go
   showVersion := flag.Bool("version", false, "Show version information and exit")
   if *showVersion {
       fmt.Printf("present2pdf version %s\n", version)
       os.Exit(0)
   }
   ```
