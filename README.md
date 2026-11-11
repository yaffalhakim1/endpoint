# Endpoint

A desktop HTTP client, written in Go with [MyGo](https://github.com/egoist/mygo). Native UI drawn on the GPU: no webview, no HTML, no JavaScript, and it opens instantly.

## Download

Grab the build for your platform from the [releases page](../../releases/latest):

- **Windows** — `endpoint-*-setup-*.exe`, or the plain `Endpoint.exe`
- **macOS** — `endpoint-*.dmg` (unsigned; right-click the app and choose Open the first time)
- **Linux** — the `.AppImage`-free archive with `install.sh`, or the `.deb`

## What it does

- Projects, folders and requests, kept in a workspace file in your user data directory
- Environments with `{{variables}}`, resolved in the URL, headers, auth and body
- Params, headers, five auth types, JSON, multipart and file bodies
- Requests sent from Go, so no CORS, any header, and cookies left alone
- Responses with status, timing, size, the body in a folding code editor, and every header
- Light and dark, following the system, on Apple's colours with Apple Music's red
- A menu bar, keyboard shortcuts, and a title bar drawn by the app

## Develop

Needs Go 1.27 or later.

```sh
go tool mygo dev     # run with live reload
go test ./...        # the suite, offline
go tool mygo build   # package for this machine
```

The suite is offline by default. To run the checks that call a real API:

```sh
ENDPOINT_LIVE=1 go test -run TestLive -v
```

## Release

Bump `version` in `mygo.json`, commit, then push a tag:

```sh
git tag v0.2.0
git push origin v0.2.0
```

The workflow in `.github/workflows/release.yml` builds every platform and attaches the results to that tag's release.

## Type

Inter and Geist Mono, both under the SIL Open Font License; their licences sit beside them in `resources/fonts`.
