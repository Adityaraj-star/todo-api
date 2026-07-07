# Todo API (Go)

A REST API for managing todos, built from scratch in Go while learning backend development. Started as a simple in-memory CRUD app and grew into something closer to a real service: PostgreSQL for persistence, JWT auth so todos are private per user, and tests that actually run against the whole thing.

---

## What it does

- Register an account, log in, get a JWT back
- Create, read, update, delete todos — but only your own
- Filter todos by status
- Paginate results with `limit` and `offset`
- Every request gets logged (method, path, time taken)
- Data survives restarts (Postgres, not a map in memory)

---

## Tech stack

- **Go** — `net/http` under the hood, `chi` for routing
- **PostgreSQL** — `database/sql` + `lib/pq`, raw SQL (no ORM)
- **JWT** (`golang-jwt/jwt`) — stateless auth
- **bcrypt** (`golang.org/x/crypto`) — password hashing
- **Docker Compose** — spins up Postgres locally with one command
- **Go's built-in testing package** + `httptest` — integration tests against the real router

---

## Architecture

The core idea running through this project: **handlers don't know or care where data is stored.** They only depend on interfaces (`TodoStore`, `UserStore`). Two things implement those interfaces — a Postgres-backed version and an in-memory version — and both can be swapped in without touching a single line of HTTP code.

```
Request
   │
   ▼
chi router  ──▶  middleware.Logger  ──▶  middleware.RequireAuth (for /todos/*)
                                              │
                                              ▼
                                      handler (todo.go / auth.go)
                                              │
                                              ▼
                                    store.TodoStore / store.UserStore   ◀── interfaces
                                        │                    │
                                        ▼                    ▼
                              PostgresTodoStore      MemoryTodoStore
                              PostgresUserStore      MemoryUserStore
                                        │
                                        ▼
                                   PostgreSQL
```

Why this matters in practice: tests run the entire router against `MemoryTodoStore`/`MemoryUserStore` — no database needed, no test containers, no cleanup between runs. Production runs the same router against Postgres. Same handlers, same middleware, because neither the handler nor the router import a concrete store type — only the interface in `store/store.go` and `store/user_store.go`.

### Folder structure

```
todo-api/
├── main.go                  # wires everything together, starts the server
├── go.mod / go.sum
├── docker-compose.yml        # local Postgres
├── .env.example
│
├── auth/
│   ├── jwt.go                # generate/parse JWTs, refuses to start without JWT_SECRET
│   └── password.go           # bcrypt hash/compare
│
├── middleware/
│   ├── logger.go              # logs every request
│   └── auth.go                # validates Bearer token, puts user id in request context
│
├── model/
│   ├── todo.go                # Todo struct
│   └── user.go                # User struct
│
├── store/
│   ├── store.go               # TodoStore interface + ListParams (filter/pagination)
│   ├── user_store.go          # UserStore interface
│   ├── postgres_todo.go       # real DB implementation
│   ├── postgres_user.go       # real DB implementation
│   ├── memory_todo.go         # in-memory implementation, used by tests
│   └── memory_user.go         # in-memory implementation, used by tests
│
├── handler/
│   ├── todo.go                 # CRUD endpoints, all scoped to the logged-in user
│   ├── auth.go                 # register / login
│   └── handler_test.go         # integration tests using httptest + memory stores
│
└── db/
    └── db.go                   # connects to Postgres, creates tables if missing
```

---

## How auth actually works here

1. `POST /register` — takes username + password, hashes the password with bcrypt (never stored in plain text), saves the user, returns the user's id and username (not the hash, not a token).
2. `POST /login` — looks up the user, compares the password against the stored hash. If it matches, generates a JWT containing the user's id, signed with `JWT_SECRET`, valid for 24 hours.
3. Every request to `/todos/*` goes through `middleware.RequireAuth`, which:
   - reads the `Authorization: Bearer <token>` header
   - parses and validates the JWT
   - if valid, pulls the user id out of the token and stores it in the request context
   - if missing or invalid, returns `401` before the request ever reaches a handler
4. Handlers read the user id back out of context (`middleware.UserIDFromContext`) and pass it into every store call. Every query — list, get, update, delete — is scoped by `user_id` at the SQL level (`WHERE user_id = $1`), not just filtered in Go after the fact. That distinction matters: even if a handler had a bug, the database itself won't return another user's row.

One deliberate choice: login returns the exact same error message ("invalid username or password") whether the username doesn't exist or the password is wrong. That's on purpose — it stops someone from using the login endpoint to figure out which usernames are registered.

Another: the server **won't start** if `JWT_SECRET` isn't set in the environment. No fallback default. A backend that silently signs tokens with a guessable secret is worse than one that just refuses to boot.

---

## Request flow example: creating a todo

```
POST /todos/  (Authorization: Bearer <token>)
   │
   ▼
chi matches route → middleware.Logger wraps it → middleware.RequireAuth runs
   │
   ├─ no/invalid token → 401, stop here
   │
   ▼ (token valid)
handler.Create
   │
   ├─ decode JSON body
   ├─ validate title isn't empty → 400 if it is
   ├─ build a Todo{} with a new UUID and the user id from context
   ▼
store.Create(todo)
   │
   ▼
PostgresTodoStore.Create → INSERT INTO todos (...)
   │
   ▼
201 Created + the todo as JSON
```

---

## Running it

### With Postgres (default, matches production setup)

```bash
git clone https://github.com/Adityaraj-star/todo-api.git
cd todo-api
cp .env.example .env      # then edit JWT_SECRET to something real
docker compose up -d      # starts Postgres on :5432
export $(cat .env | xargs)
go run main.go
```

Tables are created automatically on startup if they don't exist (`db.Migrate` in `db/db.go`) — no separate migration step needed for this project's scale.

### Without Postgres (in-memory, zero setup)

```bash
JWT_SECRET=dev-secret STORE=memory go run main.go
```

Useful for quickly poking at the API without touching Docker.

Server runs on `http://localhost:8080`.

---

## API reference

All `/todos/*` routes require `Authorization: Bearer <token>` from `/login`.

**POST /register**
```json
{ "username": "aditya", "password": "secret123" }
```
→ `201` `{ "id": "...", "username": "aditya" }` or `409` if the username's taken.

**POST /login**
```json
{ "username": "aditya", "password": "secret123" }
```
→ `200` `{ "token": "..." }` or `401` on bad credentials.

**GET /todos/** — list your todos
Query params: `?status=in-progress&limit=10&offset=0`

**GET /todos/{id}** — get one todo (404 if it's not yours or doesn't exist)

**POST /todos/**
```json
{ "title": "learn go" }
```

**PUT /todos/{id}**
```json
{ "title": "learn go properly", "status": "in-progress" }
```

**DELETE /todos/{id}** → `204` on success

---

## Testing

```bash
go test ./... -v
```

Tests spin up the real chi router (same wiring as `main.go`) against in-memory stores, so they run fast with no external dependencies. Covered:

- duplicate username on register
- wrong password on login
- hitting `/todos` with no token
- full CRUD lifecycle (create → get → update → delete → confirm 404 after delete)
- that one user genuinely cannot read another user's todo, even by guessing/knowing the id

---

## Design decisions worth explaining out loud

**Why an interface for storage instead of just using `*sql.DB` directly in handlers?**
Two reasons: testability (no database needed for tests) and the ability to swap backends later without a rewrite. It also just forces cleaner separation between "how do I store this" and "how do I respond to this HTTP request" — those are genuinely different concerns.

**Why raw SQL instead of an ORM?**
Wanted to actually understand what queries were running, connection handling, and how parameterized queries prevent SQL injection, instead of having an ORM abstract all of that away.

**Why chi instead of just `net/http`'s default mux?**
Path parameters (`{id}`) and middleware chaining (`r.Route`, `r.Use`) are much cleaner with chi. The stdlib mux needed manual string-splitting on the URL path to get an id out, which is fragile.

**Why does `GetAll`/`GetByID`/etc. take `userID` as a parameter instead of the store just trusting whatever handler calls it?**
So the authorization boundary lives at the data layer, not just in HTTP-handling code. Even a future handler that forgot to check ownership couldn't accidentally leak another user's todo, because the query itself is scoped.

---

## What I'd add next

- A real migration tool (golang-migrate/goose) instead of `CREATE TABLE IF NOT EXISTS` — the current approach can't handle altering an existing table, only creating new ones
- Rate limiting on `/login` and `/register`
- Token refresh / logout (revocation) — right now a JWT is valid for its full 24h no matter what
- More edge-case tests: malformed JSON bodies, expired tokens, negative pagination values
- Deploy it somewhere with a live link instead of just running locally

---

## Author

Aditya Raj