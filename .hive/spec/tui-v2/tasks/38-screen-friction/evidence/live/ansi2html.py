# ansi2html.py IN.ansi OUT.html W H: render the top-left W×H cells of a pane
# capture as an HTML <pre> with the SGR colours and bold.
import html, re, sys
src, out, w, h = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4])
text = open(src).read().replace('\r', '')
rows = []
for line in text.split('\n')[:h]:
    fg, bold, col, buf = None, False, 0, []
    for part in re.split(r'(\x1b\[[0-9;]*m)', line):
        m = re.fullmatch(r'\x1b\[([0-9;]*)m', part)
        if m:
            codes = [int(c) if c else 0 for c in m.group(1).split(';')]
            i = 0
            while i < len(codes):
                c = codes[i]
                if c == 0: fg, bold = None, False
                elif c == 1: bold = True
                elif c == 22: bold = False
                elif c == 39: fg = None
                elif c == 38 and codes[i+1] == 2: fg = '#%02x%02x%02x' % tuple(codes[i+2:i+5]); i += 4
                i += 1
            continue
        for ch in part:
            if col >= w: break
            style = ('color:%s;' % fg if fg else '') + ('font-weight:700;' if bold else '')
            buf.append('<span style="%s">%s</span>' % (style, html.escape(ch)) if style else html.escape(ch))
            col += 1
    buf.append(' ' * (w - col))
    rows.append(''.join(buf))
open(out, 'w').write('<!doctype html><meta charset="utf-8"><body style="margin:8px;background:#1b1a17;color:#ece6da"><pre style="font:13px/1.2 \'JetBrains Mono\',monospace;margin:0">' + '\n'.join(rows) + '</pre>')
