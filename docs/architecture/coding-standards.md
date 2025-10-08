# Principal TypeScript Engineer – Backend Coding Standards

These standards are written for an LLM acting as a Principal TypeScript Engineer on enterprise Node.js backends. Follow them precisely unless the user explicitly asks to deviate.

## 0) Golden rules

* Target Node.js LTS with modern ESM. Prefer ECMAScript 2022+ features.
* Enable strictest type safety. No `any` unless fully justified and isolated.
* Keep public APIs explicit and stable. Backwards compatibility matters.
* Validate all untrusted inputs at the boundary and convert to typed domain objects.
* Make invalid states unrepresentable with types. Prefer discriminated unions and branded types.
* Handle errors explicitly and consistently. Never throw non-`Error` values.
* Prefer small, composable modules with clear boundaries.
* Tests first for critical paths. Types catch classes of bugs, tests cover behavior.

---

## 2) Linting and formatting

* **ESLint** with `@typescript-eslint` recommended + strict. Key rules:

  * `@typescript-eslint/array-type`: array simple style.
  * `@typescript-eslint/consistent-type-definitions`: prefer `type` for unions, `interface` for object contracts.
  * `@typescript-eslint/explicit-module-boundary-types`: error on missing public function types.
  * `@typescript-eslint/no-explicit-any`: error. Allow in a single, isolated wrapper with comment.
  * `@typescript-eslint/no-floating-promises`: error.
  * `@typescript-eslint/no-unnecessary-type-assertion`: error.
  * `@typescript-eslint/prefer-nullish-coalescing`, `prefer-optional-chain`: warn.
  * `@typescript-eslint/switch-exhaustiveness-check`: error.
  * `no-restricted-imports`: enforce module boundaries.
  * `no-throw-literal`: error.
* **Prettier** for formatting. No custom styles beyond defaults.

---

## 3) Module, file, and naming conventions

* One top-level concept per file. Max 300 lines where feasible.
* File names: `kebab-case.ts`. Class names: `PascalCase`. Types and interfaces: `PascalCase`. Constants: `SCREAMING_SNAKE_CASE`.
* Public exports at the end or via a barrel `index.ts`. Avoid wide barrels that leak internals.
* Do not use default exports in libraries. Named exports improve refactors and tree shaking.

---

## 4) Runtime environment and packaging

* Use ESM. Avoid mixing `require` and `import`.
* Node flags for source maps in dev: run with `node --enable-source-maps`.
* For runtime: target native features of current LTS. Avoid polyfills unless necessary.
* For serverless, bundle with `esbuild` or `tsup`. Always run `tsc --noEmit` in CI for type checks.

---

## 5) Error handling

### 5.1 Principles

* Throw `Error` or subclasses only.
* Preserve causality using `new Error(message, { cause })`.
* Never swallow errors. Either handle and recover, or rethrow.
* Catch at boundaries: HTTP framework error middleware, job runners, message handlers.

### 5.2 Error taxonomy

```ts
export class DomainError extends Error { readonly kind = "DomainError" as const; }
export class ValidationError extends DomainError { readonly kind = "ValidationError" as const; }
export class NotFoundError extends DomainError { readonly kind = "NotFoundError" as const; }
export class PermissionError extends DomainError { readonly kind = "PermissionError" as const; }

export function ensureError(e: unknown): Error {
  return e instanceof Error ? e : new Error(String(e));
}
```

### 5.3 Result types for expected failures

```ts
type Ok<T> = { ok: true; value: T };
type Err<E extends string = string> = { ok: false; error: E; details?: unknown };
export type Result<T, E extends string = string> = Ok<T> | Err<E>;

export const ok = <T>(value: T): Ok<T> => ({ ok: true, value });
export const err = <E extends string>(error: E, details?: unknown): Err<E> => ({ ok: false, error, details });
```

Use exceptions for exceptional, Result for expected domain outcomes. Do not mix within the same path.

---

## 6) API and boundary validation

* All external inputs must be validated at the edge (HTTP, queue, cron, CLI).
* Use a schema library (e.g. Zod) to validate and transform into domain types.

```ts
import { z } from "zod";

export const CreateUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1),
});
export type CreateUser = z.infer<typeof CreateUserSchema>;

export function parseCreateUser(body: unknown): CreateUser {
  return CreateUserSchema.parse(body);
}
```

* After parsing, the rest of the system uses typed values. No re-checking inside.

---

## 7) Type system guidelines

### 7.1 Prefer unions and discriminants

```ts
type ApiResponse<T> =
  | { status: "success"; data: T }
  | { status: "error"; error: string; code?: number };

function handle<T>(r: ApiResponse<T>) {
  switch (r.status) {
    case "success": return r.data;
    case "error":   throw new Error(r.error);
    default:        const _exhaustive: never = r; return _exhaustive;
  }
}
```

### 7.2 Branded identifiers

```ts
type Brand<T, B extends string> = T & { readonly __brand: B };
export type UserId = Brand<string, "UserId">;

const asUserId = (s: string): UserId => s as UserId;
```

Do not confuse branded ids with plain strings.

### 7.3 Generics and utility types

* Use `Partial<T>` for patches, `Readonly<T>` for immutability, `Omit<T, K>` for DTOs.
* Avoid deep `any`. Prefer precise generics.

```ts
interface Repository<T, K> {
  findById(id: K): Promise<T | null>;
  save(entity: T): Promise<T>;
}
```

### 7.4 Avoid assertions

* Prefer type guards and narrowing.

```ts
function isDate(x: unknown): x is Date { return x instanceof Date; }
```

### 7.5 Catch variable typed as unknown

```ts
try {
  // ...
} catch (e: unknown) {
  const err = ensureError(e);
  logger.error({ err }, "Unhandled error");
  throw err;
}
```

### 7.6 Resource management with `using` (TS 5.2+)

```ts
class Db implements Disposable {
  [Symbol.dispose]() { this.close(); }
  close() {/*...*/}
}

using db = new Db();
// use db; auto-closed at end of scope
```

---

## 8) Concurrency and async

* Prefer `async/await`. No floating promises.
* For parallelism, use `Promise.allSettled` with typed arrays. Avoid unbounded concurrency.

```ts
const results = await Promise.allSettled(tasks.map(runTask));
const successes = results.filter((r): r is PromiseFulfilledResult<Value> => r.status === "fulfilled").map(r => r.value);
```

* Do not block the event loop with CPU-heavy work. Offload to worker threads or a job system.

---

## 9) Data access

* Prefer type-safe clients:

  * SQL: Prisma, Drizzle, Kysely. No stringly-typed SQL scattered across the app.
  * Mongo: supply collection generics `Collection<UserDoc>`.
* Centralize queries in repositories. Domain services do not embed raw queries.
* Map DB records to domain models. Do not leak DB nullability and column names through layers.

---

## 10) Logging and observability

* Use structured logging (pino or winston) with typed context.

```ts
interface LogContext { reqId: string; userId?: UserId; }
logger.info(<LogContext>{ reqId, userId }, "User updated");
```

* Correlate logs with request id. Instrument with OpenTelemetry if available.
* Never log secrets or PII. Create a `redact` list for loggers.

---

## 11) Security

* All secrets from env or vault. Validate config at startup using a schema.
* Safe defaults: HTTP security headers, CSRF where applicable, rate limiting.
* Parameterized queries. No string concatenation in SQL.
* Serialize errors for clients without leaking internals.

---

## 12) Testing strategy

* **Unit** for pure logic and type-level helpers. Fast and isolated.
* **Integration** for modules working together (HTTP handlers to DB with a real or ephemeral instance).
* **E2E** for the deployed artifact or a full local stack.
* Generate fixtures via factories typed by domain types. No `any` in tests.

```ts
describe("createUser", () => {
  it("creates a user", async () => {
    const input: CreateUser = { email: "a@b.com", name: "Alice" };
    const res = await api.createUser(input);
    expect(res.status).toBe("success");
  });
});
```

* Enforce `tsc --noEmit` in CI. Lint and tests must pass.

---

## 13) API layer standards

* Versioned routes: `/v1/...`.
* Request and response types live in `@acme/shared-types` to share with clients when feasible.
* Use discriminated unions for response envelopes. Never return raw DB rows.
* Consistent error mapping:

```ts
function toHttpError(e: Error) {
  if (e instanceof ValidationError) return { status: 400, body: { error: e.message } };
  if (e instanceof NotFoundError)   return { status: 404, body: { error: e.message } };
  return { status: 500, body: { error: "Internal Server Error" } };
}
```

---

## 14) Architectural boundaries

* Layers: `api` (transport) -> `service` (domain) -> `repository` (data) -> `infra` (adapters).
* Upward dependencies only. Lower layers must not import higher layers.
* Enforce with ESLint `no-restricted-imports` and folder tags.

---

## 15) Performance and startup

* Lazy load heavy modules if rarely used. Consider `import()`.
* Avoid global singletons with heavy initialization at import time.
* Use iterator helpers for streaming transformations where possible.

---

## 16) Documentation and comments

* Document public types and functions with JSDoc. Explain non-obvious invariants and constraints.
* Avoid redundant comments. Code and types should be clear by themselves.
* For complex types, include an example in comments.

---

## 19) Do and do not checklist

**Do**

* Validate inputs at the edge using schemas and convert to domain types.
* Use discriminated unions and exhaustive switches.
* Use branded ids for key identity types.
* Use `unknown` in catch, convert to `Error`.
* Keep public function signatures explicitly typed.

**Do not**

* Use `any` except in a local adapter with justification.
* Throw strings or numbers.
* Leak DB shapes or nullability into domain and API layers.
* Silence the compiler with `as` without a guard.
* Mix CJS and ESM in the same package.

---

## 20) Snippets you can reuse

### 20.1 Exhaustiveness helper

```ts
export function assertNever(x: never): never {
  throw new Error(`Unexpected variant: ${String(x)}`);
}
```

### 20.2 Async handler wrapper

```ts
export function wrap<A extends unknown[], R>(
  fn: (...a: A) => Promise<R>
) {
  return async (...a: A) => {
    try { return await fn(...a); }
    catch (e: unknown) {
      const err = ensureError(e);
      // attach tracing or context if needed
      throw err;
    }
  };
}
```

### 20.3 Config validation at startup

```ts
const EnvSchema = z.object({
  NODE_ENV: z.enum(["development", "test", "production"]),
  PORT: z.coerce.number().int().min(1),
  DATABASE_URL: z.string().url(),
});
export type Env = z.infer<typeof EnvSchema>;
export const env: Env = EnvSchema.parse(process.env);
```

---

## 21) Use cases and command handlers

Use cases implement business logic following the Command pattern. They encapsulate single business operations and coordinate between domain services and repositories.

### 21.1 Structure and naming

```ts
export interface CreateUserCommand {
  email: string;
  name: string;
}

export class CreateUserCommandHandler {
  constructor(
    private readonly userRepository: UserRepositoryI,
    private readonly emailService: EmailServiceI
  ) {}

  async execute(command: CreateUserCommand): Promise<void> {
    // Implementation
  }
}
```

### 21.2 Standards

* **File naming**: `kebab-case.ts` matching the use case name (e.g., `create-user.ts`)
* **Command interface**: Define explicit command interface with all required parameters
* **Handler class**: Single responsibility, named `{UseCase}CommandHandler`
* **Execute method**: Always named `execute`, takes command, returns `Promise<T>` or `Promise<void>`
* **Dependencies**: Inject repositories and services via constructor, use interface types
* **Error handling**: Throw domain errors for business rule violations
* **No direct HTTP/framework dependencies**: Use cases are transport-agnostic

### 21.3 Testing

* Test the `execute` method with mocked dependencies
* Cover happy path and error scenarios
* Use typed command objects, never `any`

```ts
describe("CreateUserCommandHandler", () => {
  it("creates user successfully", async () => {
    const command: CreateUserCommand = { email: "test@example.com", name: "Test" };
    const handler = new CreateUserCommandHandler(mockUserRepo, mockEmailService);
    
    await handler.execute(command);
    
    expect(mockUserRepo.save).toHaveBeenCalledWith(expect.objectContaining({
      email: "test@example.com"
    }));
  });
});
```

---

## 22) Policies and event handlers

Policies handle cross-cutting concerns and side effects triggered by domain events. They coordinate between use cases and external systems.

### 22.1 Structure and naming

```ts
export const conversationStarted = onMessagePublished(
  { topic: PubSubMessageTopic.ConversationStarted, memory: '512MiB' },
  async event => {
    logger.info('conversationStarted: message published.', {
      structuredData: true,
      data: event.data
    });

    const messageJson = event.data.message.json;
    const { sessionId, userId } = ConversationStartedEvent.parse(messageJson);

    // Coordinate use cases
    const handler = new EndActivePendingConversationsCommandHandler(repository);
    await handler.execute({ userId });

    logger.info('conversationStarted: completed', {
      structuredData: true,
      data: { sessionId, userId }
    });
  }
);
```

### 22.2 Standards

* **File naming**: `{event-name}-policy.ts` (e.g., `conversation-started-policy.ts`)
* **Export pattern**: Export the cloud function directly, use descriptive names
* **Event parsing**: Always validate event payloads with schemas (Zod)
* **Logging**: Log start and completion with structured data
* **Error handling**: Let framework handle retries, log errors with context
* **Use case coordination**: Instantiate and call use case handlers, don't duplicate logic
* **Memory allocation**: Specify appropriate memory limits for cloud functions

### 22.3 Barrel exports

Use `index.ts` to re-export all policies for clean imports:

```ts
export * from './conversation-started-policy';
export * from './user-registered-policy';
```

### 22.4 Testing

* Test event parsing and validation
* Mock use case handlers and verify they're called correctly
* Test error scenarios and logging

---

## 23) Feature readiness for modern TS

* Use `using` for resource cleanup where it simplifies code.
* Prefer const type parameters when preserving literal types through helpers.
* Rely on improved narrowing. Write code that the compiler can understand without casts.
* Enable latest lib features aligned with Node LTS.

---
