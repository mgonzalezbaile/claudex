# Source Tree Architecture

## Overview
This document describes the source tree structure for the Voiced application project, organized as a monorepo with multiple interconnected components.

## Root Structure

```
voiced/                      # /Users/maikel/Workspace/Pelago/voiced - root project path
├── docs/                    # Documentation and specifications
├── pelago/                  # Main application monorepo
├── simulator/               # Voice development testing tool to QA end-to-end the conversational apps contained in pelago/
```

## Core Directories

### `/docs`
Central documentation hub:
- `architecture/` - Technical architecture documents
- `prd.md` - Product Requirements Document
- `stories/` - Development stories and feature specifications

### `/pelago` (Main Monorepo)
The core application workspace containing:

#### Applications (`/apps`)
- `grain/` - Grain application
- `voiced/` - Voiced application

#### Shared Libraries (`/libs`)
Modular libraries organized by functionality:
- **Authentication**: `auth-client`, `auth-server`, `iap`, `iap-server`
- **Communication**: `llm-client`, `llm-server`, `llm-shared`, `functions-client`
- **Audio/Voice**: `voice`, `stt-server`, `tts-server`, `web2wave`
- **Storage & Data**: `storage`, `query-client`, `backend-utils`
- **UI/UX**: `design-system`, `animation`, `modals`, `navigation`
- **Analytics**: `analytics-client`, `performance-tracker-client`, `error-tracker-client`
- **Utilities**: `date-utils`, `string-helpers`, `array-helpers`, `object-helpers`
- **Configuration**: `eslint-config`, `typescript-config`, `jest-config-native`

#### Shared Components (`/components`)
Reusable UI components with TypeScript support

#### Feature Modules (`/modules`)
Self-contained feature modules:
- `auth-flow` - Authentication workflow
- `chat-list` - Chat interface components
- `error-boundary` - Error handling
- `web-view` - Web view integration
- `account-deletion-web` - Account management

#### Infrastructure (`/infra`)
Deployment and infrastructure configuration:
- `config.json` - Infrastructure settings
- `scripts/` - Deployment scripts
- `templates/` - Infrastructure templates

#### Tooling
- `turbo/` - Monorepo build system configuration
- `scripts/` - Build and utility scripts
- `patches/` - Package patches for dependencies

### `/simulator`
Development and testing utilities:
- API client implementations
- Authentication helpers
- Memory system testing tools
- TypeScript configurations

### `/web-bundles`
Agent system configurations:
- `agents/` - Individual agent definitions (analyst, architect, dev, etc.)
- `expansion-packs/` - Specialized development packs
- `teams/` - Team configuration bundles

## Key Configuration Files

- `turbo.json` - Monorepo build configuration
- `package.json` - Root package dependencies
- `prettier.config.js` - Code formatting rules
- `commitlint.config.ts` - Commit message standards

## Development Workflow

The source tree supports:
1. **Modular Development** - Libraries and modules can be developed independently
2. **Shared Components** - Common UI elements across applications
3. **Type Safety** - Comprehensive TypeScript configuration
4. **Build Optimization** - Turbo-powered monorepo builds
5. **Testing Infrastructure** - Dedicated simulator and testing tools

## Dependencies Management

- Root-level `yarn.lock` for dependency resolution
- Individual `package.json` files for specific packages
- Patch management for third-party dependencies