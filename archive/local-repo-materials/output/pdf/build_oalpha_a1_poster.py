from __future__ import annotations

import math
from pathlib import Path

from reportlab.lib import colors
from reportlab.lib.pagesizes import A1
from reportlab.lib.units import mm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfgen import canvas


ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "output" / "pdf" / "oalpha_a1_orbital_poster.pdf"

W, H = A1

INK = colors.HexColor("#d8f4ff")
MUTED = colors.HexColor("#82a5b4")
CYAN = colors.HexColor("#18d9ea")
CYAN_DARK = colors.HexColor("#0b7f8f")
YELLOW = colors.HexColor("#ffd11d")
GREEN = colors.HexColor("#42e6a4")
PINK = colors.HexColor("#ff5c9a")
CARD = colors.HexColor("#101b2c")
CARD_2 = colors.HexColor("#0d1728")
BG = colors.HexColor("#06111d")
LINE = colors.HexColor("#1a3a50")


def register_fonts() -> tuple[str, str, str]:
    inter = Path.home() / "Library" / "Fonts" / "Inter-VariableFont_opsz,wght.ttf"
    inter_italic = Path.home() / "Library" / "Fonts" / "Inter-Italic-VariableFont_opsz,wght.ttf"
    try:
        if inter.exists():
            pdfmetrics.registerFont(TTFont("Inter", str(inter)))
            if inter_italic.exists():
                pdfmetrics.registerFont(TTFont("InterItalic", str(inter_italic)))
            return "Inter", "Inter", "Courier"
    except Exception:
        pass
    return "Helvetica", "Helvetica-Bold", "Courier"


FONT, BOLD, MONO = register_fonts()


def c(hex_color: str, alpha: float = 1.0) -> colors.Color:
    base = colors.HexColor(hex_color)
    return colors.Color(base.red, base.green, base.blue, alpha=alpha)


def text_width(text: str, size: float, font: str = FONT) -> float:
    return pdfmetrics.stringWidth(text, font, size)


def draw_bg(can: canvas.Canvas) -> None:
    can.setFillColor(BG)
    can.rect(0, 0, W, H, stroke=0, fill=1)

    # Soft vertical glow and vignette built from translucent bands.
    for i in range(70):
        y = H * i / 70
        alpha = 0.08 * (1 - abs((y / H) - 0.50) * 1.8)
        can.setFillColor(c("#0b2c35", max(0, alpha)))
        can.rect(0, y, W, H / 70 + 2, stroke=0, fill=1)

    for i in range(38):
        inset = i * 10
        alpha = min(0.012 + i * 0.0018, 0.08)
        can.setStrokeColor(c("#00040a", alpha))
        can.setLineWidth(18)
        can.roundRect(inset, inset, W - 2 * inset, H - 2 * inset, 0, stroke=1, fill=0)

    # Quiet grid, like the reference poster.
    can.setLineWidth(0.45)
    can.setStrokeColor(c("#123049", 0.35))
    step = 38
    x = 0
    while x <= W:
        can.line(x, 0, x, H)
        x += step
    y = 0
    while y <= H:
        can.line(0, y, W, y)
        y += step

    # Tiny deterministic star/sensor dust.
    can.setFillColor(c("#92eaff", 0.22))
    for i in range(360):
        x = (math.sin(i * 12.9898) * 43758.5453) % 1 * W
        y = (math.sin(i * 78.233) * 24634.6345) % 1 * H
        r = 0.7 if i % 5 else 1.1
        can.circle(x, y, r, stroke=0, fill=1)

    # Crop corner marks.
    m = 34
    l = 34
    can.setStrokeColor(c("#aac5ca", 0.42))
    can.setLineWidth(1.8)
    for sx in (m, W - m):
        for sy in (m, H - m):
            dx = l if sx < W / 2 else -l
            dy = l if sy < H / 2 else -l
            can.line(sx, sy, sx + dx, sy)
            can.line(sx, sy, sx, sy + dy)


def round_rect(can: canvas.Canvas, x: float, y: float, w: float, h: float, r: float = 14, fill=CARD, stroke=LINE, alpha: float = 0.94) -> None:
    can.setFillColor(colors.Color(fill.red, fill.green, fill.blue, alpha=alpha))
    can.setStrokeColor(c("#21445d", 0.65))
    can.setLineWidth(1.1)
    can.roundRect(x, y, w, h, r, stroke=1, fill=1)
    can.setStrokeColor(c("#1ce9ff", 0.23))
    can.setLineWidth(1.0)
    can.line(x + 18, y + h, x + w - 18, y + h)


def draw_wrapped(can: canvas.Canvas, text: str, x: float, y: float, width: float, size: float, leading: float, color=MUTED, font=FONT, max_lines: int | None = None) -> float:
    words = text.split()
    lines: list[str] = []
    current = ""
    for word in words:
        test = word if not current else current + " " + word
        if text_width(test, size, font) <= width:
            current = test
        else:
            if current:
                lines.append(current)
            current = word
    if current:
        lines.append(current)
    if max_lines is not None:
        lines = lines[:max_lines]
    can.setFont(font, size)
    can.setFillColor(color)
    for line in lines:
        can.drawString(x, y, line)
        y -= leading
    return y


def label(can: canvas.Canvas, text: str, x: float, y: float, color=CYAN, size: float = 10) -> None:
    can.setFillColor(color)
    can.setFont(MONO, size)
    can.drawString(x, y, text.upper())


def polyline(can: canvas.Canvas, pts: list[tuple[float, float]]) -> None:
    can.lines([(a[0], a[1], b[0], b[1]) for a, b in zip(pts, pts[1:])])


def draw_logo(can: canvas.Canvas, cx: float, cy: float, scale: float) -> None:
    can.saveState()
    can.translate(cx, cy)
    can.setLineCap(1)
    can.setLineWidth(18 * scale)
    can.setStrokeColor(CYAN)
    can.arc(-42 * scale, -42 * scale, 42 * scale, 42 * scale, 68, 278)
    can.setStrokeColor(YELLOW)
    can.arc(-42 * scale, -42 * scale, 42 * scale, 42 * scale, 248, 98)
    can.setLineWidth(5 * scale)
    can.setStrokeColor(c("#8ffaff", 0.25))
    can.circle(0, 0, 38 * scale, stroke=1, fill=0)
    can.restoreState()


def draw_hero(can: canvas.Canvas) -> None:
    can.setFont(MONO, 16)
    can.setFillColor(c("#b1d8df", 0.78))
    top = H - 120
    can.drawCentredString(W / 2, top, "NUS ORBITAL 2026        APOLLO")
    can.setFillColor(CYAN)
    can.circle(W / 2 - 190, top + 4, 3.5, stroke=0, fill=1)
    can.setFillColor(YELLOW)
    can.circle(W / 2 + 178, top + 4, 3.5, stroke=0, fill=1)

    draw_logo(can, W / 2, H - 285, 2.0)

    title_y = H - 445
    can.setFont(BOLD, 75)
    left = W / 2 - text_width("O(Alpha)", 75, BOLD) / 2
    can.setFillColor(YELLOW)
    can.drawString(left, title_y, "O")
    can.setFillColor(INK)
    can.drawString(left + text_width("O", 75, BOLD), title_y, "(Alpha)")

    can.setFont(FONT, 44)
    can.setFillColor(c("#d7f7ff", 0.88))
    can.drawCentredString(W / 2, title_y - 74, "Quant in your pocket.")
    can.setFont(FONT, 18)
    can.setFillColor(c("#bad5dc", 0.78))
    can.drawCentredString(
        W / 2,
        title_y - 124,
        "Validation-gated portfolio research, paper execution, and full audit trails in one system.",
    )

    strip_y = title_y - 195
    chips = ["VOO CORE", "H63 RANKER", "DSR/PBO GATE", "PAPER ONLY", "DB FILLS", "ALERTS", "PIT AUDIT"]
    total = sum(text_width(s, 11, MONO) + 56 for s in chips)
    x = W / 2 - total / 2
    for i, s in enumerate(chips):
        w = text_width(s, 11, MONO) + 40
        can.setFillColor(c("#0b1a29", 0.78))
        can.roundRect(x, strip_y - 11, w, 30, 15, stroke=0, fill=1)
        can.setFillColor(CYAN if i % 2 == 0 else YELLOW)
        can.circle(x + 15, strip_y + 4, 2.7, stroke=0, fill=1)
        can.setFillColor(c("#a7c8d3", 0.86))
        can.setFont(MONO, 11)
        can.drawString(x + 24, strip_y, s)
        x += w + 16


def draw_feature_cards(can: canvas.Canvas) -> None:
    y = H - 875
    card_w = 360
    gap = 34
    x0 = (W - card_w * 3 - gap * 2) / 2
    cards = [
        ("01", "Configure risk", "Choose a risk profile; the API maps it to a matching catalog bucket and a 1Day portfolio run."),
        ("02", "Select catalog alpha", "Low-risk h63 LGBM and deterministic proxy entries pass both primary and shifted validation windows."),
        ("03", "Track the paper trail", "Fills, positions, snapshots, alerts, and active runs are persisted for the dashboard."),
    ]
    icons = ["sliders", "trend", "audit"]
    for i, (num, title, body) in enumerate(cards):
        x = x0 + i * (card_w + gap)
        round_rect(can, x, y, card_w, 166, 18, CARD)
        can.setFillColor(c("#0b2636", 1))
        can.roundRect(x + 26, y + 96, 44, 44, 8, stroke=0, fill=1)
        can.setStrokeColor(CYAN)
        can.setLineWidth(3)
        if icons[i] == "sliders":
            for yy, knob in [(126, 48), (115, 34), (104, 56)]:
                can.line(x + 34, y + yy, x + 62, y + yy)
                can.circle(x + knob, y + yy, 3.5, stroke=1, fill=0)
        elif icons[i] == "trend":
            pts = [(x + 35, y + 104), (x + 44, y + 119), (x + 53, y + 111), (x + 62, y + 134)]
            polyline(can, pts)
            can.setFillColor(YELLOW)
            can.circle(x + 44, y + 119, 4, stroke=0, fill=1)
        else:
            can.circle(x + 48, y + 119, 15, stroke=1, fill=0)
            can.line(x + 59, y + 108, x + 67, y + 100)
        can.setFillColor(c("#617e8f", 0.85))
        can.setFont(MONO, 10)
        can.drawRightString(x + card_w - 25, y + 127, num)
        can.setFillColor(INK)
        can.setFont(BOLD, 22)
        can.drawString(x + 26, y + 73, title)
        draw_wrapped(can, body, x + 26, y + 45, card_w - 52, 13.5, 18, c("#abc4cf", 0.86), FONT, 3)


def draw_dashboard(can: canvas.Canvas) -> None:
    main_x, main_y = W / 2 - 350, 940
    main_w, main_h = 700, 368
    side_w, side_h = 250, 196

    # Side panels partly behind the main screen.
    right_x = main_x + main_w - 10
    round_rect(can, right_x, main_y + 150, side_w, side_h, 18, CARD_2, alpha=0.92)
    label(can, "LOW RISK PASS", right_x + 28, main_y + 308, CYAN, 10)
    can.setFont(MONO, 13)
    can.setFillColor(c("#9db9c4", 0.75))
    can.drawString(right_x + 28, main_y + 274, "LGBM LOW")
    can.drawString(right_x + 126, main_y + 274, "PROXY LOW")
    can.setFont(FONT, 28)
    can.setFillColor(INK)
    can.drawString(right_x + 28, main_y + 234, "DSR 1.000")
    can.setFillColor(YELLOW)
    can.drawString(right_x + 28, main_y + 194, "PBO .067/.077")
    can.setStrokeColor(CYAN)
    can.setLineWidth(4)
    chart = [
        (right_x + 34, main_y + 174),
        (right_x + 60, main_y + 192),
        (right_x + 80, main_y + 184),
        (right_x + 102, main_y + 216),
        (right_x + 126, main_y + 206),
        (right_x + 154, main_y + 242),
        (right_x + 184, main_y + 263),
    ]
    polyline(can, chart)

    round_rect(can, main_x - 210, main_y + 15, 230, 160, 18, CARD_2, alpha=0.9)
    label(can, "RUN SAFETY", main_x - 178, main_y + 132, CYAN, 10)
    rows = [("PAPER", "true", CYAN), ("ORDERS", "0", YELLOW), ("BROKER", "off", PINK)]
    yy = main_y + 96
    for name, val, col in rows:
        can.setFillColor(col)
        can.circle(main_x - 170, yy + 4, 3.2, stroke=0, fill=1)
        can.setFont(MONO, 12)
        can.setFillColor(c("#a9c4cc", 0.86))
        can.drawString(main_x - 156, yy, name)
        can.setFillColor(col)
        can.drawRightString(main_x - 8, yy, val)
        yy -= 30

    round_rect(can, main_x, main_y, main_w, main_h, 16, colors.HexColor("#0f1d30"), alpha=0.98)
    label(can, "O(ALPHA) / PORTFOLIO CATALOG", main_x + 34, main_y + main_h - 44, CYAN, 10)
    can.setFillColor(c("#84f3ff", 0.9))
    can.circle(main_x + main_w - 90, main_y + main_h - 38, 4, stroke=0, fill=1)
    can.setFont(MONO, 10)
    can.drawString(main_x + main_w - 78, main_y + main_h - 43, "LIVE PAPER")

    label(can, "AGENT STATUS  -  OPTIMISING", main_x + 34, main_y + main_h - 92, CYAN, 12)
    can.setFillColor(c("#4a3511", 0.8))
    can.roundRect(main_x + main_w - 174, main_y + main_h - 106, 134, 28, 14, stroke=0, fill=1)
    can.setFont(MONO, 10)
    can.setFillColor(YELLOW)
    can.drawCentredString(main_x + main_w - 107, main_y + main_h - 97, "REGIME - AUDITED")

    can.setFont(FONT, 59)
    can.setFillColor(INK)
    can.drawString(main_x + 34, main_y + main_h - 168, "$100,000")
    can.setFont(MONO, 13)
    can.setFillColor(c("#86b5c4", 0.7))
    can.drawString(main_x + 316, main_y + main_h - 147, "DEFAULT PAPER CASH")

    # Equity/history trace.
    base_y = main_y + 122
    can.setStrokeColor(c("#28445c", 0.8))
    can.setLineWidth(1)
    for i in range(4):
        can.line(main_x + 34, base_y + i * 28, main_x + main_w - 34, base_y + i * 28)
    can.setStrokeColor(CYAN)
    can.setLineWidth(3)
    pts = []
    for i in range(22):
        x = main_x + 34 + i * ((main_w - 68) / 21)
        y = base_y + 28 + math.sin(i * 0.75) * 12 + (i % 5) * 4 - i * 0.4
        pts.append((x, y))
    polyline(can, pts)
    can.setFillColor(CYAN)
    can.circle(*pts[-1], 5, stroke=0, fill=1)

    # Sliders.
    slider_y = main_y + 58
    labels = [("RISK TOLERANCE", "LOW", 0.33, CYAN), ("STOP LOSS", "2.5%", 0.42, PINK), ("REBALANCE", "63 BARS", 0.72, YELLOW)]
    for i, (name, value, frac, col) in enumerate(labels):
        x = main_x + 34 + i * 205
        can.setFont(MONO, 8.5)
        can.setFillColor(c("#7797a5", 0.8))
        can.drawString(x, slider_y + 24, name)
        can.setFillColor(col)
        can.drawRightString(x + 158, slider_y + 24, value)
        can.setStrokeColor(c("#20384c", 1))
        can.setLineWidth(3)
        can.line(x, slider_y, x + 160, slider_y)
        can.setStrokeColor(col)
        can.line(x, slider_y, x + 160 * frac, slider_y)
        can.setFillColor(INK)
        can.circle(x + 160 * frac, slider_y, 5, stroke=0, fill=1)


def draw_system_flow(can: canvas.Canvas) -> None:
    y = 500
    label(can, "SYSTEM ARCHITECTURE - RESEARCH TO PAPER EXECUTION", 85, y + 126, CYAN, 10)
    nodes = [
        ("Research", "reports/batches"),
        ("Catalog", "risk bucket"),
        ("Agent", "1Day loop"),
        ("Router", "idempotent fills"),
        ("Postgres", "state + alerts"),
        ("Dashboard", "live reads"),
    ]
    x0 = 126
    gap = 220
    prev = None
    for i, (name, sub) in enumerate(nodes):
        x = x0 + i * gap
        can.setFillColor(c("#102133", 0.92))
        can.circle(x, y + 48, 31, stroke=0, fill=1)
        can.setStrokeColor(CYAN if i % 2 == 0 else YELLOW)
        can.setLineWidth(2.3)
        can.circle(x, y + 48, 31, stroke=1, fill=0)
        can.setFillColor(CYAN if i % 2 == 0 else YELLOW)
        can.circle(x, y + 48, 6, stroke=0, fill=1)
        can.setFillColor(INK)
        can.setFont(BOLD, 14)
        can.drawCentredString(x, y - 2, name)
        can.setFillColor(MUTED)
        can.setFont(MONO, 8.5)
        can.drawCentredString(x, y - 20, sub)
        if prev is not None:
            can.setStrokeColor(c("#4eddec", 0.38))
            can.setLineWidth(1.4)
            can.line(prev + 44, y + 48, x - 44, y + 48)
        prev = x


def draw_tech_and_footer(can: canvas.Canvas) -> None:
    label(can, "POWERED BY", W / 2 - 43, 258, c("#879eaa", 0.72), 9)
    tech = ["Go", "Gin", "Next.js", "React", "TypeScript", "PostgreSQL", "TimescaleDB", "Redis", "Docker", "Vercel", "Alpaca", "Yahoo"]
    x = 0
    widths = [text_width(t, 10, FONT) + 30 for t in tech]
    total = sum(widths) + (len(tech) - 1) * 10
    x = W / 2 - total / 2
    y = 218
    for t, w in zip(tech, widths):
        can.setFillColor(c("#111d2c", 0.96))
        can.roundRect(x, y, w, 25, 12, stroke=0, fill=1)
        can.setFillColor(c("#a9c2cb", 0.88))
        can.setFont(FONT, 10)
        can.drawCentredString(x + w / 2, y + 8, t)
        x += w + 10

    # Evidence and footers.
    can.setStrokeColor(c("#1b3548", 0.85))
    can.setLineWidth(1)
    can.line(78, 150, W - 78, 150)
    draw_logo(can, 112, 92, 0.72)
    can.setFont(BOLD, 15)
    can.setFillColor(INK)
    can.drawString(145, 98, "Built by Zhao Shizhen & Tan Jia Jun")
    can.setFont(MONO, 9)
    can.setFillColor(c("#82a5b4", 0.75))
    can.drawString(145, 78, "Team 6645  /  O(Alpha)")

    can.setFont(MONO, 7.2)
    can.setFillColor(c("#7694a0", 0.75))
    source = (
        "Evidence: README.md:3-13,23-32,218-223  |  "
        "agent_catalog_bucket_comparison.md:13-17,36-41,65-70  |  "
        "paper_ranker_signal.md:3-15,17-22  |  "
        "portfolio_agent_handler.go:13-24,47-146  |  db_execution_router.go:41-145"
    )
    can.drawRightString(W - 78, 98, "A1 portrait 594 x 841 mm")
    can.drawRightString(W - 78, 78, "Paper-only research system. No live brokerage orders.")
    can.drawCentredString(W / 2, 42, source)


def build() -> None:
    OUT.parent.mkdir(parents=True, exist_ok=True)
    can = canvas.Canvas(str(OUT), pagesize=A1)
    can.setTitle("O(Alpha) A1 Orbital Poster")
    draw_bg(can)
    draw_hero(can)
    draw_feature_cards(can)
    draw_dashboard(can)
    draw_system_flow(can)
    draw_tech_and_footer(can)
    can.showPage()
    can.save()
    print(OUT)


if __name__ == "__main__":
    build()
