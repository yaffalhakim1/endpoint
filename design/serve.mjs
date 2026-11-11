// Serves the design comp to other devices on the local network.
// Run: node design/serve.mjs
// Node is used because the Windows firewall already allows node.exe inbound;
// its rules block bun.exe.
import { createServer } from "node:http";
import { createReadStream, existsSync, statSync } from "node:fs";
import { extname, join, normalize, sep } from "node:path";
import { networkInterfaces } from "node:os";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL(".", import.meta.url));
const port = Number(process.env.PORT ?? 4173);

const types = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".mjs": "text/javascript; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
};

createServer((req, res) => {
  const pathname = decodeURIComponent(new URL(req.url, "http://local").pathname);
  const rel = pathname === "/" ? "endpoint.html" : pathname.slice(1);
  const file = normalize(join(root, rel));

  const inside = file === root || file.startsWith(root.endsWith(sep) ? root : root + sep);
  if (!inside || !existsSync(file) || !statSync(file).isFile()) {
    res.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
    res.end("Not found");
    return;
  }

  res.writeHead(200, {
    "content-type": types[extname(file).toLowerCase()] ?? "application/octet-stream",
    "cache-control": "no-store",
  });
  createReadStream(file).pipe(res);
}).listen(port, "0.0.0.0", () => {
  console.log(`local    http://localhost:${port}/`);
  for (const net of Object.values(networkInterfaces()).flat()) {
    if (net && net.family === "IPv4" && !net.internal) {
      console.log(`network  http://${net.address}:${port}/`);
    }
  }
  console.log(`serving  ${root}`);
});
