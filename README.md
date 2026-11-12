# Endpoint

A desktop HTTP client. Go and native UI, no webview.

**[Download for Windows, macOS or Linux](https://github.com/yaffalhakim1/endpoint/releases/latest)**

## Why another one

Electron clients ship a browser to draw a form. Endpoint draws its own window on the GPU, so it starts in about a third of a second and sits near 40 MB. There is no Chromium inside.

Windows, macOS and Linux. Same binary size, same speed.

## What it does

Send requests to any API. GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, with params, headers and a body.

Organize work as projects, folders and requests. A folder holds requests, a project holds folders and its own credentials and variables.

Switch environments. Write `{{baseUrl}}` in a URL or a header, and Endpoint fills it in at send time. Point staging and production at different hosts without editing every request.

Authorize with Bearer, Basic, an API key header, or the project's own credentials. Secrets stay out of the editor and out of the file: the workspace is encrypted with AES-GCM under a key in your system's credential store.

Read the answer. Status, timing and size at a glance, the body in a folding code editor, and every header the server sent on its own tab.

## 简体中文

Endpoint 是一个桌面 HTTP 客户端，用 Go 编写，自绘界面。它不打包浏览器，启动约三分之一秒，内存约 40 MB。

支持项目、文件夹与请求，支持环境变量（`{{baseUrl}}`）、五种认证方式、参数、请求头、JSON / 表单 / 文件请求体。响应可以折叠查看，也能查看所有响应头。

工作区使用 AES-GCM 加密，密钥保存在系统凭据存储中。

[下载最新版本](https://github.com/yaffalhakim1/endpoint/releases/latest)

## Getting started

Download the build for your platform from the releases page.

Windows runs the installer or the plain executable. On macOS, right-click the app and choose Open the first time: the build is unsigned. Linux takes the `.deb`, the `.rpm`, or the archive with `install.sh`.

The app opens with three projects so nothing is empty. Change the `baseUrl` variable in the project menu, then press Send.

## Build it yourself

Needs Go 1.27 or later.

```sh
go tool mygo dev     # live reload
go test ./...        # the suite, offline
go tool mygo build   # package for this machine
```

The suite stays offline. Point it at a real API when you want:

```sh
ENDPOINT_LIVE=1 go test -run TestLive -v
```

## Where things live

```
cmd/endpoint/      the entry point: window options, startup
internal/app/      the window: state, views, theme, menu
internal/request/  sending, and the formatting a response needs
internal/workspace/ the document: projects, requests, sealing, recovery
site/              the landing page, published to Pages
resources/         the app icon
```

`workspace` knows nothing about the window, and `request` knows nothing about the document beyond its types. Tests sit beside the code they cover, and the ones that drive a window live in `internal/app`.

## Status

0.2.0. It sends requests, keeps your work, and survives a damaged workspace. Signing and auto-updates come next, so the first launch on each platform shows a warning.

## Licence

MIT. Inter and Geist Mono ship with it under the SIL Open Font License.
