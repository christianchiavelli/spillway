"""Draws the map of the grid's four subsystems, and writes it where the app reads it.

The states come from Natural Earth (public domain), are projected as IBGE
projects Brazil, joined into the subsystems their power plants belong to, and
simplified together, so two neighbours keep one border between them. The app
gets one SVG path and one label point per subsystem, in its own units.

From the repository root:

    uv run --no-project --with shapely==2.2.0 --with pyproj==3.8.0 \\
      python -I -X utf8 docs/map/build.py
"""

import json
import urllib.request
from pathlib import Path

import shapely
from pyproj import Transformer
from shapely.ops import polylabel

ROOT = Path(__file__).parent.parent.parent
OUT = ROOT / "apps/web/src/lib/features/curtailment/subsystem-shapes.ts"

STATES = (
    "https://raw.githubusercontent.com/nvkelso/natural-earth-vector/"
    "v5.1.2/geojson/ne_50m_admin_1_states_provinces.geojson"
)

# The ONS draws its subsystems around the grid, not the states: a state goes
# where most of its installed capacity sits in the ONS's capacidade-geracao
# dataset. So Tocantins, whose dams feed the Southeast, is drawn there, as are
# Rondônia and Mato Grosso, and Paraná is drawn in the South though Itaipu is
# counted in the Southeast. Acre and the Federal District have no plant of
# their own: Acre is on Rondônia's lines, and the District sits inside Goiás.
SUBSYSTEMS = {
    "N": ["AM", "AP", "MA", "PA", "RR"],
    "NE": ["AL", "BA", "CE", "PB", "PE", "PI", "RN", "SE"],
    "SE": ["AC", "DF", "ES", "GO", "MG", "MS", "MT", "RJ", "RO", "SP", "TO"],
    "S": ["PR", "RS", "SC"],
}

# Brazil's Albers equal-area conic, as IBGE uses it, so no region looks larger
# than it is next to another.
ALBERS = "+proj=aea +lat_0=-12 +lon_0=-54 +lat_1=-2 +lat_2=-22 +ellps=GRS80 +units=m"

WIDTH = 480  # the map's width in its own units; the height follows
MIN_ISLAND = 300e6  # m²: islands smaller than this are a speck at any size shown
TOLERANCE = 20000  # m: about two units, where a coast keeps its shape but not its every inlet


def num(v):
    return f"{v:.1f}".rstrip("0").rstrip(".")


def without_specks(geometry):
    parts = shapely.get_parts(geometry)
    return shapely.multipolygons([p for p in parts if p.area >= MIN_ISLAND])


with urllib.request.urlopen(STATES) as response:
    features = json.load(response)["features"]
by_state = {
    f["properties"]["postal"]: shapely.geometry.shape(f["geometry"])
    for f in features
    if f["properties"]["iso_a2"] == "BR"
}
drawn = sorted(s for states in SUBSYSTEMS.values() for s in states)
assert drawn == sorted(by_state), "every state in exactly one subsystem"

to_albers = Transformer.from_crs("EPSG:4326", ALBERS, always_xy=True)
project = lambda g: shapely.transform(g, lambda xy: list(zip(*to_albers.transform(*xy.T))))  # noqa: E731

regions = {
    code: without_specks(shapely.union_all([project(by_state[s]) for s in states]))
    for code, states in SUBSYSTEMS.items()
}
simplified = dict(zip(regions, shapely.coverage_simplify(list(regions.values()), TOLERANCE)))

x0, y0, x1, y1 = shapely.total_bounds(list(simplified.values()))
k = WIDTH / (x1 - x0)
HEIGHT = (y1 - y0) * k


def point(x, y):
    return f"{num((x - x0) * k)} {num((y1 - y) * k)}"


def path(geometry):
    d = []
    for polygon in shapely.get_parts(geometry):
        for ring in [polygon.exterior, *polygon.interiors]:
            d.append("M" + "L".join(point(x, y) for x, y in ring.coords[:-1]) + "Z")
    return "".join(d)


def label_at(geometry):
    # The point farthest inside the largest part, where a label has most room.
    largest = max(shapely.get_parts(geometry), key=lambda p: p.area)
    p = polylabel(largest, tolerance=1000)
    return f"[{num((p.x - x0) * k)}, {num((y1 - p.y) * k)}]"


entries = "\n".join(
    f"  {code}: {{\n    d: '{path(g)}',\n    labelAt: {label_at(g)},\n  }},"
    for code, g in simplified.items()
)
OUT.write_text(
    f"""// Written by docs/map/build.py from Natural Earth's states: not to be edited.
import type {{ Subsystem }} from '#lib/api/generated/types.gen.ts';

export const width = {WIDTH};
export const height = {num(HEIGHT)};

/** Each subsystem's outline, and where its label goes, in the map's units. */
export const shapes: Record<Subsystem, {{ d: string; labelAt: [number, number] }}> = {{
{entries}
}};
""",
    encoding="utf-8",
    newline="\n",
)
points = sum(
    len(ring.coords) - 1
    for g in simplified.values()
    for polygon in shapely.get_parts(g)
    for ring in [polygon.exterior, *polygon.interiors]
)
print(OUT.relative_to(ROOT).as_posix())
print(f"{WIDTH} x {num(HEIGHT)}, {points} points, {OUT.stat().st_size} bytes")
