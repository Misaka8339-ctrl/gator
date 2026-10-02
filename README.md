# Gator

Gator is a local, multi-user RSS aggregator written in Go. Register users, add RSS feeds, follow feeds, collect posts into PostgreSQL, and browse recent posts in your terminal.

This is a learning project built with Go, PostgreSQL, Goose, and SQLC. User switching is local: there is no password-based application authentication.

## Requirements

- Go 1.27 or later for installation from source; also satisfy the version specified in `go.mod`.
- PostgreSQL 15 or later, running and accessible.
- Git to clone the source and obtain migration files.
- Goose to initialize and migrate the database.
- SQLC only when modifying SQL queries or schema and regenerating Go code.

Windows users should run the project inside WSL2 Ubuntu.

## Install

Once this repository is published at the path below, install the CLI with:

```bash
go install github.com/Misaka8339-ctrl/gator/internal/@latest
```

If you already have the source, install from its root directory:

```bash
go install .
```

Go installs executables into `GOBIN` when configured, otherwise into `$(go env GOPATH)/bin`. With the default configuration, add that directory to your shell's PATH:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Add this export to `~/.bashrc` to retain it in new Ubuntu terminals. If using a custom `GOBIN`, add that directory instead.

The examples assume the repository and Go module are named `github.com/Misaka8339-ctrl/gator/internal/`. If hosted elsewhere, replace that path accordingly.

## Set up PostgreSQL

On WSL Ubuntu:

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
psql --version
sudo service postgresql start
sudo -u postgres psql
```

Ensure PostgreSQL is version 15 or later. Inside `psql`, create the database and choose a database password:

```sql
CREATE DATABASE gator;
```

```text
\password postgres
\q
```

If `gator` already exists, reuse it rather than creating it again. The password above belongs to the PostgreSQL role, not the Linux system account.

For this local learning setup, use a connection URL in this format:

```text
postgres://postgres:YOUR_PASSWORD@localhost:5432/gator?sslmode=disable
```

Replace `YOUR_PASSWORD` with your database password. URL-encode reserved characters in credentials. `sslmode=disable` is for this local setup; remote databases may require TLS.

## Apply migrations

Installing the CLI does not create tables or copy migration files into your working directory. Obtain the repository to apply its migrations:

```bash
git clone https://github.com/Misaka8339-ctrl/gator/internal/.git
cd gator
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Set the connection URL in this terminal, then test and migrate:

```bash
export GATOR_DB_URL='postgres://postgres:YOUR_PASSWORD@localhost:5432/gator?sslmode=disable'
psql "$GATOR_DB_URL" -c 'SELECT current_database();'
cd sql/schema
goose postgres "$GATOR_DB_URL" up
cd ../..
```

This creates `users`, `feeds`, `feed_follows`, and `posts`, and adds feed-fetch timestamps. `GATOR_DB_URL` is a shell convenience for these setup commands; the application itself reads its JSON configuration.

## Configure Gator

Create `~/.gatorconfig.json` in your WSL home directory, not the repository directory:

```json
{
  "db_url": "postgres://postgres:YOUR_PASSWORD@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Use the same connection URL as above. Registration and login update `current_user_name` automatically.

```bash
chmod 600 ~/.gatorconfig.json
```

Do not commit real database credentials to Git.

## Quick start

Register a user and add a feed (adding a feed automatically follows it):

```bash
gator register alice
gator addfeed "Wagslane Blog" "https://www.wagslane.dev/index.xml"
gator following
```

Start the collector in one terminal:

```bash
gator agg 1m
```

It fetches one feed immediately and continues on a one-minute ticker, prioritizing feeds never fetched and then the oldest fetch attempts. Keep a single collector running for this implementation. Avoid short polling intervals against third-party sites. Press Ctrl+C to stop.

In a second terminal, browse stored posts:

```bash
gator browse
gator browse 5
```

The default limit is 2. Posts come from feeds followed by the current user and are ordered by publication time, newest first. Unknown publication times appear last.

## Commands

| Command | Purpose |
| --- | --- |
| `gator register <name>` | Create a user and make them current. Duplicate names fail. |
| `gator login <name>` | Switch to an existing user. |
| `gator users` | List users and mark the current user. |
| `gator addfeed "<name>" "<url>"` | Add a unique feed URL and follow it as the current user. |
| `gator feeds` | List all feeds, their URLs, and their creators. |
| `gator follow "<url>"` | Follow a feed already stored in the database. |
| `gator following` | List feeds followed by the current user. |
| `gator unfollow "<url>"` | Remove only the current user's follow relationship. |
| `gator agg <duration>` | Continuously fetch all stored feeds and save posts. Example: `1m`. |
| `gator browse [limit]` | Display recent posts from followed feeds; default limit: 2. |
| `gator reset` | Delete all users and their cascading feeds, follows, and posts. |

`addfeed`, `follow`, `following`, `unfollow`, and `browse` require a current user that exists in the database. `agg` collects all feeds independently of the current user.

`reset` is destructive and intended for development. It preserves the schema but removes business data. It does not clear the username in the JSON config; register a user afterward.

## Data and behavior

- `feeds.user_id` identifies the creator. `feed_follows` stores the many-to-many subscription relationship.
- A user can follow a feed only once. Different users can follow the same feed.
- Posts have globally unique URLs. Repeated collection skips existing URLs without updating existing posts.
- If the same post URL appears in multiple feeds, only its first stored record and feed association are retained.
- Deleting a feed deletes its posts and follow relationships. Deleting a user also deletes feeds they created, even if other users followed those feeds.
- Failed or missing publication dates are stored as NULL; nonempty unrecognized dates are logged.
- The collector marks a feed before fetching, so `last_fetched_at` records an attempt, not guaranteed success.

## Development

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate
go mod tidy
go fmt ./...
go build ./...
```

Goose changes the actual database schema. SQLC reads migrations and queries to generate `internal/database`; it does not apply migrations. Commit generated Go files so users can install without SQLC. Do not edit generated files manually.

Use `go run . <command>` during development. After modifying source, run `go install .` again to refresh the installed `gator` binary. Alternatively:

```bash
go build -o gator .
./gator users
```

Running the compiled executable does not require the Go toolchain, but PostgreSQL and configuration are still required. Builds target a particular OS and architecture, and builds using CGO may depend on system libraries.

## Limitations and troubleshooting

- RSS 2.0 is supported; Atom parsing is not implemented.
- Feed descriptions can contain HTML tags; terminal output does not strip them.
- There is no application password authentication. Anyone with access to this local setup can switch users.
- Feed creation and automatic following currently use separate writes; if following fails after creation, retry with `follow`.
- The collector is sequential and does not coordinate multiple simultaneous collector processes.
- `gator: command not found`: check your Go binary directory and PATH.
- Connection refused: start PostgreSQL and check the host and port.
- Missing table or column: apply all Goose migrations.
- No posts in `browse`: check the current user, their follows, and collector logs.
- Current implementation connects to PostgreSQL before command dispatch, so the database must be running for commands to start.

## References

- [Go installation commands](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)
- [PostgreSQL](https://www.postgresql.org/docs/)
- [Goose](https://github.com/pressly/goose)
- [SQLC](https://docs.sqlc.dev/)
