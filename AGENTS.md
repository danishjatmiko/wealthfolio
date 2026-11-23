# AGENTS.md

Instructions for coding agents (Codex, Claude Code, etc.) working in this repo.
For what the code does, read [`DOCUMENTATION.md`](DOCUMENTATION.md) and [`README.md`](README.md).

## Ground rules

- Use commit author and committer identity **`Danish Jatmiko <danishkho19@gmail.com>`**
  in this repository. GitHub authentication does not set Git commit identity.
- **Always ask the user before every `git commit`, `git push`, and production deploy.**
  One "yes" covers only that one action, not later ones. Show what you are about to
  commit/deploy (the diff summary or commit list) when you ask.
- Never force-push, never rewrite `main`'s history.
- Never commit `.env`, `keystore.properties`, `*.jks`, or anything under `backups/`.

## Checks before shipping

Run whatever applies to what you changed, and fix failures before asking to deploy:

```bash
cd backend && go build ./... && go test ./...
cd frontend && npm run build && npm run lint
```

## Deploying to production

Production is **https://etherna.id**, a Docker Compose project (`db`, `api`, `web`)
on the VPS `ubuntu@43.134.230.167`, checked out at `/opt/wealthfolio`.
There are no PRs and no CI deploy: releasing = push to `origin/main`, then run
the deploy script on the VPS.

`git push` and `ssh` need network access. If your sandbox blocks the network,
request approval to run these commands outside the sandbox rather than skipping them.
SSH uses the key already set up on this Mac — never ask the user for a password.

### 1. Commit (after the user says yes)

Commit on your branch with a clear message describing the user-visible change.

### 2. Push to `origin/main` (after the user says yes)

The `origin` repo belongs to the GitHub account **`danishjatmiko`**.
Use that account when pushing this repository, then restore whichever account
was active beforehand so other repositories keep their existing authentication:

```bash
previous_account="$(gh api user --jq .login)"
gh auth switch --user danishjatmiko && git push origin HEAD:main
push_status=$?
gh auth switch --user "$previous_account"
(exit "$push_status")
```

If the push is rejected as non-fast-forward, `git fetch origin main`, rebase your
work onto `origin/main`, re-run the checks, and ask again. Do not force-push.

### 3. Run the deploy on the VPS (after the user says yes)

```bash
ssh ubuntu@43.134.230.167 'cd /opt/wealthfolio && bash deploy.sh'
```

`deploy.sh` (lives only on the VPS, not in git) does, in order:
1. backs up the database to `/opt/wealthfolio/backups/wealthfolio_<timestamp>.pgdump`,
2. `git fetch` + `git merge --ff-only origin/main`,
3. `docker compose up -d --build api web` (DB migrations run when the API starts),
4. fails if the API logs contain `error|fatal|panic`,
5. fails if `https://etherna.id` does not return 200,
6. prints the restore-point dump filename.

The build takes a few minutes; give the command a long timeout (10 min).

### 4. Report

Tell the user: which commit(s) went live, whether `deploy.sh` finished with
"Deploy complete", and the restore-point filename it printed.
If it exited non-zero, show the user the failing output and **do not** attempt a
rollback without asking.

### Rollback (only when the user asks)

Code only (no bad migration):

```bash
ssh ubuntu@43.134.230.167 'cd /opt/wealthfolio && git reset --hard <previous-good-commit> && docker compose up -d --build api web'
```

Then fix forward on `main` (push a `git revert` commit), otherwise the next
`deploy.sh` fast-forwards straight back onto the bad code.

Database restore from the dump `deploy.sh` printed (destroys data written since then — confirm with the user first):

```bash
ssh ubuntu@43.134.230.167 'cd /opt/wealthfolio && docker exec -i wealthfolio-db-1 pg_restore -U wealthfolio -d wealthfolio --clean --if-exists < backups/<dump-file>.pgdump'
```

### Don't touch on the VPS

- The `stack` compose project (unrelated jusjeruk MySQL/Redis) runs on the same VPS — leave it alone.
- Untracked, host-specific files in `/opt/wealthfolio` are intentional: `.env`, `backup.sh`,
  `deploy.sh`, `caddy/`, `docker-compose.override.yml`, `rates-sync.env`, `backups/`.
  Never `git clean` there or edit them unless the user asks.
