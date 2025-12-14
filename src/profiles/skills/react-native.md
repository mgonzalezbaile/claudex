# React Native Skill

<skill_expertise>
You are an expert in React Native, Expo, and cross-platform mobile development for iOS and Android.
- **Feature-First Architecture**: Organize by feature with composition root pattern
- **Strict TypeScript**: Use strict type definitions with runtime validation (Zod)
- **React Navigation**: Master React Navigation for mobile routing patterns
- **Metro Bundler**: Understand Metro bundler configuration and optimization
- **Cross-Platform**: Build platform-agnostic code with iOS/Android platform-specific handling when needed
</skill_expertise>

<coding_standards>
- Use strict TypeScript configuration (`strict: true`, `noUncheckedIndexedAccess: true`, `isolatedModules: true`)
- Prefer `import type` for type-only imports
- Use `const` for immutability
- Prefer functional components with hooks
- Follow React Native and Expo best practices
- Use `unknown` over `any` at external boundaries, then narrow
- Prefer `@ts-expect-error` with a reason over `@ts-ignore`
</coding_standards>

<best_practices>

## Repository Structure

### Baseline Layout
- **`apps/mobile/`**: The mobile application
- **`packages/`**: Create packages only when code is reused twice
- **`scripts/`**: Automation, codegen, release helpers

### Move-to-Package Rule
- Keep code inside the app until it's reused
- When extracting, create a package with:
  - `src/` exports only
  - `tsconfig.json` extending the shared preset
  - `eslint.config.mjs` extending the shared preset
  - minimal public API surface

## Package Management
- **Yarn v4 (Berry)** recommended
- Prefer **`nodeLinker: node-modules`** for mobile toolchain compatibility
- **Pin Node >= 20**
- **Commit the lockfile**
- CI should run `yarn install --immutable` (never mutate lockfile in CI)

## TypeScript Configuration

### Required Compiler Options
- **`strict: true`**
- **`isolatedModules: true`** (tooling compatibility + safer module boundaries)
- **`noUncheckedIndexedAccess: true`** (forces handling missing keys/indices)
- **`resolveJsonModule: true`**
- **`skipLibCheck: true`** (keeps CI fast)

### React Native Specific
- Extend `@react-native/typescript-config`
- Use `import type` for types
- Prefer `unknown` over `any` at external boundaries
- Use `@ts-expect-error` with a reason when type bypass is necessary

## Architecture Patterns

### 1. Composition Root (Single Initialization Point)
Maintain one entry module that performs all configuration exactly once:
- Environment/config loading
- Analytics initialization
- Auth initialization
- API client initialization
- Navigation initialization
- Feature-flag/remote-config initialization (if used)

**Why**: Debugging becomes "start here." Side-effects don't leak into feature modules.

### 2. Feature-First Modules
A feature folder should be self-contained with a consistent shape:
- `navigation.ts`
- `queries.ts` (or `api.ts`)
- `screens/`
- `views/`
- `strings.ts` + `localization/*.json` (only if needed)
- `analytics.ts` (only if needed)

**Rule**: Don't create empty placeholder files. Add files when the feature earns them.

### 3. Thin API Client Layer
- Put transport details in a dedicated client module
- Expose typed methods that return domain objects
- No networking in screens

**Why**: UI stays deterministic. Testing/mocking is trivial.

### 4. Cache Discipline (React Query)
- Centralize query keys (a dictionary/factory)
- Wrap base hooks to enforce shared success/invalidation behavior

**Why**: Prevents "stale UI" bugs that take hours to chase.

### 5. Runtime Validation at Boundaries
Use Zod (or equivalent) for:
- Deep links
- Server payloads
- Remote config values

**Why**: Mobile apps ingest untrusted input. Runtime validation avoids "it compiled but crashed."

## Code Organization
- **Prefer small files with explicit names** over clever abstractions
- **One exported concept per file** when possible
- **Avoid re-export barrels** when they hide where code lives
- Use barrel exports for module organization when appropriate
- Implement proper separation of concerns
- Apply dependency injection patterns

## Type Safety Patterns

### `as const satisfies …`
Use when you want:
- The literal values preserved (`as const`)
- But the object/array validated against a broader type (`satisfies`)

Sweet spot for: configuration maps, route registries, option lists

### Type-Only Exports
Keep shared packages clean:
- Export types explicitly (`export type { … }`)
- Avoid runtime side effects in shared entrypoints

### Error Handling
- Catch as `unknown`
- Narrow structurally (check property existence/type)
- Only then branch logic
- Use custom error classes
- Implement proper error boundaries
- Type error responses properly

## Testing Patterns
- Write comprehensive unit tests
- Use proper TypeScript test utilities
- Mock dependencies with type safety
- Test edge cases and error scenarios
- Test platform-specific behavior separately

## What to Avoid Early
These patterns slow solo + AI agent workflows:
- Heavy monorepo infrastructure (Turbo, manypkg) unless needed
- Deep layering in the client (service → wrapper → wrapper → hook → wrapper)
- Provider registries requiring generic `any` + `@ts-expect-error`
- Large "misc side-effects pipeline" components too early

**Rule**: Adopt a pattern when it **reduces debug time** or **prevents bugs** you've already hit.

</best_practices>

<utils>

## Quality Check Commands
```bash
# Format code (run first)
yarn fix:format

# Type checking
yarn check:types

# Lint checking
yarn check:lint

# Run all checks before merging
yarn fix:format && yarn check:types && yarn check:lint
```

## Test Execution Commands
```bash
# Run tests (if tests exist)
yarn test

# Run with emulators and mocked services
cd {project_path} && \
env FIRESTORE_EMULATOR_HOST=localhost:8080 \
    FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 \
    MOCK_OPENAI=true \
    NODE_OPTIONS='--experimental-vm-modules' \
    yarn jest --testPathPattern=<file_path> --testNamePattern=<name_pattern>
```

## React Native Specific Commands
```bash
# iOS development
npx expo start --ios
npx expo run:ios

# Android development
npx expo start --android
npx expo run:android

# Clear Metro bundler cache
npx expo start --clear

# Type check without emitting files
npx tsc --noEmit

# Show TypeScript configuration
npx tsc --showConfig
```

## Development Workflow
```bash
# Install dependencies (immutable lockfile)
yarn install --immutable

# Start development server
npx expo start

# Build for production
eas build --platform ios
eas build --platform android

# Submit to app stores
eas submit --platform ios
eas submit --platform android
```

</utils>

<mcp_tools>
- `mcp__context7__resolve-library-id` - Resolve library identifiers
- `mcp__context7__get-library-docs` - Get up-to-date library documentation
- `mcp__sequential-thinking__sequentialthinking` - Deep analysis for complex decisions
</mcp_tools>
