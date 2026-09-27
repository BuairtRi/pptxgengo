#!/usr/bin/env python3
"""Extract one full-frame EMF+ 32-bit bitmap without rendering or changing pixels.

Rejects vector/multiple-image/transformed EMFs. Used only for alpha-bound
measurement; the original EMF remains embedded in the PowerPoint output.
"""
import argparse
import hashlib
import json
from pathlib import Path
import struct
import zlib


def extract(source, output):
    b = source.read_bytes()
    images, draws = [], []
    i = 0
    allowed = {0x4001, 0x4002, 0x4030, 0x4009, 0x4023, 0x4008, 0x401a}
    while i + 8 <= len(b):
        kind, size = struct.unpack_from('<II', b, i)
        if size < 8 or i + size > len(b):
            raise ValueError('Invalid EMF record')
        if kind == 70 and b[i + 12:i + 16] == b'EMF+':
            j = i + 16
            while j + 12 <= i + size:
                t, flags, n, data_size = struct.unpack_from('<HHII', b, j)
                if n < 12 or j + n > i + size or data_size > n - 12 or t not in allowed:
                    raise ValueError('Unsupported EMF+ record')
                data = b[j + 12:j + 12 + data_size]
                if t == 0x4009 and (len(data) != 4 or struct.unpack('<I', data)[0] >> 24):
                    raise ValueError('Nontransparent EMF canvas is unsupported')
                if t == 0x4030 and (flags != 2 or data != struct.pack('<f', 1.0)):
                    raise ValueError('Nonidentity page transform is unsupported')
                if t == 0x4008 and (flags >> 8) not in (5, 8):
                    raise ValueError('Additional EMF objects are unsupported')
                if t == 0x4008 and (flags >> 8) == 8 and (flags & 255 or len(data) != 24 or data[4:] != struct.pack('<5I', 0, 0, 0xffffffff, 0, 0)):
                    raise ValueError('Nondefault image attributes are unsupported')
                if t == 0x4008 and (flags >> 8) == 5:
                    version, image_type, w, h, stride, pixel_format, bitmap_type = struct.unpack_from('<7I', data)
                    if image_type != 1 or bitmap_type != 0 or stride != w * 4 or pixel_format not in (0x26200a, 0xe200b) or len(data) != 28 + w * h * 4:
                        raise ValueError('Requires one packed ARGB/PARGB bitmap')
                    images.append((flags & 255, w, h, pixel_format, data[28:]))
                if t == 0x401a:
                    if len(data) != 40 or flags & 0xff00:
                        raise ValueError('Unsupported image transform')
                    draws.append((flags & 255, struct.unpack('<II8f', data)))
                j += n
        i += size
    if len(images) != 1 or len(draws) != 1:
        raise ValueError('Requires exactly one embedded and drawn bitmap')
    identity, w, h, pixel_format, pixels = images[0]
    draw_id, d = draws[0]
    if draw_id != identity or d != (0, 2, 0., 0., float(w), float(h), 0., 0., float(w), float(h)):
        raise ValueError('Image is cropped, scaled, or offset within EMF canvas')
    # PNG stores unassociated RGBA; preserve alpha exactly.
    rows = bytearray()
    for y in range(h):
        rows.append(0)
        for x in range(w):
            blue, green, red, alpha = pixels[(y * w + x) * 4:(y * w + x + 1) * 4]
            if pixel_format == 0xe200b and alpha:
                red, green, blue = [min(255, round(v * 255 / alpha)) for v in (red, green, blue)]
            rows.extend((red, green, blue, alpha))
    def chunk(tag, value):
        return struct.pack('>I', len(value)) + tag + value + struct.pack('>I', zlib.crc32(tag + value) & 0xffffffff)
    png = b'\x89PNG\r\n\x1a\n' + chunk(b'IHDR', struct.pack('>2I5B', w, h, 8, 6, 0, 0, 0)) + chunk(b'IDAT', zlib.compress(rows)) + chunk(b'IEND', b'')
    with output.open('xb') as f:
        f.write(png)
    return {'schema': 'pptxgengo.emf-bitmap-extraction.v1', 'source': str(source), 'source_sha256': hashlib.sha256(b).hexdigest(), 'output': str(output), 'output_sha256': hashlib.sha256(png).hexdigest(), 'width': w, 'height': h, 'scope': 'Single full-frame EMF+ bitmap; alpha retained exactly; original EMF remains embedded.'}


if __name__ == '__main__':
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('source', type=Path)
    ap.add_argument('output', type=Path)
    a = ap.parse_args()
    print(json.dumps(extract(a.source.resolve(), a.output.resolve()), indent=2))
