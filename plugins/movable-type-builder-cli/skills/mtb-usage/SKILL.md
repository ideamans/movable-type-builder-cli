---
name: mtb-usage
description: Build and edit a Movable Type site with the mtb CLI (Movable Type Builder) — list, create and update entries and pages, upload assets in bulk, manage categories, custom fields, content types and templates, keep templates in Git, trigger rebuilds, and switch between several MT servers with named profiles. Use when someone asks about a Movable Type site, an MT entry or article, publishing or rebuilding an MT blog, or calling the Movable Type Data API.
license: MIT
compatibility: Requires the `mtb` binary on PATH — run the mtb-install skill if it is missing. Needs a configured profile and a login (`mtb auth login`), which asks for the MT user's *Web Services Password*. Writes to a live CMS — entries created or deleted here are real.
allowed-tools: Bash(mtb:*) Bash(jq:*) Bash(command:*) Read Write
---

# mtb-usage

Drive Movable Type through the `mtb` CLI. Commands mirror the Data API v7
paths (`mtb sites entries list` is `GET /v7/sites/{site_id}/entries`) and print
JSON.

## 1. Confirm the tool, the target and the login

```bash
command -v mtb && mtb --version
mtb profile show     # which server am I about to touch?
mtb auth status
```

Missing binary? Run the `mtb-install` skill.

`mtb profile show` is the important one. **Say out loud which server you are
about to write to** before any create, update or delete — the same commands run
against staging and production, and only the profile tells them apart.

Before writing, tag the work so it can be traced afterwards. mtb sends the text as
the `X-MTB-Audit-Note` header on every request, and servers that keep an audit log
(the MtbPro plugin does) store it with each call:

```bash
export MTB_AUDIT_NOTE="<what this task is, who asked, date>"
```

## 2. Read the reference before composing commands

```bash
mtb llm | head -120        # rules, auth, workflows, failure modes
mtb llm                    # everything, including the command catalogue
```

It is embedded in the binary, so it always matches the version installed.

**Installed extensions add commands and a chapter.** An extension gives the
endpoints an MT plugin adds (`mtb sites entries author update`, for example)
the same kind of command as the built-in ones. They are not in the command
catalogue; each has its own "拡張: <name>" chapter in `mtb llm`.

```bash
mtb extensions list                  # which extensions are loaded
mtb extensions check <name>          # does this MT have the plugin behind it?
```

`mtb plugins` operates MT's own plugins; `mtb extensions` manages mtb's — they
are different things.

**Messages, help and errors switch between Japanese and English.** The language
comes from `--lang ja` / `--lang en`, then `MTB_LANG`, then `LC_ALL` /
`LC_MESSAGES` / `LANG`; with none of those set it is English. If you match on
the text of an error or a prompt, pass `--lang` so the wording cannot move under
you. Endpoint descriptions come from the official OpenAPI definition and stay
English in both languages, and `mtb llm` is always Japanese.

## 3. Set up a connection when there is none

```bash
mtb profile add prod --base-url https://example.com/cgi-bin/mt/mt-data-api.cgi
mtb profile set --site-id 1     # then --site-id can be omitted
```

The base URL ends at `mt-data-api.cgi`. A site's top-level URL gives 404s.

**mtb reads `.env` from the current directory**, so the working directory can
decide which server you talk to (`MTB_PROFILE`, `MTB_BASE_URL`, `MTB_SITE_ID`).
Real environment variables win over the file. Run `mtb profile show` before the
first write — it prints the resolved target and the path of any `.env` it read.
To pin the target regardless, pass `--profile` or `--base-url`; both beat `.env`.

**Do not run `mtb auth login` unattended.** It asks for a password. Ask the
user to run it, or — if they want it automated — have them pipe the password in:

```bash
pass show mt/prod | mtb auth login --username alice --password-stdin
```

`--password-stdin` leaves the password nowhere: not in the process list, not in
the environment, not in your transcript. `MTB_PASSWORD` also works and is
visible to other processes on the same host. **Never use `--password`** — it puts
the secret in the process list and in whatever log holds this transcript.

The password wanted here is the MT **Web Services Password**, not the sign-in
password. That mismatch is the single most common cause of a 401.

## 4. Read before you write

```bash
mtb sites list
mtb sites entries list --site-id 1 --limit 50 --status Publish
mtb sites entries get 1 42
```

Lists return 10 items unless `--limit` says otherwise. Check `totalResults` in
the response before concluding something does not exist.

## 5. Write

```bash
mtb sites entries create --site-id 1 --entry '{"title":"…","body":"…","status":"Publish"}'
mtb sites entries update 1 42 --entry '{"title":"…"}'
mtb assets upload --site-id 1 --file ./photo.jpg --path /2026
```

Updates are partial: send only the fields that change — **except a page's `folder`.**
A page update that leaves out `folder` takes the page out of its folder, even when
it only changes the status or the title: the URL moves to the site root, and a page
whose basename is `index` overwrites the site's top page (`/index.html`) the moment
it is published. Send the folder with every page update:

```bash
mtb sites pages update 1 7 --page '{"status":"Publish","folder":{"id":4}}' --publish 1 --verify-publish
```

**Never write back a body you read with a plain `get`.** What comes back is the
*rendered* HTML, not the source. If the entry's `format` is `markdown`,
`block_editor` or similar, writing it back replaces the source with HTML — the
format stays the same, the response looks fine, and nobody notices until someone
opens the entry to edit it. Read it with `--no-text-filter 1`:

```bash
mtb sites entries get 4 42 --no-text-filter 1 --extract body --out body.md
$EDITOR body.md
mtb sites entries update 4 42 --set entry.body=@body.md
```

Updates that do not touch the body (title, status, categories) are safe as they are.
`mtb text-filters list` gives the valid `format` values.

**MT can be extended with plugins that add Data API endpoints.** Those are not
in the embedded definition, so no command is generated for them — reach them
with `mtb api`, which accepts the same `--all`, `--fields`, `--set` and
`--set-json` as the generated commands. Ask the server what it actually has:

```bash
mtb endpoints list --all --fields component,verb,route
mtb api GET /sites/1/things --all --fields id,label
```

On a stock MT 9, 246 endpoints come from Core and 20 from plugins — custom
fields and comments among them.

**You do not need to build JSON by hand, and you never need `jq`.** Use `--set`
for the input side and `--extract` for the output side:

```bash
mtb sites templates create --site-id 1 \
  --set template.name="Main Index" --set template.type=index \
  --set template.outputFile=index.html --set template.text=@index.html

mtb sites templates get 1 42 --extract text --out index.html
```

`--set` values are strings; use `--set-json` for numbers, booleans and nested
objects. The field names a resource accepts are listed in its `--help`, e.g.
`--template string (object: name, outputFile, text, type, ...)`. `--extract`
behaves like `jq -r`: strings come out raw, anything else as JSON, and dotted
paths such as `items.0.name` walk arrays. With `--out` the bytes are exact.

To find out what differs between two MT servers — staging versus production,
say — dump both and diff the directories:

```bash
mtb dump --profile stg  --dir ./dump/stg
mtb dump --profile prod --dir ./dump/prod
diff -ru ./dump/prod ./dump/stg
```

IDs, timestamps and authors are left out, objects are ordered and matched by
natural keys, and multi-line values are literal blocks, so the diff shows real
differences line by line. Add `--asset-hash` to compare file contents (one extra
request per asset). This is read-only — a safe first step before any sync work.

To keep templates in Git there are two routes. If you can put files on the
server, develop the design as a **theme** — see the `mtb-theme` skill; the theme
becomes the source of truth and a single `apply` rebuilds the site. If you only
have the API, pull the templates into a directory and push them back:

```bash
mtb sites templates pull 4 --dir ./templates
mtb sites templates push 4 --dir ./templates
```

Each template becomes one `.mtml` file; `mtb-templates.json` holds the type and
output file name and maps files to IDs. `push` sends only what changed and
**never deletes** templates that exist on MT but not locally — it reports them.
`pull` overwrites local files, so check `--dry-run` if the user has uncommitted
edits. Archive mappings are out of scope: use `sites templates templatemaps`.

For a whole directory of design assets, `--dir` sends them one by one and
mirrors subdirectories at the destination:

```bash
mtb assets upload --site-id 1 --dir ./dist/theme-images --path /theme-images
```

**An asset cannot be overwritten.** Re-uploading the same filename returns 409;
the Data API has no overwrite. Use `--replace`, which deletes the asset at that
destination first. It is destructive, so it needs `--yes` — tell the user what
will be deleted before you pass it. Only assets whose URL matches the
destination are removed; a file of the same name in another folder is untouched.

**`.js` and `.html` cannot be uploaded as assets.** MT's
`DeniedAssetFileExtensions` blocks them by default. In MT, stylesheets and
JavaScript belong in **index templates** — create them with
`--set template.type=index --set template.outputFile=theme.js`. Images, CSS,
SVG, fonts and webp upload fine.

**`mtb sites themes apply` replaces every template on the site.** It now asks
for `--confirm "<site name>"` on top of `--yes`, like the other operations that
destroy a whole container. Never pass both on the user's behalf without saying
which site you are about to overwrite.

Omitting `status` creates a **draft**. If the user asked for something published,
either set `"status":"Publish"` or tell them it was left as a draft.

Preview the request first when the shape is uncertain:

```bash
mtb sites entries create --site-id 1 --entry @entry.json --dry-run
```

## 6. Destructive operations

`DELETE` asks for confirmation, and refuses to run unattended without `--yes`.
That guard exists for you.

- Confirm the target with `mtb sites entries get` first, and show the user what
  it is.
- Pass `--yes` **only** for a deletion the user asked for in this session.
- `mtb sites logs reset` wipes the whole activity log, and
  `mtb sites refresh-templates` overwrites customised templates with the
  defaults. Neither is undoable. Get explicit agreement before either.

## 7. Publish the change

Content changes do not appear on a statically published site until it rebuilds:

```bash
mtb publish entries --ids 42,43
mtb sites templates publish 1 5
```

Say whether you rebuilt or not — "I updated the entry" and "the change is live"
are different claims.

## Recovering from errors

| What you see | What to do |
| --- | --- |
| 401 on login | Ask the user to check the *web services* password |
| 401 mid-session | `mtb auth login` again; the session expired |
| 403 with a message | Permission, not a bug. Report which operation and let the user fix the role (`Only a system administrator`: no role is enough) |
| 403 `Forbidden` with no message | The site's Data API is off (every new site starts that way); not a role problem |
| 404 `Unknown endpoint` | `mtb profile show` — the base URL is probably wrong, or that MT lacks the endpoint |
| 404 `Unknown endpoint` from an extension's command | The MT plugin behind it is missing: `mtb extensions check <name>` |
| 404 with another message (`Site not found` …) | The endpoint exists; the ID or site is wrong |
| 405 | Retry with `--method-override` (a proxy is blocking PUT/DELETE) |
| HTML instead of JSON | Same as 404: the base URL points at the admin screen |

For an endpoint the CLI does not know (a plugin's with no extension, or a newer MT),
call it by its endpoint ID and let the server tell you the parameters:

```bash
mtb endpoints list --fields id,verb,route,component   # find the ID
mtb call <endpoint-id> --describe                     # parameters, from the server's own definition
mtb call <endpoint-id> 1 42 --form name=value         # path values in route order
```

Or with an explicit path:

```bash
mtb endpoints list                       # what this server actually offers
mtb api GET /sites/1/some/plugin/path
```
