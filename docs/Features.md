FEAT (features)
feat(auth): registration form & POST
– Branch: feature/auth-register (A)
– AC: asks email/username/password; rejects duplicates; stores bcrypt hash.

feat(auth): login form & POST
– Branch: feature/auth-login (A)
– AC: validates creds; creates session; sets secure HttpOnly cookie.

feat(auth): logout endpoint
– Branch: feature/auth-logout (A)
– AC: deletes session row; clears cookie; redirects home.

feat(session): single active session per user (DB policy)
– Branch: feature/sessions-singleton (A)
– AC: unique index on user_id; newest session takes over; old becomes invalid.

feat(auth): flash messages & form errors
– Branch: feature/flash-errors (B)
– AC: standardised flash partial; shows wrong creds / duplicate warnings.

feat(posts): create post (title, body, categories[])
– Branch: feature/posts-create (B)
– AC: auth-only; validates non-empty; persists join table.

feat(posts): post details page (+ score, categories, comments)
– Branch: feature/posts-show (B)
– AC: shows vote score, author, created_at; comments list.

feat(posts): home feed with filters
– Branch: feature/posts-feed-filters (B)
– AC: ?category=slug, ?my=created, ?my=liked; paginated.

feat(comments): add comment
– Branch: feature/comments-create (B)
– AC: auth-only; non-empty; appears on post page immediately.

feat(votes): like/dislike posts
– Branch: feature/votes-posts (C)
– AC: +1/-1 unique per user+post; toggling and unvoting handled.

feat(votes): like/dislike comments
– Branch: feature/votes-comments (C)
– AC: +1/-1 unique per user+comment; score visible to all.

feat(categories): seed + admin-free list
– Branch: feature/categories-seed (C)
– AC: default categories (General/Go/Web/SQL); shown as checkboxes.

feat(ui): base layout, nav, partials
– Branch: feature/ui-base (D)
– AC: base.tmpl.html, header/footer, flash partial, post card partial.

feat(ui): auth pages (register/login)
– Branch: feature/ui-auth (D)
– AC: accessible forms; server-side validation messages.

feat(ui): create post page
– Branch: feature/ui-post-new (D)
– AC: title/body/checkboxes; server-rendered errors.

feat(ui): error pages (400/401/403/404/500)
– Branch: feature/ui-error-pages (D)
– AC: custom templates; middleware uses them.

feat(http): middleware (recover, logging, auth required)
– Branch: feature/http-middleware (C)
– AC: panic recovery → 500; route logging; gate auth endpoints.

feat(db): connection, pragmas, migration runner
– Branch: feature/db-bootstrap (A)
– AC: open DSN, PRAGMA foreign_keys=ON; run migrations/schema.sql on boot.

feat(security): password hashing (bcrypt)
– Branch: feature/bcrypt-passwords (A)
– AC: cost=12; verify on login.

feat(security): cookie flags
– Branch: feature/cookie-hardening (C)
– AC: HttpOnly, Secure, SameSite=Lax, explicit Expires.

feat(pagination): feed & comments pagination
– Branch: feature/pagination (B)
– AC: limit/offset params; stable ordering.

feat(rate-limit): minimal POST rate-limit by IP (in-mem)
– Branch: feature/rate-limit (C)
– AC: throttles abusive POSTs; returns 429.

feat(ux): small JS enhancement (vote via fetch, no frameworks)
– Branch: feature/ux-vote-fetch (D)
– AC: unobtrusive; server still handles full POST fallback.

feat(scripts): sqlite console & seeding
– Branch: feature/scripts-seed-console (B)
– AC: scripts/db_console.sh, scripts/seed.sh.

feat(docs): ERD & AUDIT_MAP.md
– Branch: feature/docs-erd-audit-map (D)
– AC: ERD diagram + audit cross-reference file.

CHORE (tooling, infra, docs)
chore(repo): init go module, .gitignore, README
– Branch: chore/init-repo (A)

chore(make): Makefile targets (run/build/test/vet/docker)
– Branch: chore/makefile (A)

chore(docker): Dockerfile (multi-stage)
– Branch: chore/dockerfile (C)

chore(compose): docker-compose.yml
– Branch: chore/docker-compose (C)

chore(ci): GitHub Actions (build + go test)
– Branch: chore/ci-gh-actions (B)

chore(git): branch protections & CODEOWNERS
– Branch: chore/git-protection (B)

chore(templates): PULL_REQUEST_TEMPLATE.md
– Branch: chore/pr-template (D)

chore(contrib): CONTRIBUTING.md
– Branch: chore/contributing (D)

chore(config): app config via env (addr, DSN, TTL)
– Branch: chore/config-env (A)

chore(logging): std logger wrapper with request id
– Branch: chore/logger (C)

chore(css): baseline styles, accessible colour contrast
– Branch: chore/css-base (D)

chore(licenses): add MIT licence
– Branch: chore/license (B)

chore/scripts: build/run helper scripts
– Branch: chore/scripts-build-run (B)

FIX (anticipated/typical bugs to address as they appear)
fix(auth): reject empty/whitespace inputs
– Branch: fix/auth-trim-validate (A)
– AC: server-side trimming & min length.

fix(session): stale cookie after new login
– Branch: fix/session-stale-cookie (A)
– AC: second browser loses validity immediately.

fix(posts): prevent empty post/comment body
– Branch: fix/empty-content-guard (B)

fix(votes): prevent like & dislike simultaneously
– Branch: fix/votes-exclusive (C)
– AC: DB PK enforces; handler handles toggles clearly.

fix(sqlite): handle “database is locked” (busy timeout)
– Branch: fix/sqlite-busy-timeout (A)
– AC: set _busy_timeout or PRAGMA busy_timeout.

fix(security): basic output escaping in templates
– Branch: fix/xss-escape (D)

fix(http): correct status codes (400/401/403/404/500)
– Branch: fix/http-statuses (C)

fix(ui): preserve form values on validation errors
– Branch: fix/ui-form-sticky (D)

fix(pagination): off-by-one / negative page params
– Branch: fix/pagination-params (B)

fix(errors): centralised error renderer
– Branch: fix/error-renderer (C)

TEST (unit/integration/manual)
test(db/users): Create/Find/Exists duplicate
– Branch: test/db-users (A)

test(db/sessions): Replace semantics & expiry
– Branch: test/db-sessions (A)

test/db/votes: unique constraints & toggling
– Branch: test/db-votes (C)

test/db/posts: create with categories + list by category
– Branch: test/db-posts-categories (B)

test/handlers/auth: register/login/logout happy & unhappy paths
– Branch: test/handlers-auth (A)

test/handlers/posts: create/empty/unauthorised
– Branch: test/handlers-posts (B)

test/handlers/comments: create/empty/unauthorised
– Branch: test/handlers-comments (B)

test/handlers/votes: like/dislike/flip/unauthorised
– Branch: test/handlers-votes (C)

test/middleware: recover/logging/auth-required
– Branch: test/middleware (C)

test/integration: end-to-end basic flows with httptest.Server
– Branch: test/integration-e2e (D)

test/ui: template rendering smoke tests (parse/execute)
– Branch: test/ui-templates (D)

test/scripts: docker build/run script smoke test (CI)
– Branch: test/scripts-docker (B)