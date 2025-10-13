git checkout develop
git pull origin develop
git switch -c feature/auth-register
# code...
git add -A
git commit -m "feat(auth): register with email+username+bcrypt"
git push -u origin feature/auth-register
# open PR: base=develop, reviewers assigned


## branch strategy (simple, professional)
Branches

main — always deployable.

feature/<short-scope> — new features (e.g., feature/auth-register).

fix/<short-scope> — bug fixes.

chore/<short-scope> — tooling/ci/docs.

release/<version> — (optional) cut tagged releases.

# Protection rules (recommended)

main: require PR, linear history, 1 review, status checks (build + test).

commit message convention (Conventional Commits)
Examples:

feat(auth): register with email+username+bcrypt

fix(votes): prevent like and dislike simultaneously

chore(docker): add compose file

test(repo): add users repo tests

Scope suggestions: auth, posts, comments, votes, db, http, ui, docker, infra, docs.

PR process & guidelines (CONTRIBUTING.md essentials)

**Branch off main, there's no develop branch for this project.**

Keep PRs small (<300 lines delta where possible).

Include:

What/Why, screenshots (UI), and Audit Map references (e.g., “Authentication: A1, A2”).

Test coverage for new repos/handlers.

Checklists in PR:

 make lint test passes

 SQL statements reviewed (CREATE/INSERT/SELECT present where needed)

 HTTP status codes correct (400/401/403/404/500)

 Single-session behaviour manually verified