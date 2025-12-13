# Config Module

Configuration file loading and parsing.

## Key Files
- **config.go** - TOML config parsing for .claudex.toml files

## Key Types
- `Config` - Main configuration struct (doc paths, no_overwrite, features)
- `Features` - Feature toggles for autodoc functionality (session_progress, session_end, frequency)

## Usage

The config module loads .claudex.toml files and provides typed configuration access. Configuration precedence: CLI flags > config file > defaults. Used by App during initialization to configure behavior.
