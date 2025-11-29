#!/usr/bin/env bash

set -e  # Exit on error
set -u  # Exit on undefined variable

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function for colored output
success() { echo -e "${GREEN}✓${NC} $1"; }
warning() { echo -e "${YELLOW}⚠${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
info() { echo -e "${BLUE}ℹ${NC} $1"; }

# ============================================================
# PROJECT DETECTION AND DYNAMIC AGENT ASSEMBLY
# ============================================================

# Detect project technology stack based on marker files (scans subdirectories)
detect_project_stack() {
    local target_dir="$1"
    local has_typescript=false
    local has_javascript=false
    local has_go=false
    local has_python=false

    # TypeScript detection (tsconfig.json or package.json with typescript)
    if find "$target_dir" -maxdepth 3 -name "tsconfig.json" -print -quit 2>/dev/null | grep -q .; then
        has_typescript=true
    elif find "$target_dir" -maxdepth 3 -name "package.json" -print 2>/dev/null | head -5 | xargs grep -l '"typescript"' 2>/dev/null | grep -q .; then
        has_typescript=true
    fi

    # JavaScript detection (package.json without typescript)
    if [ "$has_typescript" = false ]; then
        if find "$target_dir" -maxdepth 3 -name "package.json" -print -quit 2>/dev/null | grep -q .; then
            has_javascript=true
        fi
    fi

    # Go detection
    if find "$target_dir" -maxdepth 3 -name "go.mod" -print -quit 2>/dev/null | grep -q .; then
        has_go=true
    fi

    # Python detection
    if find "$target_dir" -maxdepth 3 \( -name "requirements.txt" -o -name "pyproject.toml" -o -name "setup.py" -o -name "Pipfile" \) -print -quit 2>/dev/null | grep -q .; then
        has_python=true
    fi

    # Build comma-separated list of detected stacks
    local detected_stacks=""
    if [ "$has_typescript" = true ]; then
        detected_stacks="typescript"
    elif [ "$has_javascript" = true ]; then
        detected_stacks="javascript"
    fi

    if [ "$has_go" = true ]; then
        if [ -n "$detected_stacks" ]; then
            detected_stacks="$detected_stacks,go"
        else
            detected_stacks="go"
        fi
    fi

    if [ "$has_python" = true ]; then
        if [ -n "$detected_stacks" ]; then
            detected_stacks="$detected_stacks,python"
        else
            detected_stacks="python"
        fi
    fi

    echo "$detected_stacks"
}

# Analyze project with Claude for ambiguous/multi-stack cases
analyze_project_with_claude() {
    local target_dir="$1"
    local detected_stacks="$2"

    info "Analyzing project structure with Claude..." >&2

    # List key source files for analysis (limit to first 30)
    local file_list
    file_list=$(find "$target_dir" -maxdepth 3 -type f \( -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" -o -name "*.py" -o -name "*.go" \) 2>/dev/null | head -30)

    local prompt="Based on these detected technologies: $detected_stacks
And these source files in the project:
$file_list

What is the PRIMARY technology stack for this project? Reply with ONLY one word: typescript, python, go, or javascript. Nothing else."

    local result
    result=$(echo "$prompt" | claude -p 2>/dev/null | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]' | head -1)

    # Validate result is a known stack
    case "$result" in
        typescript|python|go|javascript)
            echo "$result"
            ;;
        *)
            # Default to first detected stack if Claude gives unexpected answer
            echo "$detected_stacks" | cut -d',' -f1
            ;;
    esac
}

# Prompt user to select stack for empty/unknown projects
prompt_user_for_stack() {
    echo "" >&2
    warning "Could not detect project technology stack." >&2
    echo "" >&2
    echo "Please select the primary technology for this project:" >&2
    echo "  [1] TypeScript" >&2
    echo "  [2] Python" >&2
    echo "  [3] Go" >&2
    echo "  [4] JavaScript" >&2
    echo "  [5] All (install all engineer variants)" >&2
    echo "" >&2

    while true; do
        read -p "Your choice (1-5): " choice
        case $choice in
            1) echo "typescript"; return ;;
            2) echo "python"; return ;;
            3) echo "go"; return ;;
            4) echo "javascript"; return ;;
            5) echo "all"; return ;;
            *) error "Invalid choice. Please enter 1-5." >&2 ;;
        esac
    done
}

# Assemble a composed engineer agent from role + skill files
assemble_agent() {
    local stack="$1"
    local output_file="$2"
    local profiles_dir="$3"

    local role_file="$profiles_dir/roles/engineer.md"
    local skill_file="$profiles_dir/skills/${stack}.md"

    # Validate role file exists
    if [ ! -f "$role_file" ]; then
        error "Role file not found: $role_file"
        return 1
    fi

    # Capitalize stack name for display (e.g., typescript -> TypeScript)
    local stack_display
    case "$stack" in
        typescript) stack_display="TypeScript" ;;
        python) stack_display="Python" ;;
        go) stack_display="Go" ;;
        javascript) stack_display="JavaScript" ;;
        *) stack_display="$stack" ;;
    esac

    # Generate YAML frontmatter
    cat > "$output_file" << EOF
---
name: principal-engineer-${stack}
description: Use this agent when you need a Principal ${stack_display} Engineer for code implementation, debugging, refactoring, and development best practices. This agent executes stories by reading execution plans and implementing tasks sequentially with comprehensive testing and documentation lookup.
model: sonnet
color: blue
---

EOF

    # Append role content, replace {Stack} placeholder, and strip HTML comments
    sed "s/{Stack}/${stack_display}/g" "$role_file" | sed '/<!--/,/-->/d' >> "$output_file"

    # Append skill content if it exists (strip HTML comments)
    if [ -f "$skill_file" ]; then
        echo "" >> "$output_file"
        sed '/<!--/,/-->/d' "$skill_file" >> "$output_file"
        info "  Assembled agent with skill: ${stack}"
    else
        warning "  Skill file not found: $skill_file (agent created without skill)"
    fi

    success "  Created composed agent: principal-engineer-${stack}"
}

# Get the directory where this script is located (BMad installation directory)
BMAD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if target path is provided
if [ $# -eq 0 ]; then
    error "Error: Target project path is required"
    echo ""
    echo "Usage: $0 /path/to/your/project"
    echo ""
    echo "Example:"
    echo "  $0 /Users/username/my-project"
    echo ""
    exit 1
fi

TARGET_DIR="$1"

# Validate that target path exists
if [ ! -d "$TARGET_DIR" ]; then
    error "Error: Target directory does not exist: $TARGET_DIR"
    exit 1
fi

# Convert to absolute path
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"

info "BMad Installer"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  BMad Dir:   $BMAD_DIR"
echo "  Target Dir: $TARGET_DIR"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Define BMad source paths (absolute)
BMAD_CLAUDE_SRC="$BMAD_DIR/../.claude"
BMAD_CLAUDE_SRC="$(cd "$(dirname "$BMAD_CLAUDE_SRC")" && pwd)/$(basename "$BMAD_CLAUDE_SRC")"

BMAD_CURSOR_SRC="$BMAD_DIR/../.cursor"
BMAD_CURSOR_SRC="$(cd "$(dirname "$BMAD_CURSOR_SRC")" && pwd)/$(basename "$BMAD_CURSOR_SRC")"

# --- Check .claude in target ---
CLAUDE_PATH="$TARGET_DIR/.claude"
SKIP_CLAUDE_LINK=false

if [ -e "$CLAUDE_PATH" ]; then
    warning "Found existing .claude in target directory"
    echo ""

    # Check if it's already a symlink to our BMad installation
    if [ -L "$CLAUDE_PATH" ]; then
        LINK_TARGET=$(readlink "$CLAUDE_PATH")
        
        if [ "$LINK_TARGET" = "$BMAD_CLAUDE_SRC" ]; then
            success "BMad .claude is already correctly linked."
            SKIP_CLAUDE_LINK=true
        else
            warning "Existing symlink points to: $LINK_TARGET"
        fi
    fi

    if [ "$SKIP_CLAUDE_LINK" = false ]; then
        echo "How would you like to proceed with .claude?"
        echo "  [y] Backup existing .claude to .claude.backup and install BMad"
        echo "  [n] Overwrite existing .claude (no backup) and install BMad"
        echo "  [c] Cancel installation"
        echo ""

        while true; do
            read -p "Your choice (y/n/c): " choice
            case $choice in
                [Yy]* )
                    BACKUP_PATH="$TARGET_DIR/.claude.backup"
                    # Remove old backup if it exists
                    if [ -e "$BACKUP_PATH" ]; then
                        warning "Removing old backup: $BACKUP_PATH"
                        rm -rf "$BACKUP_PATH"
                    fi
                    info "Creating backup: .claude.backup"
                    mv "$CLAUDE_PATH" "$BACKUP_PATH"
                    success "Backup created"
                    break
                    ;;
                [Nn]* )
                    warning "Removing existing .claude without backup"
                    rm -rf "$CLAUDE_PATH"
                    success "Removed existing .claude"
                    break
                    ;;
                [Cc]* )
                    info "Installation cancelled by user"
                    exit 0
                    ;;
                * )
                    error "Invalid choice. Please enter y, n, or c."
                    ;;
            esac
        done
    fi
    echo ""
fi

# --- Check .cursor in target ---
CURSOR_PATH="$TARGET_DIR/.cursor"
SKIP_CURSOR_LINK=false

if [ -e "$CURSOR_PATH" ]; then
    warning "Found existing .cursor in target directory"
    echo ""

    # Check if it's already a symlink to our BMad installation
    if [ -L "$CURSOR_PATH" ]; then
        LINK_TARGET=$(readlink "$CURSOR_PATH")
        
        if [ "$LINK_TARGET" = "$BMAD_CURSOR_SRC" ]; then
            success "BMad .cursor is already correctly linked."
            SKIP_CURSOR_LINK=true
        else
            warning "Existing symlink points to: $LINK_TARGET"
        fi
    fi

    if [ "$SKIP_CURSOR_LINK" = false ]; then
        echo "How would you like to proceed with .cursor?"
        echo "  [y] Backup existing .cursor to .cursor.backup and install BMad"
        echo "  [n] Overwrite existing .cursor (no backup) and install BMad"
        echo "  [s] Skip .cursor installation"
        echo ""

        while true; do
            read -p "Your choice (y/n/s): " choice
            case $choice in
                [Yy]* )
                    BACKUP_PATH="$TARGET_DIR/.cursor.backup"
                    if [ -e "$BACKUP_PATH" ]; then
                         rm -rf "$BACKUP_PATH"
                    fi
                    info "Creating backup: .cursor.backup"
                    mv "$CURSOR_PATH" "$BACKUP_PATH"
                    success "Backup created"
                    break
                    ;;
                [Nn]* )
                    warning "Removing existing .cursor without backup"
                    rm -rf "$CURSOR_PATH"
                    break
                    ;;
                [Ss]* )
                    SKIP_CURSOR_LINK=true
                    info "Skipping .cursor installation"
                    break
                    ;;
                * )
                    error "Invalid choice."
                    ;;
            esac
        done
    fi
    echo ""
fi


# --- Set up BMad .claude directory ---
info "Setting up BMad .claude directory..."

# Create BMad .claude directory if it doesn't exist
if [ ! -d "$BMAD_CLAUDE_SRC" ]; then
    info "Creating BMad .claude directory: $BMAD_CLAUDE_SRC"
    mkdir -p "$BMAD_CLAUDE_SRC"
fi

# Source directory for claudex-go/.claude
CLAUDEX_CLAUDE_DIR="$BMAD_DIR/../claudex-go/.claude"

# Ensure claudex-go/.claude directory exists
if [ ! -d "$CLAUDEX_CLAUDE_DIR" ]; then
    error "Error: Claudex .claude directory not found: $CLAUDEX_CLAUDE_DIR"
    exit 1
fi

# Convert to absolute path
CLAUDEX_CLAUDE_DIR="$(cd "$CLAUDEX_CLAUDE_DIR" && pwd)"

# Create symbolic links for agent profiles
info "Setting up agent profile symbolic links..."

PROFILES_DIR="$BMAD_DIR/../claudex-go/profiles"
AGENTS_DIR="$PROFILES_DIR/agents"

# Ensure profiles directory exists
if [ ! -d "$PROFILES_DIR" ]; then
    error "Error: Profiles directory not found: $PROFILES_DIR"
    exit 1
fi

# Ensure agents directory exists
if [ ! -d "$AGENTS_DIR" ]; then
    error "Error: Agents directory not found: $AGENTS_DIR"
    exit 1
fi

# Convert to absolute paths
PROFILES_DIR="$(cd "$PROFILES_DIR" && pwd)"
AGENTS_DIR="$(cd "$AGENTS_DIR" && pwd)"

# Create symbolic link to the profiles directory itself in .claude
info "Creating profiles directory symlink in .claude..."
PROFILES_LINK_TARGET="$BMAD_CLAUDE_SRC/profiles"
if [ -e "$PROFILES_LINK_TARGET" ] || [ -L "$PROFILES_LINK_TARGET" ]; then
    rm -f "$PROFILES_LINK_TARGET"
fi
ln -s "$PROFILES_DIR" "$PROFILES_LINK_TARGET"

# Populate BMad .claude with symlinks from claudex-go/.claude
info "Populating BMad .claude with symlinks from claudex-go/.claude..."

while IFS= read -r -d '' file; do
    REL_PATH="${file#$CLAUDEX_CLAUDE_DIR/}"
    if [ "$file" = "$CLAUDEX_CLAUDE_DIR" ]; then continue; fi
    TARGET="$BMAD_CLAUDE_SRC/$REL_PATH"
    mkdir -p "$(dirname "$TARGET")"
    if [ -e "$TARGET" ] || [ -L "$TARGET" ]; then rm -f "$TARGET"; fi
    ln -s "$file" "$TARGET"
done < <(find "$CLAUDEX_CLAUDE_DIR" -type f -print0)

success "BMad .claude populated with symlinks"

# Dynamically create symbolic links for all agent profile files in .claude
info "Discovering and linking agent profiles in .claude..."

mkdir -p "$BMAD_CLAUDE_SRC/agents"
mkdir -p "$BMAD_CLAUDE_SRC/commands/agents"

while IFS= read -r -d '' profile_file; do
    PROFILE_NAME="$(basename "$profile_file")"
    if [ -d "$profile_file" ] || [[ "$PROFILE_NAME" == .* ]]; then continue; fi
    
    AGENT_TARGET="$BMAD_CLAUDE_SRC/agents/${PROFILE_NAME}.md"
    COMMAND_TARGET="$BMAD_CLAUDE_SRC/commands/agents/${PROFILE_NAME}.md"
    
    [ -e "$AGENT_TARGET" ] || [ -L "$AGENT_TARGET" ] && rm -f "$AGENT_TARGET"
    [ -e "$COMMAND_TARGET" ] || [ -L "$COMMAND_TARGET" ] && rm -f "$COMMAND_TARGET"
    
    ln -s "$profile_file" "$AGENT_TARGET"
    ln -s "$profile_file" "$COMMAND_TARGET"
    
    info "  Linked profile (claude): $PROFILE_NAME"
done < <(find "$AGENTS_DIR" -maxdepth 1 -type f -print0)

success "Agent profile symbolic links created in .claude"

# ============================================================
# DYNAMIC ENGINEER AGENT ASSEMBLY
# ============================================================

info "Detecting project technology stack..."

DETECTED_STACKS=$(detect_project_stack "$TARGET_DIR")
STACKS_TO_INSTALL=""
PRIMARY_STACK=""

if [ -z "$DETECTED_STACKS" ]; then
    # Empty project - prompt user for stack selection
    PRIMARY_STACK=$(prompt_user_for_stack)
    if [ "$PRIMARY_STACK" = "all" ]; then
        STACKS_TO_INSTALL="typescript,python,go"
    else
        STACKS_TO_INSTALL="$PRIMARY_STACK"
    fi
elif [[ "$DETECTED_STACKS" == *","* ]]; then
    # Multiple stacks detected - install all of them
    info "Multiple technologies detected: $DETECTED_STACKS"
    STACKS_TO_INSTALL="$DETECTED_STACKS"
    # Determine primary for convenience alias using Claude
    PRIMARY_STACK=$(analyze_project_with_claude "$TARGET_DIR" "$DETECTED_STACKS")
    success "Primary stack for alias: $PRIMARY_STACK"
else
    # Single stack detected
    PRIMARY_STACK="$DETECTED_STACKS"
    STACKS_TO_INSTALL="$DETECTED_STACKS"
    success "Detected technology stack: $PRIMARY_STACK"
fi

# Create generated agents directory
GENERATED_DIR="$BMAD_CLAUDE_SRC/generated"
mkdir -p "$GENERATED_DIR"

info "Assembling dynamic engineer agent(s)..."

# Install agent for each detected/selected stack
IFS=',' read -ra STACK_ARRAY <<< "$STACKS_TO_INSTALL"
for stack in "${STACK_ARRAY[@]+"${STACK_ARRAY[@]}"}"; do
    # Trim whitespace
    stack=$(echo "$stack" | tr -d '[:space:]')

    AGENT_FILE="$GENERATED_DIR/principal-engineer-${stack}"
    assemble_agent "$stack" "$AGENT_FILE" "$PROFILES_DIR"

    # Create symlinks for the generated agent
    AGENT_LINK="$BMAD_CLAUDE_SRC/agents/principal-engineer-${stack}.md"
    COMMAND_LINK="$BMAD_CLAUDE_SRC/commands/agents/principal-engineer-${stack}.md"

    [ -e "$AGENT_LINK" ] || [ -L "$AGENT_LINK" ] && rm -f "$AGENT_LINK"
    [ -e "$COMMAND_LINK" ] || [ -L "$COMMAND_LINK" ] && rm -f "$COMMAND_LINK"

    ln -s "$AGENT_FILE" "$AGENT_LINK"
    ln -s "$AGENT_FILE" "$COMMAND_LINK"
done

# Create convenience alias pointing to primary stack
if [ -n "$PRIMARY_STACK" ]; then
    ALIAS_AGENT_LINK="$BMAD_CLAUDE_SRC/agents/principal-engineer.md"
    ALIAS_COMMAND_LINK="$BMAD_CLAUDE_SRC/commands/agents/principal-engineer.md"
    PRIMARY_AGENT_FILE="$GENERATED_DIR/principal-engineer-${PRIMARY_STACK}"

    [ -e "$ALIAS_AGENT_LINK" ] || [ -L "$ALIAS_AGENT_LINK" ] && rm -f "$ALIAS_AGENT_LINK"
    [ -e "$ALIAS_COMMAND_LINK" ] || [ -L "$ALIAS_COMMAND_LINK" ] && rm -f "$ALIAS_COMMAND_LINK"

    ln -s "$PRIMARY_AGENT_FILE" "$ALIAS_AGENT_LINK"
    ln -s "$PRIMARY_AGENT_FILE" "$ALIAS_COMMAND_LINK"
    info "  Created alias principal-engineer → principal-engineer-${PRIMARY_STACK}"
fi

success "Dynamic engineer agent(s) assembled and linked"
echo ""

# --- Set up BMad .cursor directory ---
info "Setting up BMad .cursor directory..."

if [ ! -d "$BMAD_CURSOR_SRC" ]; then
    mkdir -p "$BMAD_CURSOR_SRC"
fi

if [ ! -d "$BMAD_CURSOR_SRC/rules" ]; then
    mkdir -p "$BMAD_CURSOR_SRC/rules"
fi

info "Linking agent profiles to .cursor/rules..."
while IFS= read -r -d '' profile_file; do
    PROFILE_NAME="$(basename "$profile_file")"
    if [ -d "$profile_file" ] || [[ "$PROFILE_NAME" == .* ]]; then continue; fi
    
    # Link with .mdc extension
    RULE_TARGET="$BMAD_CURSOR_SRC/rules/${PROFILE_NAME}.mdc"
    
    [ -e "$RULE_TARGET" ] || [ -L "$RULE_TARGET" ] && rm -f "$RULE_TARGET"
    
    ln -s "$profile_file" "$RULE_TARGET"
    info "  Linked rule: ${PROFILE_NAME}.mdc"
done < <(find "$AGENTS_DIR" -maxdepth 1 -type f -print0)

success "Agent profiles linked to .cursor/rules"


# --- Link target .claude ---
if [ "$SKIP_CLAUDE_LINK" = false ]; then
    info "Creating symlink from target to BMad .claude..."
    ln -s "$BMAD_CLAUDE_SRC" "$CLAUDE_PATH"
    success "Target project .claude symlinked to BMad .claude"
else
    info "Skipping .claude link creation (already linked or skipped)"
fi

# --- Link target .cursor ---
if [ "$SKIP_CURSOR_LINK" = false ]; then
    info "Creating symlink from target to BMad .cursor..."
    ln -s "$BMAD_CURSOR_SRC" "$CURSOR_PATH"
    success "Target project .cursor symlinked to BMad .cursor"
else
    info "Skipping .cursor link creation (already linked or skipped)"
fi


# --- Install binary ---
info "Installing claudex binary to PATH..."

info "Building claudex binary..."
if ! (cd "$BMAD_DIR/../claudex-go" && make build >/dev/null 2>&1); then
    error "Failed to build claudex binary. Make sure Go is installed."
    warning "Skipping binary installation."
else
    success "Claudex binary built successfully"
fi

CLAUDEX_BINARY="$BMAD_DIR/../claudex-go/claudex"
INSTALL_DIR="/usr/local/bin"

if [ ! -f "$CLAUDEX_BINARY" ]; then
    warning "Claudex binary not found: $CLAUDEX_BINARY (skipping)"
else
    if [ -w "$INSTALL_DIR" ]; then
        if [ -e "$INSTALL_DIR/claudex" ] || [ -L "$INSTALL_DIR/claudex" ]; then
            rm -f "$INSTALL_DIR/claudex"
        fi
        ln -s "$CLAUDEX_BINARY" "$INSTALL_DIR/claudex"
        if [ -e "$INSTALL_DIR/.profiles" ] || [ -L "$INSTALL_DIR/.profiles" ]; then
            rm -f "$INSTALL_DIR/.profiles"
        fi
        success "Claudex binary installed to $INSTALL_DIR/claudex"
    else
        echo ""
        warning "Installing claudex requires sudo privileges"
        if sudo -v; then
            if [ -e "$INSTALL_DIR/claudex" ] || [ -L "$INSTALL_DIR/claudex" ]; then
                sudo rm -f "$INSTALL_DIR/claudex"
            fi
            sudo ln -s "$CLAUDEX_BINARY" "$INSTALL_DIR/claudex"
            if [ -e "$INSTALL_DIR/.profiles" ] || [ -L "$INSTALL_DIR/.profiles" ]; then
                sudo rm -f "$INSTALL_DIR/.profiles"
            fi
            success "Claudex binary installed to $INSTALL_DIR/claudex"
        else
            error "Failed to get sudo privileges"
            warning "You can manually add claudex to your PATH by adding this to your shell config:"
            echo "    export PATH=\"\$PATH:$(dirname "$CLAUDEX_BINARY")\""
        fi
    fi
fi

echo ""
success "BMad installed successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  What Was Installed:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✓ BMad .claude directory created/updated"
echo "  ✓ BMad .claude populated with claudex-go/.claude symlinks"
echo "  ✓ Agent profile symlinks created in BMad .claude"
echo "  ✓ Agent profile symlinks created in BMad .cursor/rules"
echo "  ✓ Profiles directory symlinked in BMad .claude"
echo "  ✓ Dynamic engineer agent(s) assembled for: $STACKS_TO_INSTALL"
if [ "$SKIP_CLAUDE_LINK" = false ]; then
    echo "  ✓ Target project .claude symlinked to BMad .claude"
fi
if [ "$SKIP_CURSOR_LINK" = false ]; then
    echo "  ✓ Target project .cursor symlinked to BMad .cursor"
fi
echo "  ✓ Claudex binary installed to /usr/local/bin"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Next Steps:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  1. Navigate to your project:"
echo "     cd $TARGET_DIR"
echo ""
echo "  2. Use claudex command (now available globally):"
echo "     claudex [command] [options]"
echo ""
echo "  3. Or use BMad agents with claude CLI:"
echo "     claude --system-prompt \"\$(cat .claude/commands/agents/team-lead-new.md)\" init"
echo ""
echo "  4. Or use Cursor with the installed rules."
echo ""
