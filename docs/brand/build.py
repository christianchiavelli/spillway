"""Draws Spillway's wordmark, mark and favicon, and writes them where they are used.

The wordmark is Lora, the typeface of the app's page titles, turned into
outlines so no font is needed to show it. Each l is a bar of a chart: its top,
cut off by a hairline, is the share of its source's potential the ONS curtailed
in 2025, the first full year it published both: wind's l is the first, solar's
the second. The mark draws the same two l's again on a 32-unit grid, as stems
with Lora's slanted flag, because Lora's own l's blur at 16 px and read as the
number 11 on a tile.

From the repository root, after `pnpm install`:

    uv run --no-project --with fonttools==4.67.0 --with brotli==1.2.0 \\
      --with uharfbuzz==0.56.3 --with skia-pathops==0.9.2 \\
      python -I -X utf8 docs/brand/build.py
"""

import sys
from pathlib import Path

import pathops
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen

HERE = Path(__file__).parent
ROOT = HERE.parent.parent
sys.path.insert(0, str(HERE))
from glyphs import word  # noqa: E402

LORA = ROOT / "apps/web/node_modules/@fontsource-variable/lora/files/lora-latin-wght-normal.woff2"

# The 2025 shares, from the ONS's restricao_coff datasets: curtailed energy over
# generated plus curtailed, for the whole country.
SHARE = {"wind": 0.191, "solar": 0.257}

# The app's colours (app.css): ink is the text blue on light grounds and white on
# dark ones; the sources keep the charts' hues, a step lighter on dark.
LIGHT = {"ink": "#002244", "wind": "#0a64c8", "solar": "#d97706", "ground": "#ffffff"}
DARK = {"ink": "#ffffff", "wind": "#28b4ff", "solar": "#f59e0b", "ground": "#11192d"}


def num(v):
    s = f"{v:.1f}"
    return s[:-2] if s.endswith(".0") else s


def rect(x0, y0, x1, y1):
    p = pathops.Path()
    p.moveTo(x0, y0)
    p.lineTo(x1, y0)
    p.lineTo(x1, y1)
    p.lineTo(x0, y1)
    p.close()
    return p


def op(a, b, kind):
    return pathops.op(a, b, kind, fix_winding=True)


def union(paths):
    out = pathops.Path()
    for p in paths:
        out = op(out, p, pathops.PathOp.UNION)
    return out


def svg(view, body):
    return (
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="{view}" role="img" aria-label="Spillway">'
        f"<title>Spillway</title>{body}</svg>\n"
    )


# The word, set tight in Lora SemiBold.
glyphs, metrics = word(str(LORA), "Spillway", 600, tracking=-8)
TOP = 760  # just above the l's ascender, where the outlines are flipped
L_TOP = 755  # the l's height
GAP = 14  # the cut: a hairline at large sizes, gone at small ones


def cut():
    """The word's ink, and the tops of its two l's: wind's, then solar's."""
    parts, caps = [], []
    for name, p in glyphs:
        if name != "l":
            parts.append(p)
            continue
        share = SHARE["wind"] if not caps else SHARE["solar"]
        limit = L_TOP * (1 - share)
        lo, hi = p.bounds[1], p.bounds[3]
        parts.append(op(p, rect(-1e4, lo - 10, 1e4, limit), pathops.PathOp.INTERSECTION))
        caps.append(op(p, rect(-1e4, limit + GAP, 1e4, hi + 10), pathops.PathOp.INTERSECTION))
    return union(parts), caps[0], caps[1]


ink, wind_cap, solar_cap = cut()
x0, y0, x1, _ = union([p for _, p in glyphs]).bounds
WIDTH = x1 - x0
HEIGHT = TOP - y0  # down to the descenders of p and y


def wordmark(c):
    def path(fill, p):
        pen = SVGPathPen(None, ntos=num)
        p.draw(TransformPen(pen, (1, 0, 0, -1, -x0, TOP)))
        return f'<path fill="{fill}" d="{pen.getCommands()}"/>'

    body = path(c["ink"], ink) + path(c["wind"], wind_cap) + path(c["solar"], solar_cap)
    return svg(f"0 0 {num(WIDTH)} {num(HEIGHT)}", body)


def mark(c, tile):
    base, stem = 26.5, 5.2
    flag = (8.0, 5.8)  # the top's left and right ends: it rises like Lora's
    height = base - sum(flag) / 2
    gap = 1.4
    parts = {"ink": [], "wind": [], "solar": []}
    for x, source in ((9.3, "wind"), (17.5, "solar")):
        limit = base - height * (1 - SHARE[source])
        parts["ink"].append(f"M{num(x)} {num(base)}V{num(limit)}h{num(stem)}V{num(base)}z")
        top = limit - gap
        parts[source].append(
            f"M{num(x)} {num(top)}V{num(flag[0])}l{num(stem)} {num(flag[1] - flag[0])}V{num(top)}z"
        )
    ground = f'<rect width="32" height="32" rx="7.5" fill="{c["ground"]}"/>' if tile else ""
    body = ground + "".join(f'<path fill="{c[k]}" d="{"".join(v)}"/>' for k, v in parts.items())
    return svg("0 0 32 32", body)


files = {
    HERE / "wordmark-on-light.svg": wordmark(LIGHT),
    HERE / "wordmark-on-dark.svg": wordmark(DARK),
    HERE / "mark-on-light.svg": mark(LIGHT, tile=False),
    HERE / "mark-on-dark.svg": mark(DARK, tile=False),
    HERE / "icon.svg": mark(DARK, tile=True),
    # Where the app uses them: the header is always dark, and the favicon is the
    # icon, whose tile holds its own ground in a light or a dark tab bar.
    ROOT / "apps/web/src/lib/components/layout/wordmark-on-dark.svg": wordmark(DARK),
    ROOT / "apps/web/static/favicon.svg": mark(DARK, tile=True),
}
for path, content in files.items():
    path.write_text(content, encoding="utf-8", newline="\n")
    print(path.relative_to(ROOT))
print(f"wordmark {num(WIDTH)} x {num(HEIGHT)}")
