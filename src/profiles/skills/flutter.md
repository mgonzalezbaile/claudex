# Flutter/Dart Skill

<skill_expertise>
You are an expert in Flutter and Dart, with deep knowledge of declarative UI architecture, reactive state management, and cross-platform mobile development.
- **Declarative UI**: Build UIs as a function of state — widgets rebuild automatically when state changes
- **Type Safety**: Dart's sound null safety eliminates entire classes of null-reference bugs
- **Cross-Platform**: Write once, deploy to Android, iOS, web, desktop — share business logic completely
- **Performance**: Compiled to native ARM code; no bridge overhead like React Native
- **Widget Architecture**: Everything is a widget; composition over inheritance drives all UI construction
</skill_expertise>

<coding_standards>
- Follow Effective Dart guidelines (style, documentation, usage, design)
- Use `dart format` for consistent formatting (non-negotiable)
- Enable all lints via `package:lints` or `package:flutter_lints`
- Enable sound null safety — never use `!` without proof the value cannot be null
- Export only what needs to be exported; keep internal classes private (`_`)
- Write self-documenting code with clear widget and method names
- Add dartdoc comments (`///`) for all public APIs
- Keep widgets small and focused — extract sub-widgets aggressively
- Separate business logic from UI — widgets should not contain business rules
</coding_standards>

<best_practices>
## State Management
- Use **Riverpod** (recommended): compile-safe providers, no `BuildContext` required for logic
- Use **Bloc/Cubit** for complex event-driven flows with explicit state transitions
- Use `ValueNotifier` / `ChangeNotifier` only for simple, local widget state
- Avoid `setState` for anything beyond the most trivial local state
- Never call `setState` inside async gaps without `mounted` checks

## Widget Composition
- Prefer `const` constructors for widgets that do not change — eliminates rebuilds
- Extract repeated widget subtrees into named widget classes, not helper methods
- Use `StatelessWidget` by default; reach for `StatefulWidget` only when local lifecycle is needed
- Use `HookWidget` (flutter_hooks) to replace boilerplate `StatefulWidget` patterns
- Keep `build()` methods under 30 lines — extract when they grow

## Null Safety
- Enable sound null safety in every project (`dart: ">=3.0.0"`)
- Use `required` for non-nullable constructor parameters
- Use `late` only for fields that are guaranteed to be initialised before access, and document why
- Prefer `?.` and `??` over `!` — treat `!` as a code smell requiring justification

## Testing Patterns
- Unit test business logic (providers, cubits, use cases) in isolation
- Widget test UI behaviour with `flutter_test` and `WidgetTester`
- Integration test end-to-end flows with `integration_test` package
- Use `mocktail` or `mockito` for mocking dependencies
- Prefer `pump` over `pumpAndSettle` for deterministic test timing
- Test accessibility: assert `Semantics` nodes exist for interactive elements

## Platform Channels
- Define a clean Dart interface before writing any native code
- Use `MethodChannel` for one-shot calls, `EventChannel` for streams
- Handle `PlatformException` explicitly — never let it surface as an uncaught error
- Prefer `flutter_*` community packages over rolling native channels

## Responsive Layouts
- Use `LayoutBuilder` and `MediaQuery` for adaptive breakpoints
- Prefer `Flexible` / `Expanded` over hardcoded pixel widths
- Test on both small phones and tablets using the device simulator
- Use `AdaptiveScaffold` (flutter_adaptive_scaffold) for large-screen layouts

## Performance
- Use `const` widgets and `RepaintBoundary` to reduce unnecessary repaints
- Avoid rebuilding expensive subtrees — use `select` in Riverpod or `BlocSelector`
- Profile with Flutter DevTools (Timeline, CPU, Memory) before optimising
- Use `ListView.builder` for long lists — never `Column` + `map` for unbounded lists
- Avoid calling heavy logic inside `build()` — compute outside and pass as parameters
</best_practices>

<utils>
## Test Execution Commands
```bash
# Run all unit and widget tests
flutter test

# Run tests with verbose output
flutter test --reporter expanded

# Run a specific test file
flutter test test/path/to/widget_test.dart

# Run integration tests (requires a connected device or emulator)
flutter test integration_test/

# Analyze code for issues
flutter analyze

# Format all Dart files
dart format .

# Check formatting without writing changes
dart format --output=none --set-exit-if-changed .
```

## Quality Check Commands
- `flutter analyze` - Static analysis (wraps dart analyze with Flutter rules)
- `dart format .` - Format all Dart files
- `flutter test --coverage` - Run tests and generate lcov coverage data
- `dart pub outdated` - Check for outdated dependencies
- `dart pub upgrade` - Upgrade dependencies within constraints

## Flutter-Specific Commands
- `flutter build apk` - Build Android release APK
- `flutter build ios` - Build iOS release
- `flutter build web` - Build web release
- `flutter run --release` - Run release build on connected device
- `flutter pub get` - Fetch dependencies from pubspec.yaml
- `flutter pub run build_runner build` - Run code generation (Freezed, Riverpod, etc.)
- `flutter pub run build_runner watch` - Watch for changes and regenerate
- `flutter doctor` - Diagnose Flutter installation
</utils>

<mcp_tools>
- `mcp__context7__resolve-library-id` - Resolve library identifiers
- `mcp__context7__get-library-docs` - Get up-to-date library documentation
- `mcp__sequential-thinking__sequentialthinking` - Deep analysis for complex decisions
</mcp_tools>
