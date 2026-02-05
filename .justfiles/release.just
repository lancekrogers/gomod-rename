# Release recipes

VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

# Show available release commands
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Show current version
[no-cd]
version:
    @echo "{{VERSION}}"

# Tag a new release (e.g., just release tag v1.0.0)
[no-cd]
tag TAG:
    #!/usr/bin/env bash
    set -euo pipefail
    if git rev-parse "{{TAG}}" >/dev/null 2>&1; then
        echo "Error: tag {{TAG}} already exists"
        exit 1
    fi
    if ! [[ "{{TAG}}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Error: tag must be semver (e.g., v1.0.0)"
        exit 1
    fi
    echo "Current version: {{VERSION}}"
    echo "New tag: {{TAG}}"
    echo ""
    read -p "Create and push tag {{TAG}}? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag "{{TAG}}"
        git push origin "{{TAG}}"
        echo "Tagged and pushed {{TAG}}"
    else
        echo "Aborted."
    fi

# Create a GitHub release from the latest tag
[no-cd]
github TAG="":
    #!/usr/bin/env bash
    set -euo pipefail
    tag="{{TAG}}"
    if [ -z "$tag" ]; then
        tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
        if [ -z "$tag" ]; then
            echo "Error: no tags found. Run 'just release tag v1.0.0' first."
            exit 1
        fi
    fi
    echo "Creating GitHub release for $tag..."
    gh release create "$tag" --generate-notes --title "$tag"

# Show all tags
[no-cd]
tags:
    @git tag -l --sort=-v:refname

# Full release: check, tag, push, create GitHub release
[no-cd]
publish TAG:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Running checks..."
    just check
    echo ""
    echo "All checks passed. Proceeding with release."
    echo ""
    just release tag "{{TAG}}"
    just release github "{{TAG}}"
