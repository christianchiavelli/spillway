"""Outlines of a word set in Lora, as pathops paths in font units, y up. build.py draws with them."""
import io

import pathops
import uharfbuzz as hb
from fontTools.pens.recordingPen import DecomposingRecordingPen
from fontTools.pens.transformPen import TransformPen
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont


def load(path, weight):
    font = TTFont(path)
    instantiateVariableFont(font, {"wght": weight}, inplace=True)
    font.flavor = None
    buf = io.BytesIO()
    font.save(buf)
    data = buf.getvalue()
    return TTFont(io.BytesIO(data)), data


def word(path, text, weight, tracking=0):
    """One pathops path per glyph, placed along the baseline, with its name."""
    font, data = load(path, weight)
    face = hb.Face(data)
    hbfont = hb.Font(face)
    buf = hb.Buffer()
    buf.add_str(text)
    buf.guess_segment_properties()
    hb.shape(hbfont, buf, {"kern": True, "liga": False})
    glyphset = font.getGlyphSet()
    order = font.getGlyphOrder()
    x = 0
    out = []
    for info, pos in zip(buf.glyph_infos, buf.glyph_positions):
        name = order[info.codepoint]
        p = pathops.Path()
        pen = TransformPen(p.getPen(), (1, 0, 0, 1, x + pos.x_offset, pos.y_offset))
        rec = DecomposingRecordingPen(glyphset)
        glyphset[name].draw(rec)
        rec.replay(pen)
        out.append((name, p))
        x += pos.x_advance + tracking
    os2 = font["OS/2"]
    metrics = {
        "upm": font["head"].unitsPerEm,
        "xh": os2.sxHeight,
        "cap": os2.sCapHeight,
        "asc": font["hhea"].ascent,
        "desc": font["hhea"].descent,
        "advance": x - tracking,
    }
    return out, metrics

