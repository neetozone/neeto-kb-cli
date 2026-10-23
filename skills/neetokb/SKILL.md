---
name: neetokb
description: >
  Manage NeetoKB from the command line.
  Use when the user asks about operations exposed by the NeetoKB CLI.
---

## Prerequisites

Run `neetokb doctor` to check authentication and connectivity.
If not authenticated, run `neetokb login`.

## Authentication & multi-subdomain

Credentials for every logged-in subdomain are stored together in
`~/.config/neetokb/auth.json`. A command that talks to the API picks which
subdomain to use by these rules:

- 0 subdomains authenticated → every credential-using command errors with
  "Not authenticated. Run 'neetokb login' to authenticate.".
- 1 subdomain authenticated → that one is the implicit default; `--subdomain`
  may be omitted.
- 2+ subdomains authenticated → **`--subdomain <name>` is required** on every
  credential-using command, including `doctor`. The error lists every
  authenticated subdomain so the agent can offer a choice.

`login` / `logout` / `whoami` have dedicated behavior:

| Command | Behavior |
|---|---|
| `neetokb login --subdomain <name>` | Adds or refreshes the entry for `<name>`. No flag → prompts for the subdomain. |
| `neetokb logout --subdomain <name>` | Removes that one entry. |
| `neetokb logout --all` | Removes every entry. |
| `neetokb logout` (no flag) | Removes the only entry if exactly one is logged in; errors if multiple. |
| `neetokb whoami` | Lists every logged-in account. Marks the entry `(default)` when exactly one. |
| `neetokb whoami --subdomain <name>` | Shows just that one. |

## Global flags (persistent on every command)

| Flag | Purpose |
|---|---|
| `--subdomain <name>` | Select which logged-in subdomain the command targets. Required when multiple are logged in. |
| `--json` | Force JSON envelope output even on a TTY. |
| `--quiet` | Emit only the raw payload — no envelope, no breadcrumbs. For action commands (create/update), emits just the resource identifier; `delete` emits `success`. Designed for scripting. |
| `--toon` | Emit TOON (Token Optimized Output Notation). Preferred for feeding list/show output back to an LLM; ~30–60% fewer tokens than JSON. |

Precedence if multiple are set: `--toon` > `--quiet` > `--json` > pretty.

## Output modes & response envelope

**Pretty (default on a TTY)** — tables for arrays, key-value for objects,
breadcrumbs appended. Not intended for machine consumption.

**JSON envelope** (non-TTY, or `--json`):
```json
{
  "data": <resource body>,
  "breadcrumbs": [{ "label": "List", "command": "neetokb <resource> list" }],
  "pagination": {
    "current_page_number": 1,
    "total_pages": 10,
    "total_records": 250
  }
}
```
`breadcrumbs` is omitted when empty. `pagination` is present only for list
commands.

**Quiet** (`--quiet`) — `data` contents only, no envelope. For action
commands `PrintQuiet` unwraps a single-key wrapper and prints the first of
`sid` / `id` / `name`. For `delete` it prints `success`.

**TOON** (`--toon`) — same data as JSON, re-encoded into TOON. Shape is
equivalent but whitespace/keys are compressed. Parse by re-reading keys as
you would JSON.

### Pagination

List commands accept `--page` (1-indexed) and `--page-size` (max 100).
The envelope's `pagination` field always exposes:
`current_page_number`, `total_pages`, `total_records`. Agents should loop
by incrementing `--page` until `current_page_number == total_pages`.

## Discovery

The full, always-accurate command tree (including any flags added after
this skill was built) is available as JSON:

```bash
neetokb commands
```

Each catalog entry has `command`, `description`, optional `flags` (with
`name`, `type`, `default`, `description`, `required`), and `subcommands`.
Use this whenever a user asks about a flag or command not covered below.

## Diagnostics & IDE setup

| Command | Purpose |
|---|---|
| `doctor` | Auth check + API reachability + version. Uses `--subdomain` when multiple are logged in. |
| `version` | Print CLI version / commit / build date. |
| `commands` | Emit the full command/flag catalog as JSON. |
| `setup claude` | Install NeetoKB plugin into Claude Code (`plugin.json`, hooks, this SKILL.md). |
| `setup cursor` / `windsurf` / `copilot` / `gemini` / `codex` | Write NeetoKB rule files into the current project directory; re-run after an upgrade to refresh them. |

## Environment variable override

Set `NEETOKB_BASE_URL` to point the CLI at a staging or local server:

```bash
export NEETOKB_BASE_URL=http://acme.lvh.me:8980
neetokb login --subdomain acme
```

## Error surface

Every command exits non-zero on failure and writes a single-line message to
stderr. Common errors the agent should expect:

- `Not authenticated. Run 'neetokb login' to authenticate.` — empty credential store.
- `Multiple subdomains authenticated (acme, beta); specify --subdomain.` — pick one.
- `Not authenticated for "foo". Authenticated subdomains: acme, beta.` — bad `--subdomain`.
- `required flag(s) "xxx" not set` (from cobra) — missing required flag.
- API errors come through with the server's message body; inspect the
  JSON envelope (or the `--quiet` payload) for `error` / `errors` / `notice`
  keys and any suggestions the API returns.

## Product-specific commands

| Group | Commands |
|---|---|
| `articles` | `list`, `show`, `create`, `update`, `trash`, `bulk-trash`, `archive`, `unarchive`, `discard-draft`, `clone`, `bulk-update` |
| `articles unlisted-links` | `get`, `regenerate` |
| `articles schedules` | `list`, `create`, `delete` |
| `articles seo` | `show`, `update` |
| `articles revisions` | `list`, `show`, `restore` |
| `trash` | `list`, `restore`, `delete` |
| `attachments` | `create` (alias `upload`) |
| `categories` | `list`, `show`, `create`, `update`, `delete`, `reorder-articles` |
| `tags` | `list`, `create`, `update`, `delete`, `merge` |
| `authors` | `list` |
| `recommendations` | `list`, `show`, `create`, `update`, `delete` |
| `redirections` | `list`, `show`, `create`, `update`, `delete` |
| `navigation-links` | `list`, `update` |
| `reports` | `popular-searches`, `failed-searches`, `invalid-urls`, `visits` |
| `reports clear` | `popular-searches`, `failed-searches`, `invalid-urls` |
| `search` | (top-level) full-text article search; `--search-term` required |
| `team-members` | `list`, `show`, `create`, `update`, `delete` |
| `workspace` | `info`, `update` |
| `workspace seo` | `show`, `update` |
| `preferences` | `show`, `update` |
| `integrations` | `neeto-chat` |
| `themes` | `list`, `show`, `update`, `delete`, `clone`, `publish` |
| `themes config` | `show`, `update` |
| `themes fonts` | `create`, `update`, `delete` |
| `webhooks` | `list`, `show`, `create`, `update`, `delete`, `events` |

Article commands accept the article's slug, its permalink identifier
(`a-XXXXXXXX`), or its UUID as `<id>`.

`articles create --category` takes the category path most specific first:
`--category "Installation,Getting Started"` files the article in
`Installation`, whose parent is `Getting Started`. Names are matched, not IDs
or slugs, and missing categories are created.

`search` prints only the article ID, the matched snippet and the article URL
as a table, with the snippet flattened to plain text so the highlight markup
and the article's own line breaks cannot break the column alignment. `--json`
and `--toon` still carry every field the API returns, including `title`,
`category`, `category_url`, `matched_title` and `matched_category`.

`--sort-by` and `--order-by` (`asc` or `desc`) are checked before the request is sent.
`articles list --state` takes `draft`, `published` or `all` and defaults to
published, so a full export needs `--state all`. List rows carry `created_at`,
`updated_at`, `author`, `tags`, `url` and `category.is_private`. `--sort-by`
takes `title`, `slug`, `state`, `category_name`, `views_count`, `created_at` or
`updated_at`; `--order-by` takes `asc` or `desc`.

`articles update --category "Installation,Getting Started"` moves an article to
another category path, most specific first. Missing categories are created.

Publishing (`create --state published`, and any update to a published article)
needs the Manage articles permission. Drafts need Modify articles.

`articles unlisted-links get` never creates a link; it fails when the article
has none or the link expired. Use `regenerate` to create the first one. It
replaces the old one, and the old URL stops working.

`authors list` rows carry `articles_count`, `published_articles_count` and
`draft_articles_count`. Use `--sort-by articles_count` (or `name`) with
`--order-by asc|desc`.

`categories list --flat` returns every category, including empty and private
ones, with `is_private` and published counts. The default tree omits empty
categories, and omits private ones for API keys.

`attachments create --file ./diagram.png` or `--url https://...` uploads a
file (`--filename` and `--content-type` override the defaults). Exactly one of
`--url` or `--file` is allowed, and `--file` is capped at 10 MB (use `--url` for
larger files). Paste `.data.attachable_html` into
`articles update --html-content`.

Unlisted links exist only for published articles; requesting one for a draft
errors with "Article must be published before an unlisted link can be
generated." `regenerate` invalidates the previous URL immediately and takes
`--expiration-type` (`never`, `one_day`, `seven_days`, `thirty_days`,
`custom`) with `--expiration-date` required only for `custom`.

Removing an article has three levels. `articles trash` moves it to the trash for
30 days and `trash restore` brings it back. `articles archive` keeps it but turns
it into a draft hidden from the public site (`unarchive` reverses it; list with
`articles list --state archived`). `articles discard-draft` only throws away
unpublished draft edits and keeps the live version. `trash delete` removes a
trashed article for good and cannot be undone. `bulk-trash` and `bulk-update`
take `--ids` (up to 100) and are all or nothing.

`articles schedules create <id> --kind publish|unpublish --datetime <ISO 8601>
--time-zone <IANA name>` needs a future time. An article has one schedule per kind,
and when it has both, the unpublish time must come before the publish time. `articles create` and `articles update` also take
`--tags`, `--related-articles` (both replace the whole list; an empty value
clears it) and `--note`. `articles revisions restore` copies a revision into the
draft only; the live content changes when the draft is published.

`categories delete <id>` moves the articles and subcategories to
`--move-to-category-id`, which a top-level category that has content needs. The last
category cannot be deleted. `categories update --parent ""` moves a category to the
top level. `tags merge --primary-id <id> --secondary-ids a,b` keeps the primary tag
and deletes the secondary tags after giving their articles the primary tag.

`reports` commands take `--duration` (`all_time`, `last_24_hours`, `last_7_days`,
`this_week`, `last_week`, `last_30_days`, `this_month`, `last_month`, `this_quarter`,
`last_quarter` or `custom`); `custom` needs both `--start-date` and `--end-date`.
`reports clear` deletes the data for that period for good, and everything of that
kind when no period is given. `navigation-links update --json-file links.json` takes
`{"header_links":[...],"footer_links":[...]}`; a set that is sent replaces the whole
set and links left out are deleted. `redirections create --type` is `page` (default)
or `domain`. `recommendations create --entity Article:<id>` can repeat; on update
`--entity` replaces all entities and `--clear-entities` removes them.

`themes publish <id>` makes that theme live on the public site. Built-in themes
cannot be edited: run `themes clone <id>`, then change the clone with
`themes config update <id> --color header-bg=#ffffff --custom-css ...`.
`--brand-image` and `--header-image` take a local file, which is uploaded first, or a
signed ID from `attachments create`. A published theme cannot be deleted.

A webhook secret is write-only: set it with `webhooks create|update --secret` and it
is never shown again. `--event` repeats (`create`, `update`, `delete`, `publish`,
`unpublish`); on update it replaces the list. `preferences` belongs to a signed-in
user, so it works with a CLI login and fails with an API key. `workspace update` and
`workspace seo update` take `--allow-signup=false` style booleans.

Run `neetokb commands` for the authoritative flag list, and see
<https://apidocs.neetokb.com/cli/introduction> for the full reference.
