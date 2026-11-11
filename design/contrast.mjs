// Checks every text/background pair of the design against WCAG AA (4.5:1).
// Run: node design/contrast.mjs
const hex = (h) => h.replace("#", "").match(/../g).map((v) => parseInt(v, 16));

const toLinear = (c) => {
  const s = c / 255;
  return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
};

const luminance = ([r, g, b]) => 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b);

const ratio = (a, b) => {
  const l1 = luminance(a);
  const l2 = luminance(b);
  return (Math.max(l1, l2) + 0.05) / (Math.min(l1, l2) + 0.05);
};

const blend = (fg, alpha, bg) => fg.map((c, i) => Math.round(c * alpha + bg[i] * (1 - alpha)));

const themes = {
  light: {
    panel: "#ffffff",
    panel2: "#f7f7f9",
    ink: "#16161a",
    ink2: "#55555e",
    ink3: "#6e6e79",
    red: "#e81832",
    redInk: "#c5001f",
    green: "#17784a",
    amber: "#8a5a06",
    blue: "#1f5fbf",
  },
  dark: {
    panel: "#141417",
    panel2: "#191920",
    ink: "#ececf1",
    ink2: "#a5a5b0",
    ink3: "#8c8c97",
    red: "#e81832",
    redInk: "#ff6b78",
    green: "#5fbf8b",
    amber: "#d9a441",
    blue: "#8ab4f8",
  },
};

const pairs = [
  ["ink", "panel"],
  ["ink", "panel2"],
  ["ink2", "panel"],
  ["ink2", "panel2"],
  ["ink3", "panel"],
  ["ink3", "panel2"],
  ["green", "panel"],
  ["amber", "panel"],
  ["blue", "panel"],
  ["redInk", "panel"],
  ["redInk", "panel2"],
];

let failures = 0;

for (const [theme, t] of Object.entries(themes)) {
  console.log(`\n${theme}`);

  const rows = [...pairs, ["white", "red"]];

  for (const [fgName, bgName] of rows) {
    const fg = fgName === "white" ? hex("#ffffff") : hex(t[fgName]);
    const bg = hex(t[bgName]);
    const r = ratio(fg, bg);
    const ok = r >= 4.5;
    if (!ok) failures++;
    console.log(`  ${ok ? "pass" : "FAIL"}  ${r.toFixed(2)}  ${fgName} on ${bgName}`);
  }

  // Pill and chip fills: 12% of the ink over the panel.
  for (const name of ["redInk", "green", "amber", "blue"]) {
    const bg = blend(hex(t[name]), 0.12, hex(t.panel));
    const r = ratio(hex(t[name]), bg);
    const ok = r >= 4.5;
    if (!ok) failures++;
    console.log(`  ${ok ? "pass" : "FAIL"}  ${r.toFixed(2)}  ${name} on ${name} 12%`);
  }
}

console.log(`\n${failures === 0 ? "all pairs pass AA" : failures + " failing pairs"}`);
process.exit(failures === 0 ? 0 : 1);
