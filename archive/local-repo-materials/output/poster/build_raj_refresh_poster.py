from __future__ import annotations

import math
import random
from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter, ImageFont


ROOT = Path("/Users/shizhen/Documents/O-Alpha")
OUT_DIR = ROOT / "output" / "poster"
DOWNLOADS = Path("/Users/shizhen/Downloads")

W, H = 3508, 4967
BG = (5, 18, 30)
PANEL = (16, 31, 50, 238)
PANEL_2 = (11, 24, 41, 238)
LINE = (41, 92, 115)
CYAN = (16, 221, 238)
YELLOW = (255, 209, 35)
GREEN = (66, 230, 164)
PINK = (255, 92, 154)
PURPLE = (150, 125, 255)
WHITE = (229, 247, 253)
MUTED = (150, 181, 195)
DIM = (94, 124, 140)


FONT_CANDIDATES = {
    "avenir": "/System/Library/Fonts/Avenir Next.ttc",
    "avenir_cond": "/System/Library/Fonts/Avenir Next Condensed.ttc",
    "mono": "/System/Library/Fonts/SFNSMono.ttf",
}


def font(name: str, size: int, idx: int = 0) -> ImageFont.FreeTypeFont:
    try:
        return ImageFont.truetype(FONT_CANDIDATES[name], size, index=idx)
    except Exception:
        return ImageFont.load_default()


F = {
    "display": lambda s: font("avenir", s, 5),
    "display_reg": lambda s: font("avenir", s, 0),
    "body": lambda s: font("avenir", s, 0),
    "body_med": lambda s: font("avenir", s, 5),
    "cond": lambda s: font("avenir_cond", s, 5),
    "mono": lambda s: font("mono", s),
}


def text_wh(draw: ImageDraw.ImageDraw, text: str, fnt: ImageFont.FreeTypeFont) -> tuple[int, int]:
    box = draw.textbbox((0, 0), text, font=fnt)
    return box[2] - box[0], box[3] - box[1]


def wrap(draw: ImageDraw.ImageDraw, text: str, fnt: ImageFont.FreeTypeFont, max_w: int, max_lines: int | None = None) -> list[str]:
    words = text.split()
    lines: list[str] = []
    line = ""
    for word in words:
        candidate = word if not line else f"{line} {word}"
        if text_wh(draw, candidate, fnt)[0] <= max_w:
            line = candidate
        else:
            if line:
                lines.append(line)
            line = word
    if line:
        lines.append(line)
    if max_lines is not None:
        lines = lines[:max_lines]
    return lines


def draw_wrapped(
    draw: ImageDraw.ImageDraw,
    xy: tuple[int, int],
    text: str,
    fnt: ImageFont.FreeTypeFont,
    fill: tuple[int, int, int],
    max_w: int,
    line_gap: int = 8,
    max_lines: int | None = None,
) -> int:
    x, y = xy
    lines = wrap(draw, text, fnt, max_w, max_lines)
    _, line_h = text_wh(draw, "Ag", fnt)
    for line in lines:
        draw.text((x, y), line, font=fnt, fill=fill)
        y += line_h + line_gap
    return y


def rounded(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int], radius: int, fill, outline=None, width: int = 1) -> None:
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def glow(img: Image.Image, center: tuple[int, int], radius: int, color: tuple[int, int, int], alpha: int = 70, blur: int = 95) -> None:
    layer = Image.new("RGBA", img.size, (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    x, y = center
    d.ellipse((x - radius, y - radius, x + radius, y + radius), fill=(*color, alpha))
    img.alpha_composite(layer.filter(ImageFilter.GaussianBlur(blur)))


def draw_background(img: Image.Image, draw: ImageDraw.ImageDraw) -> None:
    draw.rectangle((0, 0, W, H), fill=BG)

    # Subtle vertical glow, grid, and sensor dust.
    for i in range(90):
        y = int(H * i / 90)
        a = max(0, int(16 * (1 - abs((y / H) - 0.46) * 1.9)))
        if a:
            layer = Image.new("RGBA", img.size, (0, 0, 0, 0))
            ImageDraw.Draw(layer).rectangle((0, y, W, y + H // 90 + 2), fill=(8, 50, 62, a))
            img.alpha_composite(layer)

    grid = Image.new("RGBA", img.size, (0, 0, 0, 0))
    gd = ImageDraw.Draw(grid)
    for x in range(0, W, 110):
        gd.line((x, 0, x, H), fill=(25, 68, 91, 48), width=1)
    for y in range(0, H, 110):
        gd.line((0, y, W, y), fill=(25, 68, 91, 42), width=1)
    img.alpha_composite(grid)

    random.seed(6645)
    for _ in range(1200):
        x = random.randrange(70, W - 70)
        y = random.randrange(70, H - 70)
        col = (110, 230, 245, random.randrange(24, 74))
        draw.point((x, y), fill=col)

    # A1 crop marks.
    m, l = 110, 70
    for x in (m, W - m):
        for y in (m, H - m):
            dx = l if x < W / 2 else -l
            dy = l if y < H / 2 else -l
            draw.line((x, y, x + dx, y), fill=(142, 179, 190), width=3)
            draw.line((x, y, x, y + dy), fill=(142, 179, 190), width=3)


def draw_brand(img: Image.Image, x: int, y: int, size: int) -> None:
    brand_path = ROOT / "frontend" / "public" / "brand-mark.png"
    glow(img, (x + size // 2, y + size // 2), size // 2, CYAN, 75, 80)
    mark = Image.open(brand_path).convert("RGBA")
    mark.thumbnail((size, size), Image.Resampling.LANCZOS)
    img.alpha_composite(mark, (x + (size - mark.width) // 2, y + (size - mark.height) // 2))


def label(draw: ImageDraw.ImageDraw, xy: tuple[int, int], text: str, color=CYAN, size: int = 25) -> None:
    draw.text(xy, text.upper(), font=F["mono"](size), fill=color)


def pill(draw: ImageDraw.ImageDraw, x: int, y: int, text: str, color=CYAN, size: int = 24) -> int:
    fnt = F["mono"](size)
    tw, th = text_wh(draw, text, fnt)
    w = tw + 70
    rounded(draw, (x, y, x + w, y + 48), 24, (17, 34, 54, 235), None, 0)
    draw.ellipse((x + 22, y + 18, x + 32, y + 28), fill=color)
    draw.text((x + 46, y + 13), text, font=fnt, fill=(176, 209, 220))
    return w


def card(
    draw: ImageDraw.ImageDraw,
    box: tuple[int, int, int, int],
    eyebrow: str,
    title: str,
    body: str,
    accent=CYAN,
    number: str | None = None,
) -> None:
    x0, y0, x1, y1 = box
    rounded(draw, box, 36, PANEL, LINE, 3)
    draw.line((x0 + 55, y0, x1 - 55, y0), fill=accent, width=5)
    if number:
        draw.text((x1 - 58, y0 + 40), number, font=F["mono"](22), fill=(92, 125, 143), anchor="ra")
    draw.ellipse((x0 + 54, y0 + 52, x0 + 70, y0 + 68), fill=accent)
    label(draw, (x0 + 88, y0 + 44), eyebrow, accent, 23)
    draw.text((x0 + 54, y0 + 98), title, font=F["body_med"](46), fill=WHITE)
    draw_wrapped(draw, (x0 + 54, y0 + 172), body, F["body_med"](29), MUTED, x1 - x0 - 108, 10, 4)


def mini_stat(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int], label_text: str, value: str, note: str, accent=CYAN) -> None:
    x0, y0, x1, y1 = box
    rounded(draw, box, 26, (11, 25, 43, 242), (32, 74, 96), 2)
    label(draw, (x0 + 34, y0 + 28), label_text, DIM, 19)
    draw.text((x0 + 34, y0 + 72), value, font=F["display"](48), fill=accent)
    draw_wrapped(draw, (x0 + 34, y0 + 132), note, F["body_med"](23), MUTED, x1 - x0 - 68, 7, 2)


def line_chart(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int], color=CYAN, width: int = 8) -> None:
    x0, y0, x1, y1 = box
    vals = [0.18, 0.25, 0.22, 0.33, 0.30, 0.42, 0.38, 0.48, 0.54, 0.46, 0.59, 0.63, 0.57, 0.71]
    pts = []
    for i, v in enumerate(vals):
        x = x0 + (x1 - x0) * i / (len(vals) - 1)
        y = y1 - (y1 - y0) * v
        pts.append((x, y))
    draw.line((x0, y1, x1, y1), fill=(46, 84, 105), width=3)
    draw.line(pts, fill=(0, 240, 255, 90), width=width + 9, joint="curve")
    draw.line(pts, fill=color, width=width, joint="curve")
    x, y = pts[-1]
    draw.ellipse((x - 15, y - 15, x + 15, y + 15), fill=color)


def draw_dashboard(draw: ImageDraw.ImageDraw) -> None:
    x0, y0, x1, y1 = 640, 2025, 2870, 2815
    rounded(draw, (x0, y0, x1, y1), 44, (13, 28, 47, 248), (37, 91, 116), 3)
    label(draw, (x0 + 70, y0 + 58), "paper-agent command center", CYAN, 24)
    rounded(draw, (x1 - 315, y0 + 46, x1 - 72, y0 + 92), 23, (18, 53, 64, 240), None, 0)
    draw.text((x1 - 194, y0 + 58), "PAPER ONLY", font=F["mono"](21), fill=CYAN, anchor="ma")

    draw.text((x0 + 70, y0 + 150), "One active run per user", font=F["body_med"](42), fill=WHITE)
    draw_wrapped(
        draw,
        (x0 + 70, y0 + 218),
        "The worker evaluates daily bars, turns target weights into idempotent paper fills, and persists positions, cash ledger, snapshots, trades, and alerts.",
        F["body_med"](30),
        MUTED,
        850,
        10,
        4,
    )

    draw.text((x0 + 70, y0 + 412), "$100,000", font=F["display_reg"](112), fill=WHITE)
    draw.text((x0 + 78, y0 + 545), "DEFAULT PAPER CASH", font=F["mono"](27), fill=DIM)
    line_chart(draw, (x0 + 1120, y0 + 175, x1 - 90, y0 + 510), CYAN, 8)

    metrics = [
        ("RISK", "LOW", CYAN),
        ("STOP LOSS", "2.5%", PINK),
        ("REBALANCE", "63 bars", YELLOW),
    ]
    for idx, (k, v, col) in enumerate(metrics):
        sx = x0 + 78 + idx * 640
        sy = y1 - 145
        label(draw, (sx, sy - 62), k, DIM, 19)
        draw.line((sx, sy, sx + 410, sy), fill=(50, 87, 106), width=8)
        draw.line((sx, sy, sx + 250 + idx * 36, sy), fill=col, width=8)
        draw.ellipse((sx + 245 + idx * 36, sy - 16, sx + 277 + idx * 36, sy + 16), fill=WHITE)
        draw.text((sx + 438, sy - 17), v, font=F["mono"](27), fill=col)


def feature_group_panel(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int]) -> None:
    x0, y0, x1, y1 = box
    rounded(draw, box, 34, PANEL, LINE, 3)
    label(draw, (x0 + 48, y0 + 44), "feature groups", CYAN, 25)
    rows = [
        ("ACCESS", "Login, risk profile, backtest acceptance gate", CYAN),
        ("RESEARCH", "Harness reports, DSR/PBO, cost stress, no-lookahead", YELLOW),
        ("EXECUTION", "Paper fills, positions, cash ledger, snapshots", GREEN),
        ("OBSERVE", "Dashboard state, trades, alerts, exports", PURPLE),
        ("SAFETY", "One active agent, guarded settings, fail-closed artifacts", PINK),
    ]
    row_h = 111
    y = y0 + 116
    for name, body, col in rows:
        rounded(draw, (x0 + 48, y, x1 - 48, y + 82), 18, (14, 29, 48, 240), None, 0)
        draw.rectangle((x0 + 48, y, x0 + 57, y + 82), fill=col)
        draw.text((x0 + 82, y + 21), name, font=F["mono"](24), fill=col)
        draw.text((x0 + 315, y + 21), body, font=F["body_med"](27), fill=MUTED)
        y += row_h


def evidence_panel(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int]) -> None:
    x0, y0, x1, y1 = box
    rounded(draw, box, 34, PANEL, LINE, 3)
    label(draw, (x0 + 48, y0 + 44), "testing & evidence", CYAN, 25)
    items = [
        ("GO TESTS", "54 backend test files; `go test ./...` passed", GREEN),
        ("TYPECHECK", "`npm run typecheck` passed for the frontend", CYAN),
        ("RESEARCH", "Metrics cite committed `reports/batches` artifacts", YELLOW),
        ("PLANNED", "Add integration + UI smoke tests before M3", PINK),
    ]
    y = y0 + 125
    for title, body, col in items:
        draw.ellipse((x0 + 54, y + 9, x0 + 72, y + 27), fill=col)
        draw.text((x0 + 92, y), title, font=F["mono"](25), fill=col)
        draw_wrapped(draw, (x0 + 92, y + 38), body, F["body_med"](28), MUTED, x1 - x0 - 145, 8, 2)
        y += 135


def motivation_panel(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int]) -> None:
    x0, y0, x1, y1 = box
    rounded(draw, box, 34, PANEL, LINE, 3)
    label(draw, (x0 + 48, y0 + 44), "motivation", CYAN, 25)
    draw.text((x0 + 48, y0 + 105), "The problem", font=F["body_med"](38), fill=PINK)
    draw_wrapped(
        draw,
        (x0 + 48, y0 + 160),
        "Retail traders see charts and signals, but rarely see whether a strategy survived costs, out-of-sample checks, overfitting tests, and safe execution constraints.",
        F["body_med"](29),
        MUTED,
        x1 - x0 - 96,
        9,
        5,
    )
    draw.text((x0 + 48, y0 + 360), "Our solution", font=F["body_med"](38), fill=GREEN)
    draw_wrapped(
        draw,
        (x0 + 48, y0 + 415),
        "A paper-only platform where strategy research, user onboarding, paper execution, portfolio state, and audit logs are connected through one backend.",
        F["body_med"](29),
        MUTED,
        x1 - x0 - 96,
        9,
        5,
    )
    draw.text((x0 + 48, y1 - 165), "Reviewer fix", font=F["body_med"](34), fill=YELLOW)
    draw_wrapped(
        draw,
        (x0 + 48, y1 - 112),
        "This poster now foregrounds software engineering: architecture, tests, lifecycle safety, database design, and artifact-backed validation.",
        F["body_med"](27),
        MUTED,
        x1 - x0 - 96,
        8,
        3,
    )


def architecture_flow(draw: ImageDraw.ImageDraw) -> None:
    x0, x1 = 220, W - 220
    y = 4300
    label(draw, (x0, y - 145), "system architecture - research to paper execution", CYAN, 27)
    nodes = [
        ("Research", "cmd/alpha-research"),
        ("Reports", "JSON + MD artifacts"),
        ("Catalog", "risk-bucket strategy"),
        ("Agent", "daily worker loop"),
        ("Router", "idempotent fills"),
        ("Postgres", "ledger + snapshots"),
        ("Dashboard", "state + alerts"),
    ]
    gap = (x1 - x0) / (len(nodes) - 1)
    for i, (name, sub) in enumerate(nodes):
        cx = int(x0 + i * gap)
        col = [CYAN, YELLOW, GREEN, PURPLE, CYAN, YELLOW, PINK][i]
        if i > 0:
            prev = int(x0 + (i - 1) * gap)
            draw.line((prev + 80, y, cx - 80, y), fill=(50, 109, 129), width=5)
            draw.polygon([(cx - 82, y), (cx - 112, y - 16), (cx - 112, y + 16)], fill=(50, 109, 129))
        rounded(draw, (cx - 63, y - 63, cx + 63, y + 63), 30, (15, 31, 51, 245), col, 4)
        draw.ellipse((cx - 15, y - 15, cx + 15, y + 15), fill=col)
        draw.text((cx, y + 97), name, font=F["body_med"](31), fill=WHITE, anchor="ma")
        draw_wrapped(draw, (cx - 105, y + 140), sub, F["mono"](19), DIM, 210, 3, 2)


def draw_footer(img: Image.Image, draw: ImageDraw.ImageDraw) -> None:
    y = H - 260
    draw.line((150, y - 70, W - 150, y - 70), fill=(35, 76, 98), width=2)
    draw_brand(img, 160, y - 5, 92)
    draw.text((280, y + 18), "Built by Zhao Shizhen & Tan Jia Jun", font=F["body_med"](29), fill=WHITE)
    draw.text((280, y + 62), "Team 6645  /  O(Alpha)  /  Milestone 2 refresh", font=F["mono"](20), fill=DIM)
    draw.text((W - 170, y + 18), "Paper-only research system. No live brokerage orders.", font=F["mono"](22), fill=MUTED, anchor="ra")
    draw.text((W - 170, y + 62), "A1 portrait 594 x 841 mm", font=F["mono"](20), fill=DIM, anchor="ra")


def build() -> None:
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    img = Image.new("RGBA", (W, H), (*BG, 255))
    draw = ImageDraw.Draw(img)
    draw_background(img, draw)

    # Top metadata.
    label(draw, (W - 480, 245), "Tan Jia Jun", MUTED, 22)
    label(draw, (W - 480, 285), "Zhao Shizhen", MUTED, 22)
    draw.text((W // 2, 240), "NUS ORBITAL 2026  /  APOLLO  /  TEAM 6645", font=F["mono"](28), fill=(155, 190, 204), anchor="ma")

    # Hero.
    draw_brand(img, W // 2 - 150, 390, 300)
    title = "O(Alpha)"
    title_font = F["display"](145)
    tw, _ = text_wh(draw, title, title_font)
    tx = W // 2 - tw // 2
    draw.text((tx, 760), "O", font=title_font, fill=YELLOW)
    draw.text((tx + text_wh(draw, "O", title_font)[0], 760), "(Alpha)", font=title_font, fill=WHITE)
    draw.text((W // 2, 930), "Quant in your pocket.", font=F["display_reg"](74), fill=WHITE, anchor="ma")
    draw.text(
        (W // 2, 1015),
        "Software-engineered strategy validation, paper execution, and audit-ready portfolio state.",
        font=F["body_med"](35),
        fill=MUTED,
        anchor="ma",
    )
    chip_texts = ["FAIL-CLOSED GATES", "DB AUDIT TRAIL", "WORKER LIFECYCLE", "CI + TESTS", "PAPER ONLY"]
    x = W // 2 - 930
    for idx, t in enumerate(chip_texts):
        x += pill(draw, x, 1105, t, [CYAN, YELLOW, GREEN, PURPLE, PINK][idx], 23) + 24

    # Top feature cards.
    card_w, card_h, gap = 960, 315, 72
    x_start = (W - (card_w * 3 + gap * 2)) // 2
    y = 1285
    cards = [
        ("01", "Access & onboarding", "Guided paper setup", "Login, choose risk profile, run a catalog backtest, then accept the result before the dashboard unlocks.", CYAN),
        ("02", "Validation harness", "Metrics are evidence", "Strategy results come from committed reports with DSR, PBO, cost stress, OOS trades, and benchmark checks.", YELLOW),
        ("03", "Paper execution", "No hidden live orders", "The worker writes simulated fills, positions, cash ledger, snapshots, trades, and alerts for inspection.", GREEN),
    ]
    for i, (num, eye, title, body, col) in enumerate(cards):
        card(draw, (x_start + i * (card_w + gap), y, x_start + i * (card_w + gap) + card_w, y + card_h), eye, title, body, col, num)

    # Evidence stats.
    stat_y = 1670
    stat_w = 705
    stats = [
        ("LOW-RISK PASS", "2 entries", "h63 LGBM + proxy pass both windows", GREEN),
        ("PBO GATE", ".067/.077", "low-risk catalog PBO across windows", YELLOW),
        ("BACKEND TESTS", "passed", "`go test ./...` verified locally", CYAN),
        ("FRONTEND", "typecheck", "`npm run typecheck` verified locally", PURPLE),
    ]
    sx = x_start
    for i, (k, v, note, col) in enumerate(stats):
        mini_stat(draw, (sx + i * (stat_w + 44), stat_y, sx + i * (stat_w + 44) + stat_w, stat_y + 210), k, v, note, col)

    draw_dashboard(draw)

    # Lower content, bigger and more SE-focused than the original poster.
    lower_y = 2950
    col_w = 990
    motivation_panel(draw, (220, lower_y, 220 + col_w, lower_y + 790))
    feature_group_panel(draw, (W // 2 - col_w // 2, lower_y, W // 2 + col_w // 2, lower_y + 790))
    evidence_panel(draw, (W - 220 - col_w, lower_y, W - 220, lower_y + 790))

    architecture_flow(draw)
    draw_footer(img, draw)

    png = OUT_DIR / "oalpha_poster_raj_refresh.png"
    jpg = OUT_DIR / "oalpha_poster_raj_refresh.jpg"
    pdf = OUT_DIR / "oalpha_poster_raj_refresh.pdf"
    img_rgb = img.convert("RGB")
    img_rgb.save(png, quality=95)
    img_rgb.save(jpg, quality=95, dpi=(150, 150))
    img_rgb.save(pdf, "PDF", resolution=150.0)

    # Also copy user-facing deliverables to Downloads.
    img_rgb.save(DOWNLOADS / "6645_raj_refresh.png", quality=95)
    img_rgb.save(DOWNLOADS / "6645_raj_refresh.jpg", quality=95, dpi=(150, 150))
    img_rgb.save(DOWNLOADS / "6645_raj_refresh.pdf", "PDF", resolution=150.0)

    print(png)
    print(jpg)
    print(pdf)
    print(DOWNLOADS / "6645_raj_refresh.jpg")


if __name__ == "__main__":
    build()
