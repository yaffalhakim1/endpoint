# Changelog

## 0.1.0

The first build anyone else can run.

- Projects, folders and requests, in a workspace file in the user data directory
- Environments with `{{variables}}`, resolved in the URL, headers, auth and body
- Params, headers, five auth types, and JSON, multipart and file bodies
- Requests sent from Go, so no CORS, any header, any method
- Responses with status, timing, size, the body in a folding editor, and every header
- Light and dark from Apple's palette, with Apple Music red, in Inter and Geist Mono
- A title bar drawn by the app: sidebar toggle, environment, menu, window controls
- The workspace is sealed with AES-GCM under a key in the system credential store,
  keeps a backup of the previous save, and starts fresh with a message rather than
  failing when a file is damaged
- A menu bar with every command and its shortcut
