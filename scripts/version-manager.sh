#!/bin/bash

# Certificate Monkey Version Manager
# Handles semantic versioning based on conventional commits

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CURRENT_VERSION_FILE="VERSION"
CHANGELOG_FILE="CHANGELOG.md"

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Get current version
get_current_version() {
    if [ -f "$CURRENT_VERSION_FILE" ]; then
        cat "$CURRENT_VERSION_FILE" | tr -d '\n'
    else
        echo "0.1.0"
    fi
}

# Get last git tag
get_last_tag() {
    git describe --tags --abbrev=0 2>/dev/null || echo ""
}

# Analyze commits since last tag/version
analyze_commits() {
    local since_ref="$1"
    local has_breaking=false
    local has_feat=false
    local has_fix=false

    if [ -z "$since_ref" ]; then
        # If no tag exists, analyze all commits
        commits=$(git rev-list --reverse HEAD)
    else
        # Analyze commits since last tag
        commits=$(git rev-list --reverse "${since_ref}..HEAD" 2>/dev/null || git rev-list --reverse HEAD)
    fi

    for commit in $commits; do
        message=$(git log --format=%s -n 1 "$commit")
        body=$(git log --format=%b -n 1 "$commit")

        # Check for breaking changes
        if echo "$message" | grep -q "!:" || echo "$body" | grep -q "BREAKING CHANGE"; then
            has_breaking=true
        elif echo "$message" | grep -q "^feat"; then
            has_feat=true
        elif echo "$message" | grep -q "^fix"; then
            has_fix=true
        fi
    done

    # Return bump type
    if [ "$has_breaking" = true ]; then
        echo "major"
    elif [ "$has_feat" = true ]; then
        echo "minor"
    elif [ "$has_fix" = true ]; then
        echo "patch"
    else
        echo "none"
    fi
}

# Calculate next version
calculate_next_version() {
    local current_version="$1"
    local bump_type="$2"

    IFS='.' read -r major minor patch <<< "$current_version"

    case "$bump_type" in
        major)
            echo "$((major + 1)).0.0"
            ;;
        minor)
            echo "$major.$((minor + 1)).0"
            ;;
        patch)
            echo "$major.$minor.$((patch + 1))"
            ;;
        none)
            echo "$current_version"
            ;;
        *)
            log_error "Invalid bump type: $bump_type"
            exit 1
            ;;
    esac
}

# Generate changelog entry
generate_changelog_entry() {
    local version="$1"
    local since_ref="$2"
    local date=$(date +%Y-%m-%d)

    echo "## [$version] - $date"
    echo ""

    # Get commits since last tag/version
    if [ -z "$since_ref" ]; then
        commits=$(git rev-list --reverse HEAD)
    else
        commits=$(git rev-list --reverse "${since_ref}..HEAD" 2>/dev/null || git rev-list --reverse HEAD)
    fi

    # Categorize commits
    local features=()
    local fixes=()
    local breaking=()
    local docs=()
    local chores=()
    local others=()

    for commit in $commits; do
        message=$(git log --format=%s -n 1 "$commit")
        body=$(git log --format=%b -n 1 "$commit")
        short_hash=$(git log --format=%h -n 1 "$commit")

        # Check for breaking changes first
        if echo "$message" | grep -q "!:" || echo "$body" | grep -q "BREAKING CHANGE"; then
            breaking+=("- $message ([${short_hash}](../../commit/${commit}))")
        elif echo "$message" | grep -q "^feat"; then
            features+=("- $message ([${short_hash}](../../commit/${commit}))")
        elif echo "$message" | grep -q "^fix"; then
            fixes+=("- $message ([${short_hash}](../../commit/${commit}))")
        elif echo "$message" | grep -q "^docs"; then
            docs+=("- $message ([${short_hash}](../../commit/${commit}))")
        elif echo "$message" | grep -q "^chore\|^build\|^ci\|^style\|^refactor\|^perf\|^test"; then
            chores+=("- $message ([${short_hash}](../../commit/${commit}))")
        else
            others+=("- $message ([${short_hash}](../../commit/${commit}))")
        fi
    done

    # Output sections
    if [ ${#breaking[@]} -gt 0 ]; then
        echo "### 💥 BREAKING CHANGES"
        echo ""
        printf '%s\n' "${breaking[@]}"
        echo ""
    fi

    if [ ${#features[@]} -gt 0 ]; then
        echo "### ✨ Features"
        echo ""
        printf '%s\n' "${features[@]}"
        echo ""
    fi

    if [ ${#fixes[@]} -gt 0 ]; then
        echo "### 🐛 Bug Fixes"
        echo ""
        printf '%s\n' "${fixes[@]}"
        echo ""
    fi

    if [ ${#docs[@]} -gt 0 ]; then
        echo "### 📚 Documentation"
        echo ""
        printf '%s\n' "${docs[@]}"
        echo ""
    fi

    if [ ${#chores[@]} -gt 0 ]; then
        echo "### 🔧 Maintenance"
        echo ""
        printf '%s\n' "${chores[@]}"
        echo ""
    fi

    if [ ${#others[@]} -gt 0 ]; then
        echo "### 📦 Other Changes"
        echo ""
        printf '%s\n' "${others[@]}"
        echo ""
    fi
}

# Update changelog
update_changelog() {
    local version="$1"
    local since_ref="$2"

    log_info "Updating changelog for version $version..."

    # Create backup
    cp "$CHANGELOG_FILE" "${CHANGELOG_FILE}.backup"

    # Generate new entry
    local new_entry=$(generate_changelog_entry "$version" "$since_ref")

    # Create temporary file with new content
    {
        # Keep header until [Unreleased] section
        sed -n '1,/## \[Unreleased\]/p' "$CHANGELOG_FILE"

        # Add empty unreleased section
        echo ""
        echo "### Added"
        echo "- TBD"
        echo ""
        echo "### Changed"
        echo "- TBD"
        echo ""
        echo "### Deprecated"
        echo "- TBD"
        echo ""
        echo "### Removed"
        echo "- TBD"
        echo ""
        echo "### Fixed"
        echo "- TBD"
        echo ""
        echo "### Security"
        echo "- TBD"
        echo ""

        # Add new version entry
        echo "$new_entry"

        # Add rest of changelog (skip unreleased section)
        sed -n '/## \[Unreleased\]/,$p' "$CHANGELOG_FILE" | tail -n +2 | sed '/^### /,/^$/d' | sed '/^$/N;/^\n$/d'

    } > "${CHANGELOG_FILE}.tmp"

    mv "${CHANGELOG_FILE}.tmp" "$CHANGELOG_FILE"

    log_success "Changelog updated successfully"
}

# Update version file
update_version_file() {
    local version="$1"
    echo "$version" > "$CURRENT_VERSION_FILE"
    log_success "Version file updated to $version"
}

# Create git tag
create_git_tag() {
    local version="$1"
    local tag_name="v$version"

    log_info "Creating git tag $tag_name..."

    # Create annotated tag with changelog entry
    local tag_message="Release $version

$(generate_changelog_entry "$version" "$(get_last_tag)")"

    git tag -a "$tag_name" -m "$tag_message"
    log_success "Git tag $tag_name created"
}

# Preview Docker tags for a version
preview_docker_tags() {
    local version="$1"

    IFS='.' read -r major minor patch <<< "$version"

    echo "🐳 Docker Tags for v${version}"
    echo "=============================="
    echo ""
    echo "Semantic Version Tags:"
    echo "  - ${version}             (patch-pinned)"
    echo "  - v${version}            (patch-pinned with v)"
    echo "  - ${major}.${minor}      (minor-tracking)"
    echo "  - v${major}.${minor}     (minor-tracking with v)"
    echo "  - ${major}               (major-tracking)"
    echo "  - v${major}              (major-tracking with v)"
    echo ""
    echo "Special Tags:"
    echo "  - latest                 (if from main branch)"
    echo "  - main-\$(git-sha)       (immutable reference)"
    echo ""
    echo "Usage Examples:"
    echo "  Production:  docker pull ghcr.io/org/certificate-monkey:${version}"
    echo "  Staging:     docker pull ghcr.io/org/certificate-monkey:${major}.${minor}"
    echo "  Development: docker pull ghcr.io/org/certificate-monkey:${major}"
    echo ""
    echo "Helm values.yaml:"
    echo "  image:"
    echo "    tag: \"${version}\"  # Recommended for production"
    echo ""
}

# Validate version is Docker/Helm compatible
validate_docker_version() {
    local version="$1"

    # Must be valid SemVer
    if ! echo "$version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
        log_error "Version $version is not valid SemVer"
        log_error "Docker/Helm require format: MAJOR.MINOR.PATCH (e.g., 1.2.3)"
        return 1
    fi

    log_success "Version $version is Docker/Helm compatible"
    return 0
}

# Main commands
cmd_preview() {
    log_info "Analyzing commits for version preview..."

    local current_version=$(get_current_version)
    local last_tag=$(get_last_tag)
    local bump_type=$(analyze_commits "$last_tag")
    local next_version=$(calculate_next_version "$current_version" "$bump_type")

    echo ""
    echo "📋 Version Analysis"
    echo "==================="
    echo "Current version: $current_version"
    echo "Last git tag: ${last_tag:-"(none)"}"
    echo "Bump type: $bump_type"
    echo "Next version: $next_version"
    echo ""

    if [ "$bump_type" != "none" ]; then
        echo "🔮 Preview of changelog entry:"
        echo "=============================="
        generate_changelog_entry "$next_version" "$last_tag"
        echo ""
        echo ""
        preview_docker_tags "$next_version"
    else
        log_warning "No version bump needed (no feat/fix commits found)"
    fi
}

cmd_bump() {
    local bump_type="$1"

    if [ -z "$bump_type" ]; then
        log_error "Bump type required. Usage: $0 bump <major|minor|patch|auto>"
        exit 1
    fi

    local current_version=$(get_current_version)
    local last_tag=$(get_last_tag)

    if [ "$bump_type" = "auto" ]; then
        bump_type=$(analyze_commits "$last_tag")
        if [ "$bump_type" = "none" ]; then
            log_warning "No version bump needed (no feat/fix commits found)"
            exit 0
        fi
    fi

    local next_version=$(calculate_next_version "$current_version" "$bump_type")

    # Validate Docker/Helm compatibility
    if ! validate_docker_version "$next_version"; then
        exit 1
    fi

    log_info "Bumping version from $current_version to $next_version ($bump_type)"

    # Update files
    update_version_file "$next_version"
    update_changelog "$next_version" "$last_tag"

    log_success "Version bumped successfully!"
    echo ""
    preview_docker_tags "$next_version"
    echo ""
    echo "Next steps:"
    echo "1. Review the updated CHANGELOG.md"
    echo "2. Commit changes: git add . && git commit -m 'chore: release v$next_version'"
    echo "3. Create tag: $0 tag"
    echo "4. Push: git push origin main --tags"
}

cmd_tag() {
    local current_version=$(get_current_version)
    create_git_tag "$current_version"
}

cmd_release() {
    log_info "Creating complete release..."

    local current_version=$(get_current_version)
    local last_tag=$(get_last_tag)
    local bump_type=$(analyze_commits "$last_tag")

    if [ "$bump_type" = "none" ]; then
        log_warning "No version bump needed (no feat/fix commits found)"
        exit 0
    fi

    local next_version=$(calculate_next_version "$current_version" "$bump_type")

    log_info "Creating release $next_version ($bump_type bump)"

    # Update files
    update_version_file "$next_version"
    update_changelog "$next_version" "$last_tag"

    # Commit changes
    git add "$CURRENT_VERSION_FILE" "$CHANGELOG_FILE"
    git commit -m "chore: release v$next_version"

    # Create tag
    create_git_tag "$next_version"

    log_success "Release v$next_version created successfully!"
    echo ""
    echo "To publish the release:"
    echo "git push origin main --tags"
}

# Command to preview Docker tags
cmd_docker_tags() {
    local version="${1:-$(get_current_version)}"

    if ! validate_docker_version "$version"; then
        exit 1
    fi

    preview_docker_tags "$version"
}

# Help
cmd_help() {
    echo "Certificate Monkey Version Manager"
    echo "=================================="
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  preview              Preview next version based on commits"
    echo "  bump <type>          Bump version (major|minor|patch|auto)"
    echo "  tag                  Create git tag for current version"
    echo "  release              Complete release process (bump + commit + tag)"
    echo "  docker-tags [ver]    Preview Docker tags for version (default: current)"
    echo "  help                 Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 preview           # Preview what the next version would be"
    echo "  $0 bump auto         # Automatically determine and bump version"
    echo "  $0 bump minor        # Force a minor version bump"
    echo "  $0 release           # Complete release process"
    echo "  $0 docker-tags       # Show Docker tags for current version"
    echo "  $0 docker-tags 1.2.3 # Show Docker tags for specific version"
    echo ""
}

# Main script
case "${1:-help}" in
    preview)
        cmd_preview
        ;;
    bump)
        cmd_bump "$2"
        ;;
    tag)
        cmd_tag
        ;;
    release)
        cmd_release
        ;;
    docker-tags)
        cmd_docker_tags "$2"
        ;;
    help|--help|-h)
        cmd_help
        ;;
    *)
        log_error "Unknown command: $1"
        echo ""
        cmd_help
        exit 1
        ;;
esac
