# Certificate Monkey Versioning Strategy

This document describes the versioning strategy for Certificate Monkey, which uses **Conventional Commits** and **Semantic Versioning** to automate releases while maintaining human control.

## Overview

Certificate Monkey uses a **hybrid versioning approach** that combines:

- ✅ **Conventional Commits** for structured commit messages
- ✅ **Semantic Versioning (SemVer)** for version numbers
- ✅ **Automated version calculation** based on commit analysis
- ✅ **Human control** over when releases are created
- ✅ **Automated changelog generation** from commit messages

## Conventional Commits

All commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types

| Type | Description | Version Impact | Example |
|------|-------------|----------------|---------|
| `feat` | New feature | **Minor** bump (0.1.0 → 0.2.0) | `feat: add user authentication` |
| `fix` | Bug fix | **Patch** bump (0.1.0 → 0.1.1) | `fix: resolve memory leak in processing` |
| `docs` | Documentation only | No version bump | `docs: update API documentation` |
| `style` | Code style changes | No version bump | `style: format code with prettier` |
| `refactor` | Code refactoring | No version bump | `refactor: simplify validation logic` |
| `perf` | Performance improvements | No version bump | `perf: optimize certificate parsing` |
| `test` | Adding/updating tests | No version bump | `test: add unit tests for auth service` |
| `build` | Build system changes | No version bump | `build: update dependencies` |
| `ci` | CI/CD changes | No version bump | `ci: add security scanning` |
| `chore` | Maintenance tasks | No version bump | `chore: update gitignore` |
| `revert` | Reverting changes | No version bump | `revert: undo breaking API change` |

### Breaking Changes

Breaking changes trigger a **Major** version bump (0.1.0 → 1.0.0):

```bash
# Method 1: Add ! after type
feat!: redesign API endpoints

# Method 2: Add BREAKING CHANGE in footer
feat: add new authentication system

BREAKING CHANGE: The old API keys are no longer supported.
Users must migrate to the new authentication system.
```

### Examples

```bash
# Feature (minor bump)
feat: add certificate expiration notifications
feat(api): add bulk certificate upload endpoint

# Bug fix (patch bump)
fix: resolve SSL handshake timeout issue
fix(docker): correct environment variable handling

# Breaking change (major bump)
feat!: redesign REST API structure
fix!: change certificate ID format

# No version bump
docs: update installation instructions
test: add integration tests for PFX generation
chore: update dependencies to latest versions
```

## Version Management Commands

### Preview Next Version

See what the next version would be based on commits:

```bash
make version-preview
# or
./scripts/version-manager.sh preview
```

### Bump Version

Automatically calculate and bump version:

```bash
# Automatic bump based on commits
make version-bump-auto

# Manual bump
make version-bump-patch    # 0.1.0 → 0.1.1
make version-bump-minor    # 0.1.0 → 0.2.0
make version-bump-major    # 0.1.0 → 1.0.0
```

### Create Release

Complete release process:

```bash
# Automated release (bump + commit + tag)
make release-auto

# Or step by step
make version-bump-auto     # Update version and changelog
git add . && git commit -m "chore: release v1.2.3"
make version-tag          # Create git tag
git push origin main --tags
```

## Automated Workflows

### Pull Request Validation

Every PR automatically:

1. **Validates commit messages** against conventional commits format
2. **Previews the next version** based on commits in the PR
3. **Comments on the PR** with version impact information

### Release Automation

When commits are pushed to `main`:

1. **Analyzes commits** since the last release
2. **Calculates next version** based on commit types
3. **Automatically creates releases** for `feat` and `fix` commits
4. **Generates changelogs** from commit messages
5. **Creates GitHub releases** with detailed release notes

## Workflow Examples

### Adding a New Feature

```bash
# 1. Create feature branch
git checkout -b feat/user-notifications

# 2. Make changes and commit with conventional format
git commit -m "feat: add email notifications for certificate expiry"

# 3. Create PR - automation will preview version bump
# 4. After PR merge, automation creates minor release
```

### Fixing a Bug

```bash
# 1. Create fix branch
git checkout -b fix/ssl-timeout

# 2. Fix the issue and commit
git commit -m "fix: resolve SSL handshake timeout in certificate validation"

# 3. Create PR - automation will preview patch bump
# 4. After PR merge, automation creates patch release
```

### Making Breaking Changes

```bash
# 1. Create feature branch
git checkout -b feat/api-redesign

# 2. Make breaking changes
git commit -m "feat!: redesign API endpoints for better consistency

BREAKING CHANGE: All API endpoints now use /api/v2/ prefix.
The old /api/v1/ endpoints are deprecated and will be removed in the next major version."

# 3. Create PR - automation will warn about major version bump
# 4. After PR merge, automation creates major release
```

### Documentation Updates

```bash
# Documentation changes don't trigger releases
git commit -m "docs: update installation guide with Docker instructions"
git commit -m "docs: add troubleshooting section to README"
```

## Manual Release Process

If you prefer manual control:

```bash
# 1. Preview what would happen
make version-preview

# 2. Manually bump version
make version-bump-minor

# 3. Review changes
git diff VERSION CHANGELOG.md

# 4. Commit and tag
git add VERSION CHANGELOG.md
git commit -m "chore: release v0.2.0"
make version-tag

# 5. Push
git push origin main --tags
```

## Changelog Management

The changelog is automatically maintained in `CHANGELOG.md`:

- **Automatic generation** from conventional commits
- **Categorized entries** (Features, Bug Fixes, Breaking Changes, etc.)
- **Links to commits** for detailed information
- **Keep a Changelog** format compliance

### Manual Changelog Editing

You can manually edit the changelog after generation:

```bash
# Generate version and changelog
make version-bump-auto

# Edit CHANGELOG.md as needed
vim CHANGELOG.md

# Commit changes
git add CHANGELOG.md
git commit -m "docs: update changelog with additional context"
```

## GitHub Integration

### Release Notes

GitHub releases are automatically created with:

- **Version tag** (e.g., `v1.2.3`)
- **Release title** (e.g., "Release v1.2.3")
- **Detailed release notes** extracted from changelog
- **Commit links** for traceability

### PR Comments

Pull requests automatically receive comments showing:

- **Current version**
- **Next version** (if applicable)
- **Bump type** (major/minor/patch/none)
- **Impact explanation**

## Best Practices

### For Contributors

1. **Use conventional commit format** for all commits
2. **Be descriptive** in commit messages
3. **Use scopes** when helpful (e.g., `feat(api):`, `fix(docker):`)
4. **Mark breaking changes** clearly with `!` or `BREAKING CHANGE:`
5. **Review PR version preview** before merging

### For Maintainers

1. **Review version impact** in PR comments
2. **Ensure changelog accuracy** after releases
3. **Communicate breaking changes** clearly
4. **Use manual releases** for special cases
5. **Monitor automated releases** for issues

## Troubleshooting

### Commit Message Validation Failed

```bash
# Error: Commit message doesn't follow conventional format
# Fix: Use proper format
git commit --amend -m "feat: add proper commit message"
```

### Version Not Bumping

```bash
# Check what commits are being analyzed
make version-preview

# Ensure commits follow conventional format
git log --oneline
```

### Changelog Issues

```bash
# Backup and regenerate changelog
cp CHANGELOG.md CHANGELOG.md.backup
make version-bump-auto
```

### Manual Override

```bash
# Force a specific version bump
make version-bump-major  # Override automatic calculation
```

## Migration from Manual Versioning

If migrating from manual versioning:

1. **Ensure current VERSION file** is accurate
2. **Create initial git tag** if none exists:
   ```bash
   git tag -a v$(cat VERSION) -m "Initial version tag"
   git push origin --tags
   ```
3. **Start using conventional commits** for new changes
4. **Let automation handle** future releases

## Configuration

### Customizing Commit Types

Edit `.github/workflows/conventional-commits.yml` to modify:

- Allowed commit types
- Scope requirements
- Subject patterns

### Customizing Version Calculation

Edit `scripts/version-manager.sh` to modify:

- Bump type logic
- Changelog format
- Tag naming

## Helm Chart Integration

Certificate Monkey is prepared for future Helm chart packaging with:

- ✅ **Helm-compatible versioning** (strict SemVer compliance)
- ✅ **OCI-compliant Docker images** with comprehensive metadata
- ✅ **Multiple image tags** (major, minor, patch) for flexible deployments
- ✅ **Automated release process** ready for chart synchronization

### Docker Image Tags for Helm

Every release creates multiple tags suitable for different Helm deployment strategies:

```yaml
# Production: Pinned version (recommended)
image:
  tag: "0.1.0"  # or "v0.1.0"

# Staging: Auto-receive patches
image:
  tag: "0.1"    # or "v0.1"

# Development: Auto-receive features
image:
  tag: "0"      # or "v0"
```

### Future Chart Versioning

When Helm charts are created:

- **Chart version** will track Helm template changes independently
- **App version** (`appVersion` in Chart.yaml) will always match the application VERSION
- **Synchronization** will be automated in the release workflow
- Charts will be published to OCI registry (GitHub Container Registry)

📖 **[Complete Helm Integration Guide](HELM_INTEGRATION.md)** - Detailed documentation for future chart implementation

### Image Metadata

All Docker images include:

- **OCI labels** for discoverability and Helm compatibility
- **ArtifactHub annotations** for chart repository integration
- **Security scanning** results and SBOM
- **Multi-platform support** (linux/amd64, linux/arm64)

## Support

For questions about the versioning strategy:

1. Check this documentation
2. Run `make commit-help` for quick reference
3. Use `make version-preview` to understand impact
4. See [HELM_INTEGRATION.md](HELM_INTEGRATION.md) for Helm-specific guidance
5. Create an issue for complex scenarios

---

**Remember**: The goal is to make versioning predictable and automated while maintaining the flexibility for human oversight when needed.
