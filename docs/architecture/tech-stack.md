# Tech Stack

## Overview
This document outlines the technology stack for the Voiced application project, organized as a monorepo with multiple interconnected components.

## Core Technologies

### Runtime Environment
- **Node.js LTS** - Server-side JavaScript runtime
- **TypeScript** - Type-safe JavaScript with strict configuration
- **ESM Modules** - Modern ECMAScript module system

### Frontend Technologies
- **React** - UI framework for web applications
- **React Native** - Mobile application development
- **TypeScript** - Type safety across all frontend code

### Backend Technologies
- **Node.js** - Server runtime environment
- **Express/Fastify** - HTTP server frameworks
- **TypeScript** - Server-side type safety

### Audio & Voice Processing
- **Speech-to-Text (STT)** - Voice input processing
- **Text-to-Speech (TTS)** - Voice output generation
- **Web Audio API** - Browser-based audio processing
- **Voice Processing Libraries** - Custom voice handling

### AI & Language Models
- **LLM Integration** - Large Language Model clients and servers
- **Function Calling** - AI-driven function execution
- **Chat Interface** - Conversational AI components

### Authentication & Security
- **OAuth/IAP** - Identity and Access Platform
- **JWT Tokens** - Secure authentication tokens
- **Session Management** - User session handling

### Data & Storage
- **Database** - Data persistence layer
- **Query Clients** - Type-safe database queries
- **Storage APIs** - File and data storage

### Development Tools

#### Build System
- **Turbo** - Monorepo build orchestration
- **TypeScript Compiler** - Type checking and compilation
- **ESBuild/TSUP** - Fast bundling for production

#### Code Quality
- **ESLint** - TypeScript-specific linting rules
- **Prettier** - Code formatting
- **Commitlint** - Commit message standards

#### Testing
- **Jest** - Testing framework with native configuration
- **Integration Tests** - Module interaction testing
- **E2E Testing** - End-to-end application testing

#### Package Management
- **Yarn** - Package manager with workspace support
- **Patch Package** - Third-party dependency patches

### Infrastructure

#### Deployment
- **Infrastructure as Code** - Configuration-driven deployments
- **Deployment Scripts** - Automated deployment workflows
- **Environment Configuration** - Multi-environment support

#### Monitoring & Analytics
- **Performance Tracking** - Application performance monitoring
- **Error Tracking** - Error reporting and debugging
- **Analytics** - User behavior and application metrics

### Development Workflow

#### Monorepo Architecture
- **Shared Libraries** - Reusable code across applications
- **Module Federation** - Independent module development
- **Type Sharing** - Consistent types across packages

#### Code Organization
- **Feature Modules** - Self-contained functionality
- **Design System** - Consistent UI components
- **Utility Libraries** - Shared helper functions

### Key Dependencies

#### Core Libraries
- **Date Utilities** - Date manipulation and formatting
- **String/Array/Object Helpers** - Data manipulation utilities
- **Animation Libraries** - UI animations and transitions
- **Navigation** - Application routing and navigation
- **Modals** - UI overlay components

#### Configuration
- **ESLint Config** - Shared linting configuration
- **TypeScript Config** - Shared type configuration
- **Jest Config** - Testing configuration for React Native

## Architecture Principles

### Type Safety
- Strict TypeScript configuration
- No `any` types without justification
- Branded types for domain identifiers
- Discriminated unions for state management

### Error Handling
- Explicit error types and handling
- Result types for expected failures
- Structured error boundaries
- Consistent error reporting

### Performance
- Lazy loading for heavy modules
- Iterator helpers for streaming
- Optimized build pipeline
- Resource management with modern TypeScript features

### Security
- Input validation at boundaries
- Parameterized queries
- Secure defaults for HTTP headers
- Environment-based configuration management

## Version Compatibility

- **Node.js**: LTS versions
- **TypeScript**: Latest stable
- **React**: Current stable
- **React Native**: Current stable

## Development Standards

All code follows the established coding standards documented in `coding-standards.md`, with emphasis on:
- Type safety and explicit boundaries
- Modular architecture with clear dependencies
- Comprehensive testing strategy
- Security-first development practices