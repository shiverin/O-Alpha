from PIL import Image, ImageDraw, ImageFont, ImageFilter
from pathlib import Path
import json
import random

ROOT = Path("/Users/shizhen/Documents/O-Alpha")
OUT = ROOT / "output/poster/drafts"
OUT.mkdir(parents=True, exist_ok=True)

W, H = 2480, 3508
CYAN = (18, 220, 238)
YELLOW = (255, 209, 38)
PINK = (255, 92, 168)
INK = (9, 17, 29)
WHITE = (238, 247, 252)
MUTED = (147, 172, 188)

FONT_CANDIDATES = {
    "avenir": "/System/Library/Fonts/Avenir Next.ttc",
    "avenir_cond": "/System/Library/Fonts/Avenir Next Condensed.ttc",
    "mono": "/System/Library/Fonts/SFNSMono.ttf",
}


def font(name, size, idx=0):
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

BRAND = Image.open(ROOT / "frontend/public/brand-mark.png").convert("RGBA")

CONTENT = {
    "eyebrow": "NUS ORBITAL 2026  /  APOLLO  /  TEAM 6645",
    "title": "O(Alpha)",
    "tagline": "Quant in your pocket.",
    "sub": "Validation-gated alpha research to auditable paper execution.",
    "cards": [
        ("1", "Validate", "Walk-forward folds, cost stress, DSR/PBO gates, and committed reports decide catalog eligibility."),
        ("2", "Select", "Risk profiles map to paper-only catalog buckets; low-risk h63 entries pass two validation windows."),
        ("3", "Run", "A daily-bar portfolio agent turns target weights into idempotent paper fills and snapshots."),
        ("4", "Inspect", "The dashboard reads live backend state: runs, regime, history, allocation, trades, and alerts."),
    ],
    "metrics": [
        ("DSR", "1.000"),
        ("PBO", ".067 / .077"),
        ("OOS", "49-103 trades"),
        ("Mode", "Paper only"),
    ],
    "pipeline": ["Research harness", "reports/batches", "Strategy catalog", "Portfolio agent", "DB router", "Postgres state", "Dashboard"],
    "stack": ["Go", "Gin", "Next.js", "React", "TypeScript", "PostgreSQL", "TimescaleDB", "Redis", "Docker", "Vercel", "Alpaca", "Yahoo"],
    "footer": "Evidence: reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_summary/agent_catalog_bucket_comparison.md  |  paper_ranker_signal.md  |  README.md",
}


def text_size(draw, text, fnt):
    box = draw.textbbox((0, 0), text, font=fnt)
    return box[2] - box[0], box[3] - box[1]


def wrap_to_width(draw, text, fnt, max_w):
    words = text.split()
    lines = []
    line = ""
    for word in words:
        test = word if not line else line + " " + word
        if text_size(draw, test, fnt)[0] <= max_w:
            line = test
        else:
            if line:
                lines.append(line)
            line = word
    if line:
        lines.append(line)
    return "\n".join(lines)


def rounded(draw, box, radius, fill, outline=None, width=1):
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def add_noise(img, amount=5, seed=1):
    random.seed(seed)
    px = img.load()
    for _ in range(12000):
        x = random.randrange(W)
        y = random.randrange(H)
        r, g, b = px[x, y][:3]
        d = random.randint(-amount, amount)
        px[x, y] = (max(0, min(255, r + d)), max(0, min(255, g + d)), max(0, min(255, b + d)))


def glow_circle(img, center, radius, color, alpha=80, blur=80):
    layer = Image.new("RGBA", img.size, (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    x, y = center
    d.ellipse((x - radius, y - radius, x + radius, y + radius), fill=(*color, alpha))
    img.alpha_composite(layer.filter(ImageFilter.GaussianBlur(blur)))


def draw_brand(img, x, y, size, glow=True):
    if glow:
        glow_circle(img, (x + size // 2, y + size // 2), size // 2, CYAN, 56, 68)
    mark = BRAND.copy()
    mark.thumbnail((size, size), Image.Resampling.LANCZOS)
    img.alpha_composite(mark, (x + (size - mark.width) // 2, y + (size - mark.height) // 2))


def line_chart(draw, box, color=CYAN, fill=None, thick=7):
    x0, y0, x1, y1 = box
    vals = [0.25, 0.35, 0.39, 0.38, 0.31, 0.20, 0.19, 0.25, 0.37, 0.52, 0.39, 0.44, 0.42, 0.36, 0.31, 0.14, 0.22, 0.35, 0.46, 0.54, 0.33]
    pts = []
    for i, v in enumerate(vals):
        x = x0 + (x1 - x0) * i / (len(vals) - 1)
        y = y1 - (y1 - y0) * v
        pts.append((x, y))
    if fill:
        draw.polygon([(x0, y1)] + pts + [(x1, y1)], fill=fill)
    draw.line(pts, fill=color, width=thick, joint="curve")
    x, y = pts[-1]
    draw.ellipse((x - 12, y - 12, x + 12, y + 12), fill=color)


def quadratic_points(p0, c, p1, steps=24):
    pts = []
    for i in range(steps + 1):
        t = i / steps
        x = (1 - t) * (1 - t) * p0[0] + 2 * (1 - t) * t * c[0] + t * t * p1[0]
        y = (1 - t) * (1 - t) * p0[1] + 2 * (1 - t) * t * c[1] + t * t * p1[1]
        pts.append((x, y))
    return pts


def dashboard_pnl_sparkline(draw, box):
    x0, y0, x1, y1 = box
    # Mirrors BalanceCard.tsx: viewBox 0 0 100 100, y = 85 - normalized * 70,
    # no area fill, rounded cyan stroke, glow, bottom border, endpoint dot.
    values = [
        100000.00,
        100018.20,
        100011.70,
        100045.30,
        100036.90,
        100062.50,
        100058.10,
        100081.40,
        100074.80,
        100103.65,
    ]
    min_val = min(values)
    max_val = max(values)
    val_range = max_val - min_val
    pts = []
    for i, value in enumerate(values):
        px = (i / (len(values) - 1)) * 100
        py = 85 - ((value - min_val) / val_range) * 70
        pts.append((x0 + (px / 100) * (x1 - x0), y0 + (py / 100) * (y1 - y0)))

    draw.line((x0, y1, x1, y1), fill=(42, 73, 92), width=2)
    draw.line(pts, fill=(0, 240, 255, 70), width=12, joint="curve")
    draw.line(pts, fill=(0, 240, 255, 255), width=4, joint="curve")
    lx, ly = pts[-1]
    draw.ellipse((lx - 14, ly - 14, lx + 14, ly + 14), fill=(0, 240, 255, 58))
    draw.ellipse((lx - 7, ly - 7, lx + 7, ly + 7), fill=(0, 240, 255, 255))


def card(draw, box, title, body, icon, accent=CYAN, theme="dark", body_size=24):
    x0, y0, x1, y1 = box
    fill = (18, 34, 55, 235) if theme == "dark" else (255, 255, 255, 238)
    outline = (34, 92, 117) if theme == "dark" else (205, 216, 222)
    text = WHITE if theme == "dark" else (17, 24, 39)
    muted = (145, 169, 184) if theme == "dark" else (88, 102, 117)
    rounded(draw, box, 32, fill, outline, 2)
    draw.text((x0 + 36, y0 + 26), icon, font=F["mono"](26), fill=accent)
    draw.text((x0 + 36, y0 + 72), title, font=F["body_med"](36), fill=text)
    body_wrapped = wrap_to_width(draw, body, F["body_med"](body_size), x1 - x0 - 72)
    draw.multiline_text((x0 + 36, y0 + 130), body_wrapped, font=F["body_med"](body_size), fill=muted, spacing=7)


def dashboard_panel(draw, box, variant="dark"):
    x0, y0, x1, y1 = box
    bg = (13, 26, 43, 245) if variant != "light" else (244, 249, 250, 245)
    outline = (31, 91, 119) if variant != "light" else (189, 203, 211)
    text = WHITE if variant != "light" else (18, 25, 38)
    muted = (129, 156, 174) if variant != "light" else (92, 108, 122)
    rounded(draw, box, 42, bg, outline, 2)
    draw.text((x0 + 48, y0 + 42), "O(ALPHA) / PORTFOLIO CATALOG", font=F["mono"](24), fill=CYAN)
    draw.text((x1 - 48, y0 + 42), "LIVE PAPER", font=F["mono"](24), fill=CYAN, anchor="ra")
    draw.text((x0 + 48, y0 + 112), "Agent status  -  Optimising", font=F["mono"](28), fill=CYAN)
    draw.text((x0 + 48, y0 + 178), "$100,000", font=F["display_reg"](110), fill=text)
    draw.text((x0 + 650, y0 + 218), "DEFAULT PAPER CASH", font=F["mono"](23), fill=muted)
    for i in range(4):
        y = y0 + 310 + i * 64
        draw.line((x0 + 48, y, x1 - 48, y), fill=(37, 70, 91) if variant != "light" else (212, 222, 226), width=2)
    line_chart(draw, (x0 + 48, y0 + 278, x1 - 48, y0 + 520), CYAN, fill=(18, 220, 238, 28), thick=7)
    labels = [("Risk", "LOW", CYAN), ("Stop", "2.5%", PINK), ("Rebal.", "63 bars", YELLOW)]
    usable = x1 - x0 - 120
    gap = usable / 3
    for idx, (label, val, col) in enumerate(labels):
        lx = x0 + 58 + int(idx * gap)
        draw.text((lx, y1 - 138), label.upper(), font=F["mono"](19), fill=muted)
        draw.line((lx, y1 - 74, lx + 210, y1 - 74), fill=(46, 84, 105) if variant != "light" else (189, 203, 211), width=5)
        draw.ellipse((lx + 86, y1 - 86, lx + 110, y1 - 62), fill=col)
        draw.text((lx + 245, y1 - 92), val, font=F["mono"](21), fill=col)


def evidence_card(draw, box, theme="dark"):
    x0, y0, x1, y1 = box
    fill = (16, 31, 50, 240) if theme == "dark" else (255, 255, 255, 242)
    outline = (37, 100, 124) if theme == "dark" else (198, 211, 219)
    text = WHITE if theme == "dark" else (17, 24, 39)
    muted = (139, 164, 180) if theme == "dark" else (91, 106, 121)
    rounded(draw, box, 36, fill, outline, 2)
    draw.text((x0 + 38, y0 + 36), "LOW-RISK PASS", font=F["mono"](24), fill=CYAN)
    draw.text((x0 + 38, y0 + 98), "LGBM h63 + proxy", font=F["body_med"](36), fill=text)
    my = y0 + 164
    for k, v in CONTENT["metrics"]:
        draw.text((x0 + 38, my), k, font=F["mono"](22), fill=muted)
        draw.text((x0 + 38, my + 36), v, font=F["body_med"](35), fill=YELLOW if k == "PBO" else text)
        my += 92
    line_chart(draw, (x1 - 320, y1 - 210, x1 - 54, y1 - 52), CYAN, thick=7)


def pipeline(draw, x, y, width, theme="dark"):
    n = len(CONTENT["pipeline"])
    text = WHITE if theme == "dark" else (18, 25, 38)
    for i, label in enumerate(CONTENT["pipeline"]):
        cx = x + width * i / (n - 1)
        col = CYAN if i % 2 == 0 else YELLOW
        draw.ellipse((cx - 34, y - 34, cx + 34, y + 34), outline=col, width=6)
        draw.ellipse((cx - 9, y - 9, cx + 9, y + 9), fill=col)
        if i < n - 1:
            nx = x + width * (i + 1) / (n - 1)
            draw.line((cx + 44, y, nx - 44, y), fill=(35, 78, 95) if theme == "dark" else (174, 193, 202), width=3)
        wrapped = wrap_to_width(draw, label, F["body_med"](23), 165)
        draw.multiline_text((cx, y + 60), wrapped, font=F["body_med"](23), fill=text, anchor="ma", align="center", spacing=4)


def tech_pills(draw, y, theme="dark"):
    widths = []
    total = 0
    for s in CONTENT["stack"]:
        tw, _ = text_size(draw, s, F["mono"](19))
        widths.append(tw + 40)
        total += tw + 40 + 13
    x = (W - total) // 2
    for s, wid in zip(CONTENT["stack"], widths):
        fill = (22, 39, 60) if theme == "dark" else (232, 239, 243)
        txt = (171, 194, 207) if theme == "dark" else (67, 78, 92)
        rounded(draw, (x, y, x + wid, y + 42), 20, fill)
        draw.text((x + 20, y + 12), s, font=F["mono"](19), fill=txt)
        x += wid + 13


def save(img, name):
    path = OUT / name
    img.convert("RGB").save(path, quality=95)
    return str(path)


def draw_footer(draw, img, theme="dark"):
    light = theme == "light"
    draw_brand(img, 115, H - 188, 85, False)
    draw.text((225, H - 155), "Built by Zhao Shizhen & Tan Jia Jun", font=F["body_med"](28), fill=INK if light else WHITE)
    draw.text((225, H - 114), "Team 6645  /  O(Alpha)", font=F["mono"](20), fill=(86, 103, 116) if light else MUTED)
    draw.text((W // 2, H - 66), CONTENT["footer"], font=F["mono"](14), fill=(104, 119, 130) if light else (92, 112, 127), anchor="ma")


def technical_card(draw, box, number, title, bullets, accent=CYAN, cols=1):
    x0, y0, x1, y1 = box
    rounded(draw, box, 28, (14, 29, 47, 242), (35, 89, 112), 2)
    draw.text((x0 + 28, y0 + 25), str(number), font=F["mono"](23), fill=accent)
    draw.text((x0 + 76, y0 + 23), title.upper(), font=F["body_med"](30), fill=WHITE)
    y = y0 + 82
    col_w = (x1 - x0 - 70) / cols
    for idx, item in enumerate(bullets):
        col = idx % cols
        row = idx // cols
        bx = x0 + 32 + int(col * col_w)
        by = y + row * 62
        draw.ellipse((bx, by + 8, bx + 13, by + 21), fill=accent if idx % 2 == 0 else YELLOW)
        wrapped = wrap_to_width(draw, item, F["body_med"](21), int(col_w - 34))
        draw.multiline_text((bx + 25, by), wrapped, font=F["body_med"](21), fill=(176, 197, 208), spacing=4)


def mini_schema(draw, box):
    x0, y0, x1, y1 = box
    rounded(draw, box, 24, (9, 19, 31, 235), (32, 91, 115), 2)
    draw.text((x0 + 28, y0 + 24), "POSTGRES STATE MODEL", font=F["mono"](22), fill=CYAN)
    tables = [
        ("agent_runs", "strategy_key, runtime_state, heartbeat"),
        ("orders", "client_order_id, status"),
        ("fills", "symbol, side, qty, price"),
        ("positions", "qty, avg_entry, mark"),
        ("snapshots", "cash, equity, pnl"),
        ("alerts", "severity, source, metadata"),
        ("bars", "symbol, timeframe, dataset"),
        ("ml_artifacts", "artifact_uri, promoted"),
    ]
    cols = 2
    cell_w = (x1 - x0 - 72) // cols
    for i, (name, desc) in enumerate(tables):
        cx = x0 + 28 + (i % cols) * (cell_w + 18)
        cy = y0 + 78 + (i // cols) * 74
        rounded(draw, (cx, cy, cx + cell_w, cy + 54), 14, (18, 36, 56, 235), (48, 92, 111), 1)
        draw.text((cx + 14, cy + 8), name, font=F["mono"](18), fill=YELLOW if i % 2 else CYAN)
        draw.text((cx + 14, cy + 30), desc, font=F["body"](17), fill=(145, 168, 181))


def architecture_boxes(draw, box):
    x0, y0, x1, y1 = box
    rounded(draw, box, 30, (14, 29, 47, 242), (35, 89, 112), 2)
    draw.text((x0 + 32, y0 + 25), "SYSTEM ARCHITECTURE", font=F["body_med"](32), fill=WHITE)
    draw.text((x1 - 32, y0 + 31), "research -> catalog -> paper execution", font=F["mono"](20), fill=CYAN, anchor="ra")
    labels = [
        ("CLI", "alpha-research\nbacktest\nranker parity"),
        ("Reports", "JSON + MD\ncommitted evidence\nfail-closed gates"),
        ("API", "Go/Gin auth\ncatalog routes\nstreamed backtest"),
        ("Agent", "daily bars\nHMM state\nruntime settings"),
        ("Router", "sell then buy\nidempotent fills\nalerts"),
        ("UI", "Next.js dashboard\nNDJSON live stream\nstate polling"),
    ]
    n = len(labels)
    y = y0 + 132
    for i, (head, body) in enumerate(labels):
        cx = x0 + 80 + i * ((x1 - x0 - 160) / (n - 1))
        color = CYAN if i % 2 == 0 else YELLOW
        rounded(draw, (cx - 95, y, cx + 95, y + 145), 20, (9, 19, 31, 240), (44, 88, 105), 2)
        draw.text((cx, y + 19), head, font=F["mono"](22), fill=color, anchor="ma")
        draw.multiline_text((cx, y + 58), body, font=F["body_med"](17), fill=(177, 198, 210), anchor="ma", align="center", spacing=3)
        if i < n - 1:
            nx = x0 + 80 + (i + 1) * ((x1 - x0 - 160) / (n - 1))
            draw.line((cx + 102, y + 72, nx - 102, y + 72), fill=(61, 100, 116), width=3)
            draw.polygon([(nx - 112, y + 64), (nx - 96, y + 72), (nx - 112, y + 80)], fill=(61, 100, 116))


def compact_dashboard(draw, box):
    x0, y0, x1, y1 = box
    rounded(draw, box, 28, (9, 19, 31, 238), (35, 89, 112), 2)
    draw.text((x0 + 28, y0 + 25), "O(ALPHA) / PORTFOLIO CATALOG", font=F["mono"](20), fill=CYAN)
    draw.text((x1 - 28, y0 + 25), "LIVE PAPER", font=F["mono"](20), fill=CYAN, anchor="ra")
    draw.text((x0 + 28, y0 + 78), "Agent status  -  Optimising", font=F["mono"](24), fill=CYAN)
    draw.text((x0 + 28, y0 + 128), "$100,000", font=F["display_reg"](76), fill=WHITE)
    draw.text((x0 + 455, y0 + 162), "DEFAULT PAPER CASH", font=F["mono"](18), fill=(129, 156, 174))
    chart_top = y0 + 218
    for i in range(3):
        y = chart_top + i * 50
        draw.line((x0 + 28, y, x1 - 28, y), fill=(37, 70, 91), width=2)
    line_chart(draw, (x0 + 28, chart_top - 10, x1 - 28, y1 - 48), CYAN, fill=(18, 220, 238, 38), thick=6)


def compact_evidence(draw, box):
    x0, y0, x1, y1 = box
    rounded(draw, box, 28, (9, 19, 31, 238), (35, 89, 112), 2)
    draw.text((x0 + 24, y0 + 24), "LOW-RISK PASS", font=F["mono"](20), fill=CYAN)
    draw.text((x0 + 24, y0 + 68), "LGBM h63 + proxy", font=F["body_med"](28), fill=WHITE)
    rows = [("DSR", "1.000"), ("PBO", ".067 / .077"), ("OOS", "49-103 trades"), ("Mode", "Paper only")]
    y = y0 + 122
    for k, v in rows:
        draw.text((x0 + 24, y), k, font=F["mono"](17), fill=(139, 164, 180))
        draw.text((x0 + 24, y + 25), v, font=F["body_med"](25), fill=YELLOW if k == "PBO" else WHITE)
        y += 62


def draft_technical_blueprint():
    img = Image.new("RGBA", (W, H), (6, 13, 22, 255))
    d = ImageDraw.Draw(img, "RGBA")
    for x in range(0, W, 110):
        d.line((x, 0, x, H), fill=(20, 48, 65, 40), width=1)
    for y in range(0, H, 110):
        d.line((0, y, W, y), fill=(20, 48, 65, 40), width=1)
    glow_circle(img, (W - 360, 340), 450, CYAN, 45, 150)
    glow_circle(img, (390, 1220), 420, YELLOW, 28, 180)
    glow_circle(img, (1670, 2460), 520, CYAN, 27, 210)
    add_noise(img, 4, 7)

    rounded(d, (70, 70, W - 70, 470), 46, (13, 28, 45, 248), (39, 92, 113), 2)
    draw_brand(img, 120, 125, 230, False)
    d.text((395, 130), CONTENT["eyebrow"], font=F["mono"](25), fill=(142, 165, 178))
    d.text((395, 186), "O(Alpha)", font=F["display_reg"](92), fill=WHITE)
    d.text((395, 292), "Quant in your pocket.", font=F["display"](54), fill=(211, 234, 242))
    d.text((395, 358), "Validation-gated alpha research, paper execution, and full audit trails.", font=F["body_med"](29), fill=(158, 182, 196))
    d.text((W - 120, 138), "PAPER\nONLY", font=F["cond"](86), fill=YELLOW, anchor="ra", align="right", spacing=-8)
    d.text((W - 120, 322), "No live brokerage orders.\nMetrics cite committed reports.", font=F["mono"](22), fill=(150, 174, 187), anchor="ra", align="right", spacing=8)

    technical_card(
        d,
        (90, 540, 770, 905),
        1,
        "Motivation",
        [
            "Most trading demos stop at a pretty backtest.",
            "O(Alpha) keeps promotion strict: no report artifact, no claim.",
            "Goal: convert validated candidates into auditable paper runs.",
        ],
        CYAN,
    )
    technical_card(
        d,
        (815, 540, 1545, 905),
        2,
        "Research Harness",
        [
            "cmd/alpha-research writes JSON + MD under reports/batches.",
            "Walk-forward train/test folds with min OOS trades.",
            "DSR, PBO, cost stress, benchmark, turnover, and no-lookahead gates.",
        ],
        YELLOW,
    )
    technical_card(
        d,
        (1590, 540, 2390, 905),
        3,
        "Strategy Catalog",
        [
            "9 paper-only entries across low, medium, and high risk buckets.",
            "LGBM h63 rankers load local model artifacts and fail closed.",
            "Low-risk h63 LGBM + proxy pass both 2015 and shifted 2016 windows.",
        ],
        CYAN,
    )

    architecture_boxes(d, (90, 980, 2390, 1315))

    technical_card(
        d,
        (90, 1385, 785, 1775),
        4,
        "Execution Router",
        [
            "Portfolio targets become per-symbol notional deltas.",
            "Long-only v1 sells reductions before buy legs.",
            "Deterministic client_order_id prevents duplicate paper fills.",
            "Rebalances and risk exits create persisted alerts.",
        ],
        YELLOW,
    )
    technical_card(
        d,
        (845, 1385, 1540, 1775),
        5,
        "Runtime Agent",
        [
            "One active portfolio run per user.",
            "Warmup daily bars, evaluate latest panel, apply runtime settings.",
            "HMM regime state, last rebalance, and heartbeat saved to agent_runs.",
            "Backend restart can resume active portfolio runs.",
        ],
        CYAN,
    )
    mini_schema(d, (1600, 1385, 2390, 1775))

    rounded(d, (90, 1850, 1535, 2425), 38, (14, 29, 47, 242), (35, 89, 112), 2)
    d.text((130, 1890), "LIVE PAPER DASHBOARD", font=F["body_med"](34), fill=WHITE)
    d.text((130, 1942), "Backend state, not hand-entered poster numbers", font=F["mono"](22), fill=CYAN)
    compact_dashboard(d, (130, 2010, 1115, 2340))
    compact_evidence(d, (1150, 2010, 1500, 2340))

    technical_card(
        d,
        (1590, 1850, 2390, 2140),
        6,
        "Frontend Flow",
        [
            "Onboarding: choose risk -> select catalog strategy -> stream backtest.",
            "User accepts the backtest before completing onboarding.",
            "Dashboard reads summary, positions, history, fills, alerts, and active run status.",
        ],
        YELLOW,
    )
    technical_card(
        d,
        (1590, 2180, 2390, 2480),
        7,
        "SWE Practices",
        [
            "Migrations define accounts, orders, fills, positions, snapshots, bars, alerts.",
            "Tests cover catalog specs, router idempotency, resume behavior, parity, and strategies.",
            "API is authenticated; model artifacts are registered by URI, not stored as DB blobs.",
        ],
        CYAN,
    )

    rounded(d, (90, 2525, 2390, 2925), 38, (14, 29, 47, 242), (35, 89, 112), 2)
    d.text((130, 2567), "VERIFIED TECHNICAL CLAIMS", font=F["body_med"](34), fill=WHITE)
    claims = [
        ("Low-risk pass", "lgbm_ranker_h63_low and ranker_proxy_h63_low promote in both catalog-bucket windows."),
        ("Paper signal", "orders_enabled=false, broker_connected=false, orders_submitted=0 in paper_ranker_signal.md."),
        ("Current target model", "VOO core with h63 ranker active sleeve; model metadata includes feature count and artifact SHA."),
        ("Guardrail", "Medium/high entries are diagnostics or experimental because catalog-bucket PBO fails."),
    ]
    for i, (head, body) in enumerate(claims):
        x = 130 + (i % 2) * 1090
        y = 2630 + (i // 2) * 120
        draw_col = CYAN if i % 2 == 0 else YELLOW
        d.text((x, y), head.upper(), font=F["mono"](22), fill=draw_col)
        d.multiline_text((x, y + 34), wrap_to_width(d, body, F["body_med"](23), 960), font=F["body_med"](23), fill=(179, 200, 211), spacing=5)

    pipeline(d, 220, 3095, W - 440, "dark")
    tech_pills(d, 3288, "dark")
    draw_footer(d, img)
    return save(img, "v4_technical_blueprint.png")


def marketing_feature(draw, box, label, title, body, accent=CYAN):
    x0, y0, x1, y1 = box
    rounded(draw, box, 30, (14, 28, 45, 242), (42, 96, 118), 2)
    draw.text((x0 + 30, y0 + 26), label, font=F["mono"](21), fill=accent)
    draw.text((x0 + 30, y0 + 62), title, font=F["body_med"](33), fill=WHITE)
    wrapped = wrap_to_width(draw, body, F["body_med"](23), x1 - x0 - 60)
    draw.multiline_text((x0 + 30, y0 + 115), wrapped, font=F["body_med"](23), fill=(170, 193, 207), spacing=6)


def metric_tile(draw, box, label, value, caption, accent=CYAN):
    x0, y0, x1, y1 = box
    rounded(draw, box, 24, (8, 18, 30, 248), (44, 92, 111), 2)
    draw.text((x0 + 22, y0 + 20), label.upper(), font=F["mono"](19), fill=(132, 157, 172))
    draw.text((x0 + 22, y0 + 55), value, font=F["body_med"](37), fill=accent)
    draw.text((x0 + 22, y0 + 106), caption, font=F["body_med"](18), fill=(155, 177, 190))


def marketing_dashboard(draw, box):
    x0, y0, x1, y1 = box
    rounded(draw, box, 46, (12, 25, 41, 248), (41, 100, 123), 3)
    draw.text((x0 + 46, y0 + 44), "O(ALPHA) / STRATEGY COMMAND", font=F["mono"](25), fill=CYAN)
    draw.text((x1 - 46, y0 + 44), "VALIDATION GATE: PASSED", font=F["mono"](25), fill=YELLOW, anchor="ra")
    draw.text((x0 + 46, y0 + 112), "h63 Ranker Catalog", font=F["body_med"](58), fill=WHITE)
    draw.text((x0 + 46, y0 + 180), "Research artifact -> catalog strategy -> portfolio automation", font=F["body_med"](28), fill=(156, 181, 195))
    chart = (x0 + 46, y0 + 280, x1 - 46, y0 + 560)
    for i in range(5):
        y = chart[1] + i * 58
        draw.line((chart[0], y, chart[2], y), fill=(38, 72, 92), width=2)
    line_chart(draw, chart, CYAN, fill=(18, 220, 238, 36), thick=8)
    tiles = [
        ("DSR", "1.000", "deflated Sharpe gate", CYAN),
        ("PBO", ".067/.077", "primary + shifted windows", YELLOW),
        ("OOS", "49-103", "validated trades", WHITE),
    ]
    tx = x0 + 46
    for label, value, caption, accent in tiles:
        metric_tile(draw, (tx, y1 - 190, tx + 300, y1 - 38), label, value, caption, accent)
        tx += 330


def draft_chief_marketer():
    img = Image.new("RGBA", (W, H), (5, 10, 18, 255))
    d = ImageDraw.Draw(img, "RGBA")
    for x in range(0, W, 96):
        d.line((x, 0, x, H), fill=(22, 52, 70, 32), width=1)
    for y in range(0, H, 96):
        d.line((0, y, W, y), fill=(22, 52, 70, 32), width=1)
    for i in range(-H, W, 260):
        d.line((i, 0, i + H, H), fill=(20, 82, 92, 30), width=2)
    glow_circle(img, (W // 2, 680), 600, CYAN, 55, 180)
    glow_circle(img, (1900, 1180), 420, YELLOW, 28, 180)
    glow_circle(img, (410, 2260), 520, CYAN, 24, 220)
    add_noise(img, 4, 11)

    d.text((W // 2, 115), CONTENT["eyebrow"], font=F["mono"](27), fill=(146, 169, 181), anchor="ma")
    draw_brand(img, W // 2 - 150, 190, 300, True)
    d.text((W // 2, 540), "O(Alpha)", font=F["display_reg"](118), fill=WHITE, anchor="ma")
    d.text((W // 2, 655), "Quant in your pocket.", font=F["display"](82), fill=(215, 238, 246), anchor="ma")
    d.text((W // 2, 745), "A validation-gated alpha engine with portfolio automation and audit-grade state.", font=F["body_med"](35), fill=(160, 185, 199), anchor="ma")

    pill_x = W // 2 - 575
    for label, color in [("REPORT-BACKED", CYAN), ("DSR/PBO GATE", YELLOW), ("H63 RANKER", CYAN), ("POSTGRES LEDGER", YELLOW), ("LIVE DASHBOARD", CYAN)]:
        tw, _ = text_size(d, label, F["mono"](20))
        rounded(d, (pill_x, 830, pill_x + tw + 42, 874), 22, (18, 36, 57, 230))
        d.ellipse((pill_x + 16, 848, pill_x + 26, 858), fill=color)
        d.text((pill_x + 36, 843), label, font=F["mono"](20), fill=(184, 204, 216))
        pill_x += tw + 62

    marketing_dashboard(d, (305, 1010, 2175, 1645))

    cards = [
        ("01", "Validation Engine", "Walk-forward folds, cost stress, DSR/PBO, benchmark checks, turnover checks, and no-lookahead audits decide what graduates."),
        ("02", "Catalog Intelligence", "Nine catalog strategies span risk buckets; h63 LGBM and deterministic proxy entries carry the strongest low-risk evidence."),
        ("03", "Autonomous Loop", "The portfolio agent warms daily bars, evaluates the latest panel, applies runtime settings, and records every state transition."),
        ("04", "Audit-Grade State", "Orders, fills, positions, cash ledger, snapshots, alerts, runtime state, and heartbeats persist in Postgres."),
    ]
    sx, sy, cw, ch, gap = 165, 1735, 520, 270, 48
    for i, (label, title, body) in enumerate(cards):
        marketing_feature(d, (sx + i * (cw + gap), sy, sx + i * (cw + gap) + cw, sy + ch), label, title, body, CYAN if i % 2 == 0 else YELLOW)

    rounded(d, (165, 2085, 2315, 2545), 38, (14, 28, 45, 242), (42, 96, 118), 2)
    d.text((215, 2130), "TECHNICAL GROUNDING", font=F["body_med"](38), fill=WHITE)
    d.text((2315 - 50, 2137), "what makes it real", font=F["mono"](23), fill=CYAN, anchor="ra")
    left = [
        ("Research harness", "cmd/alpha-research writes committed JSON/MD artifacts under reports/batches."),
        ("Strategy provenance", "Catalog entries expose family, risk bucket, deployment status, evidence paths, and notes."),
        ("Execution router", "Target weights become deterministic, idempotent long-side fill events."),
    ]
    right = [
        ("Runtime intelligence", "HMM regime state, rebalance cadence, and active worker heartbeat are saved with each run."),
        ("Dashboard state", "Next.js reads summary, history, allocation, trades, alerts, active runs, and live marks."),
        ("Data layer", "Bars, universes, model artifact metadata, trials, orders, fills, and snapshots share one schema."),
    ]
    for col, items in enumerate([left, right]):
        x = 215 + col * 1075
        for j, (head, body) in enumerate(items):
            y = 2205 + j * 95
            d.ellipse((x, y + 8, x + 16, y + 24), fill=CYAN if (j + col) % 2 == 0 else YELLOW)
            d.text((x + 32, y), head, font=F["body_med"](26), fill=WHITE)
            d.text((x + 32, y + 36), wrap_to_width(d, body, F["body_med"](21), 890), font=F["body_med"](21), fill=(166, 189, 202))

    rounded(d, (165, 2620, 2315, 2875), 36, (10, 22, 35, 245), (42, 96, 118), 2)
    d.text((215, 2660), "SYSTEM FLOW", font=F["body_med"](34), fill=WHITE)
    flow = ["Research", "Reports", "Gate", "Catalog", "Agent", "Router", "Ledger", "Dashboard"]
    start, end, y = 300, 2210, 2775
    for i, label in enumerate(flow):
        cx = start + i * ((end - start) / (len(flow) - 1))
        color = CYAN if i % 2 == 0 else YELLOW
        d.ellipse((cx - 32, y - 32, cx + 32, y + 32), outline=color, width=6)
        d.ellipse((cx - 8, y - 8, cx + 8, y + 8), fill=color)
        d.text((cx, y + 54), label, font=F["body_med"](22), fill=WHITE, anchor="ma")
        if i < len(flow) - 1:
            nx = start + (i + 1) * ((end - start) / (len(flow) - 1))
            d.line((cx + 42, y, nx - 42, y), fill=(50, 91, 107), width=3)

    tech_pills(d, 3095, "dark")
    draw_brand(img, 115, H - 185, 82, False)
    d.text((225, H - 154), "Built by Zhao Shizhen & Tan Jia Jun", font=F["body_med"](29), fill=WHITE)
    d.text((225, H - 113), "Team 6645  /  O(Alpha)", font=F["mono"](20), fill=(142, 164, 177))
    d.text((W - 115, H - 154), "Evidence-backed metrics from committed validation artifacts.", font=F["body_med"](25), fill=(170, 193, 206), anchor="ra")
    d.text((W // 2, H - 66), "Evidence: agent_catalog_bucket_comparison.md  |  paper_ranker_signal.md  |  README.md", font=F["mono"](14), fill=(92, 112, 127), anchor="ma")
    return save(img, "v5_chief_marketer.png")


def agent_command_center(draw, box):
    x0, y0, x1, y1 = box
    rounded(draw, box, 48, (12, 25, 41, 248), (41, 100, 123), 3)
    draw.text((x0 + 48, y0 + 44), "O(ALPHA) / AGENT COMMAND CENTER", font=F["mono"](25), fill=CYAN)
    draw.text((x1 - 48, y0 + 44), "WORKER: HEALTHY", font=F["mono"](25), fill=YELLOW, anchor="ra")
    draw.text((x0 + 48, y0 + 124), "Autonomous Portfolio Agent", font=F["body_med"](60), fill=WHITE)
    draw.text((x0 + 48, y0 + 196), "Backtest a strategy, activate an agent, and let the portfolio loop run in the background.", font=F["body_med"](29), fill=(158, 182, 196))

    status = [
        ("Agent", "ACTIVE", CYAN),
        ("Strategy", "Verified Momentum", YELLOW),
        ("Market Scope", "100 symbols", CYAN),
        ("Mode", "Autopilot", WHITE),
        ("Rebalance", "Scheduled", YELLOW),
        ("Heartbeat", "Synced", CYAN),
    ]
    sx = x0 + 48
    sy = y0 + 270
    for i, (k, v, col) in enumerate(status):
        tx = sx + (i % 3) * 360
        ty = sy + (i // 3) * 102
        rounded(draw, (tx, ty, tx + 310, ty + 74), 18, (8, 18, 30, 245), (42, 86, 105), 1)
        draw.text((tx + 20, ty + 14), k.upper(), font=F["mono"](17), fill=(130, 156, 171))
        draw.text((tx + 20, ty + 40), v, font=F["body_med"](26), fill=col)

    chart = (x0 + 48, y0 + 500, x1 - 48, y1 - 94)
    dashboard_pnl_sparkline(draw, chart)
    draw.text((chart[0], y1 - 66), "24H AGO", font=F["mono"](15), fill=(111, 135, 150))
    draw.text((chart[2], y1 - 66), "NOW", font=F["mono"](15), fill=(111, 135, 150), anchor="ra")
    draw.text((x0 + 54, y1 - 38), "LIVE PORTFOLIO STATE", font=F["mono"](20), fill=CYAN)
    draw.text((x1 - 54, y1 - 38), "positions + fills + alerts + snapshots", font=F["mono"](20), fill=(151, 176, 190), anchor="ra")


def product_panel(draw, box, num, title, body, bullets, accent=CYAN):
    x0, y0, x1, y1 = box
    rounded(draw, box, 30, (14, 28, 45, 242), (42, 96, 118), 2)
    draw.text((x0 + 28, y0 + 25), num, font=F["mono"](22), fill=accent)
    draw.text((x0 + 72, y0 + 23), title, font=F["body_med"](32), fill=WHITE)
    body_text = wrap_to_width(draw, body, F["body_med"](22), x1 - x0 - 58)
    draw.multiline_text((x0 + 28, y0 + 82), body_text, font=F["body_med"](22), fill=(168, 191, 204), spacing=5)
    body_lines = body_text.count("\n") + 1
    y = y0 + 108 + body_lines * 31
    for i, bullet in enumerate(bullets):
        draw.ellipse((x0 + 30, y + 8, x0 + 43, y + 21), fill=accent if i % 2 == 0 else YELLOW)
        bullet_text = wrap_to_width(draw, bullet, F["body_med"](20), x1 - x0 - 94)
        draw.multiline_text((x0 + 58, y), bullet_text, font=F["body_med"](20), fill=(183, 204, 216), spacing=3)
        y += 38 + bullet_text.count("\n") * 24


def soft_shadow_panel(img, box, radius=34, fill=(255, 255, 255, 255), outline=(209, 222, 232), shadow=(28, 76, 112, 22), width=2, offset=(0, 8), blur=18):
    shadow_layer = Image.new("RGBA", img.size, (0, 0, 0, 0))
    sd = ImageDraw.Draw(shadow_layer, "RGBA")
    x0, y0, x1, y1 = box
    dx, dy = offset
    sd.rounded_rectangle((x0 + dx, y0 + dy, x1 + dx, y1 + dy), radius=radius, fill=shadow)
    img.alpha_composite(shadow_layer.filter(ImageFilter.GaussianBlur(blur)))
    d = ImageDraw.Draw(img, "RGBA")
    d.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def numbered_label(draw, x, y, num, title, accent=(17, 93, 191), dark=(22, 37, 58)):
    rounded(draw, (x, y, x + 66, y + 66), 12, accent)
    draw.text((x + 33, y + 33), num, font=F["body_med"](30), fill=WHITE, anchor="mm")
    draw.text((x + 88, y + 12), title, font=F["body_med"](40), fill=dark)


def clean_bullet_list(draw, x, y, items, max_w, accent=(17, 93, 191), text=(54, 69, 91)):
    for i, item in enumerate(items):
        yy = y + i * 52
        draw.ellipse((x, yy + 10, x + 16, yy + 26), fill=accent if i % 2 == 0 else (255, 125, 31))
        draw.multiline_text((x + 34, yy), wrap_to_width(draw, item, F["body_med"](23), max_w - 34), font=F["body_med"](23), fill=text, spacing=4)


def clean_feature_tile(img, box, title, body, icon, accent=(17, 93, 191)):
    d = ImageDraw.Draw(img, "RGBA")
    soft_shadow_panel(img, box, radius=22, fill=(255, 255, 255, 255), outline=(214, 226, 235), shadow=(28, 76, 112, 14), width=2, blur=12)
    x0, y0, x1, y1 = box
    rounded(d, (x0 + 24, y0 + 28, x0 + 94, y0 + 98), 18, (*accent[:3], 28), outline=accent, width=2)
    d.text((x0 + 59, y0 + 63), icon, font=F["mono"](26), fill=accent, anchor="mm")
    d.text((x0 + 118, y0 + 24), title, font=F["body_med"](30), fill=(21, 41, 66))
    d.multiline_text((x0 + 118, y0 + 70), wrap_to_width(d, body, F["body_med"](22), x1 - x0 - 148), font=F["body_med"](22), fill=(72, 87, 108), spacing=4)


def clean_agent_mockup(img, box):
    d = ImageDraw.Draw(img, "RGBA")
    x0, y0, x1, y1 = box
    soft_shadow_panel(img, box, radius=42, fill=(9, 20, 35, 255), outline=(34, 96, 123), shadow=(8, 43, 82, 48), width=2, blur=24)
    d.text((x0 + 46, y0 + 42), "O(ALPHA) AGENT", font=F["mono"](24), fill=CYAN)
    d.text((x1 - 46, y0 + 42), "WORKER HEALTHY", font=F["mono"](24), fill=YELLOW, anchor="ra")
    d.text((x0 + 46, y0 + 112), "Autonomous Portfolio Agent", font=F["body_med"](48), fill=WHITE)
    d.text((x0 + 46, y0 + 172), "Backtest verified. Launched. Running in the background.", font=F["body_med"](25), fill=(166, 190, 204))
    stat_w = 255
    stat_gap = 24
    stats = [
        ("AGENT", "ACTIVE", CYAN),
        ("STRATEGY", "Verified", YELLOW),
        ("SCOPE", "100 symbols", CYAN),
        ("MODE", "Autopilot", WHITE),
        ("REBALANCE", "Scheduled", YELLOW),
        ("HEARTBEAT", "Synced", CYAN),
    ]
    sx, sy = x0 + 46, y0 + 242
    for i, (k, v, col) in enumerate(stats):
        tx = sx + (i % 3) * (stat_w + stat_gap)
        ty = sy + (i // 3) * 88
        rounded(d, (tx, ty, tx + stat_w, ty + 64), 14, (16, 32, 51, 255), (38, 78, 96), 1)
        d.text((tx + 18, ty + 12), k, font=F["mono"](15), fill=(123, 150, 166))
        d.text((tx + 18, ty + 36), v, font=F["body_med"](22), fill=col)
    chart = (x0 + 46, y0 + 438, x1 - 46, y1 - 76)
    for i in range(4):
        y = chart[1] + i * 54
        d.line((chart[0], y, chart[2], y), fill=(34, 61, 78), width=1)
    line_chart(d, chart, CYAN, fill=(18, 220, 238, 30), thick=7)
    d.text((x0 + 46, y1 - 46), "live portfolio state", font=F["mono"](18), fill=CYAN)
    d.text((x1 - 46, y1 - 46), "positions / fills / alerts / snapshots", font=F["mono"](18), fill=(138, 164, 180), anchor="ra")


def clean_flow(draw, x0, y0, width):
    flow = [
        ("User", "risk + capital"),
        ("Backtest", "proof of fit"),
        ("Launch", "create agent"),
        ("Worker", "run loop"),
        ("Router", "orders"),
        ("Ledger", "state"),
        ("Dashboard", "control"),
    ]
    for i, (head, body) in enumerate(flow):
        cx = x0 + i * (width / (len(flow) - 1))
        color = (17, 93, 191) if i % 2 == 0 else (255, 125, 31)
        if i < len(flow) - 1:
            nx = x0 + (i + 1) * (width / (len(flow) - 1))
            draw.line((cx + 38, y0, nx - 38, y0), fill=(164, 190, 209), width=3)
        draw.ellipse((cx - 34, y0 - 34, cx + 34, y0 + 34), fill=(255, 255, 255), outline=color, width=6)
        draw.ellipse((cx - 9, y0 - 9, cx + 9, y0 + 9), fill=color)
        draw.text((cx, y0 + 58), head, font=F["body_med"](24), fill=(23, 42, 67), anchor="ma")
        draw.text((cx, y0 + 92), body, font=F["body_med"](17), fill=(83, 98, 118), anchor="ma")


def draft_clean_product_exhibition():
    img = Image.new("RGBA", (W, H), (247, 250, 252, 255))
    d = ImageDraw.Draw(img, "RGBA")
    navy = (19, 39, 66)
    blue = (17, 93, 191)
    orange = (255, 125, 31)
    pale_blue = (235, 244, 252)

    d.rectangle((0, 0, W, 430), fill=(255, 255, 255, 255))
    d.rounded_rectangle((W - 465, -110, W + 120, 315), radius=68, fill=(236, 246, 254, 255))
    d.rounded_rectangle((W - 352, -50, W + 70, 245), radius=58, fill=(248, 251, 255, 255), outline=(217, 230, 242), width=2)
    d.line((W - 390, 210, W - 85, 65), fill=(58, 132, 219), width=8)
    d.polygon([(W - 85, 65), (W - 190, 82), (W - 125, 132)], fill=(20, 102, 207))
    d.line((W - 360, 235, W - 290, 185, W - 225, 196, W - 160, 118), fill=orange, width=6)

    draw_brand(img, 122, 108, 170, False)
    d.text((322, 118), CONTENT["eyebrow"], font=F["mono"](23), fill=(107, 126, 143))
    d.text((322, 176), "O(Alpha)", font=F["display_reg"](86), fill=navy)
    d.text((322, 284), "Deploy your own trading agent.", font=F["body_med"](44), fill=blue)
    d.text((322, 344), "Pick a strategy. Prove it in a backtest. Launch an always-on portfolio worker.", font=F["body_med"](28), fill=(69, 84, 103))

    clean_agent_mockup(img, (1230, 520, 2320, 1225))

    soft_shadow_panel(img, (90, 520, 1135, 1225), radius=34, fill=(255, 255, 255, 255), outline=(211, 224, 234), shadow=(28, 76, 112, 16), width=2)
    numbered_label(d, 138, 570, "1", "The Product", blue, navy)
    product_copy = (
        "O(Alpha) turns the trading workflow into a personal autonomous agent: users choose their risk profile, "
        "select a proven strategy, launch a portfolio worker, and monitor execution from one command center."
    )
    d.multiline_text((138, 674), wrap_to_width(d, product_copy, F["body_med"](29), 900), font=F["body_med"](29), fill=(58, 73, 94), spacing=8)
    d.rounded_rectangle((138, 940, 1055, 1106), radius=22, fill=(255, 246, 236), outline=(255, 170, 97), width=2)
    d.text((176, 972), "ALWAYS-ON AGENT", font=F["mono"](21), fill=orange)
    d.multiline_text((176, 1012), "A trading agent that keeps working after the user closes the laptop.", font=F["body_med"](30), fill=(165, 77, 18), spacing=6)

    soft_shadow_panel(img, (90, 1315, 2390, 1810), radius=34, fill=(255, 255, 255, 255), outline=(211, 224, 234), shadow=(28, 76, 112, 12), width=2)
    numbered_label(d, 138, 1360, "2", "Core Capabilities", blue, navy)
    tiles = [
        ("Strategy Onboarding", "Risk-profile matching, catalog selection, and streamed backtest acceptance before launch.", "01", blue),
        ("Always-On Worker", "A background portfolio loop warms market data, evaluates bars, and tracks heartbeat health.", "02", orange),
        ("Execution Engine", "Target weights become deterministic buy/sell deltas with idempotent order keys.", "03", blue),
        ("Live Control Room", "Dashboard views portfolio value, allocation, trades, alerts, runtime state, and live marks.", "04", orange),
    ]
    tx, ty = 140, 1470
    tile_w, tile_h = 520, 150
    for i, (title, body, icon, accent) in enumerate(tiles):
        x = tx + (i % 2) * 1080
        y = ty + (i // 2) * 185
        clean_feature_tile(img, (x, y, x + 970, y + tile_h), title, body, icon, accent)

    soft_shadow_panel(img, (90, 1905, 1685, 2478), radius=34, fill=(255, 255, 255, 255), outline=(211, 224, 234), shadow=(28, 76, 112, 12), width=2)
    numbered_label(d, 138, 1950, "3", "How It Works", blue, navy)
    clean_flow(d, 230, 2200, 1340)
    d.text((138, 2358), "from user intent to autonomous execution", font=F["mono"](21), fill=blue)
    d.multiline_text((138, 2400), "Choose a strategy, verify it, launch the agent, and watch every portfolio action flow back into dashboard state.", font=F["body_med"](24), fill=(67, 82, 103), spacing=5)

    soft_shadow_panel(img, (1760, 1905, 2390, 2478), radius=34, fill=(255, 255, 255, 255), outline=(211, 224, 234), shadow=(28, 76, 112, 12), width=2)
    numbered_label(d, 1810, 1950, "4", "Target User", blue, navy)
    d.text((1810, 2072), "Independent traders", font=F["body_med"](28), fill=blue)
    d.multiline_text((1810, 2115), wrap_to_width(d, "Users who want disciplined strategy execution without sitting in front of charts all day.", F["body_med"](24), 500), font=F["body_med"](24), fill=(58, 73, 94), spacing=6)
    d.text((1810, 2265), "Key benefit", font=F["body_med"](28), fill=orange)
    d.multiline_text((1810, 2308), wrap_to_width(d, "A personal agent that turns backtested strategy intent into a continuously monitored portfolio loop.", F["body_med"](24), 500), font=F["body_med"](24), fill=(58, 73, 94), spacing=6)

    soft_shadow_panel(img, (90, 2585, 2390, 3048), radius=34, fill=(255, 255, 255, 255), outline=(211, 224, 234), shadow=(28, 76, 112, 12), width=2)
    numbered_label(d, 138, 2630, "5", "Technical Foundation", blue, navy)
    foundations = [
        ("Backend orchestration", "Go/Gin API, agent manager, portfolio worker, runtime settings, heartbeat, restart recovery."),
        ("Market + account state", "Bars, orders, fills, cash ledger, positions, snapshots, alerts, and strategy metadata in Postgres."),
        ("Frontend experience", "Next.js onboarding, streamed backtest, launch/stop controls, allocation, trade log, alerts, and live price updates."),
    ]
    for i, (head, body) in enumerate(foundations):
        x = 150 + i * 735
        d.rounded_rectangle((x, 2748, x + 640, 2928), radius=22, fill=pale_blue if i % 2 == 0 else (255, 246, 236), outline=(209, 224, 235), width=2)
        d.text((x + 28, 2782), head, font=F["body_med"](25), fill=blue if i % 2 == 0 else orange)
        d.multiline_text((x + 28, 2824), wrap_to_width(d, body, F["body_med"](19), 580), font=F["body_med"](19), fill=(66, 81, 101), spacing=4)

    stack = ["Go", "Gin", "Next.js", "React", "TypeScript", "PostgreSQL", "TimescaleDB", "Redis", "Docker", "Vercel", "Alpaca", "Yahoo"]
    widths = [text_size(d, s, F["mono"](19))[0] + 42 for s in stack]
    x = (W - sum(widths) - 12 * (len(stack) - 1)) // 2
    for s, wid in zip(stack, widths):
        d.rounded_rectangle((x, 3165, x + wid, 3208), radius=20, fill=(232, 240, 248), outline=(213, 225, 236), width=1)
        d.text((x + 21, 3177), s, font=F["mono"](19), fill=(64, 82, 105))
        x += wid + 12

    d.line((95, H - 270, W - 95, H - 270), fill=(209, 222, 232), width=2)
    draw_brand(img, 118, H - 208, 78, False)
    d.text((220, H - 176), "Built by Zhao Shizhen & Tan Jia Jun", font=F["body_med"](27), fill=navy)
    d.text((220, H - 136), "Team 6645  /  O(Alpha)", font=F["mono"](20), fill=(91, 111, 130))
    d.text((W - 118, H - 174), "An autonomous portfolio-agent platform for hands-free strategy execution.", font=F["body_med"](25), fill=(47, 63, 84), anchor="ra")
    d.text((W // 2, H - 74), "Code-grounded: portfolio_agent_handler.go  |  user_manager.go  |  db_execution_router.go  |  dashboard/page.tsx", font=F["mono"](14), fill=(112, 130, 146), anchor="ma")
    return save(img, "v7_clean_product_exhibition.png")


def draft_autonomous_agent_product():
    img = Image.new("RGBA", (W, H), (5, 10, 18, 255))
    d = ImageDraw.Draw(img, "RGBA")
    for y in range(H):
        shade = int(10 + (y / H) * 8)
        d.line((0, y, W, y), fill=(5, shade, 20 + int((y / H) * 8), 255))
    glow_circle(img, (W // 2, 520), 540, CYAN, 52, 170)
    glow_circle(img, (1860, 980), 440, YELLOW, 30, 190)
    glow_circle(img, (430, 2360), 500, CYAN, 24, 220)

    d.text((W // 2, 98), CONTENT["eyebrow"], font=F["mono"](25), fill=(145, 168, 181), anchor="ma")
    draw_brand(img, W // 2 - 128, 160, 256, True)
    d.text((W // 2, 460), "O(Alpha)", font=F["display_reg"](106), fill=WHITE, anchor="ma")
    d.text((W // 2, 570), "Deploy your own trading agent.", font=F["display"](78), fill=(218, 241, 248), anchor="ma")
    d.text((W // 2, 660), "Pick a strategy. Prove it in a backtest. Launch an always-on portfolio worker.", font=F["body_med"](34), fill=(163, 188, 202), anchor="ma")

    badge_items = [("PERSONAL AGENT", CYAN), ("BACKTEST VERIFIED", YELLOW), ("AUTO REBALANCE", CYAN), ("RISK CONTROLS", YELLOW), ("LIVE MONITORING", CYAN), ("AUDIT TRAIL", YELLOW)]
    badge_gap = 24
    badge_widths = [text_size(d, label, F["mono"](20))[0] + 56 for label, _ in badge_items]
    pill_x = (W - sum(badge_widths) - badge_gap * (len(badge_items) - 1)) // 2
    for (label, color), pill_w in zip(badge_items, badge_widths):
        tw, _ = text_size(d, label, F["mono"](20))
        rounded(d, (pill_x, 744, pill_x + pill_w, 788), 22, (18, 36, 57, 230))
        d.ellipse((pill_x + 16, 762, pill_x + 26, 772), fill=color)
        d.text((pill_x + 36, 757), label, font=F["mono"](20), fill=(184, 204, 216))
        pill_x += pill_w + badge_gap

    agent_command_center(d, (245, 905, 2235, 1605))

    panel_y0, panel_y1 = 1695, 2140
    panel_w, panel_gap = 515, 60
    panel_x = [120 + i * (panel_w + panel_gap) for i in range(4)]

    product_panel(
        d,
        (panel_x[0], panel_y0, panel_x[0] + panel_w, panel_y1),
        "01",
        "Strategy Onboarding",
        "The user starts with a risk profile and a strategy catalog, then runs a streamed backtest before activation.",
        ["risk-profile matching", "catalog strategy selection", "backtest acceptance"],
        CYAN,
    )
    product_panel(
        d,
        (panel_x[1], panel_y0, panel_x[1] + panel_w, panel_y1),
        "02",
        "Always-On Worker",
        "A portfolio worker warms market data, evaluates the latest daily panel, and continues running in the background.",
        ["one active run per user", "heartbeat tracking", "restart resume"],
        YELLOW,
    )
    product_panel(
        d,
        (panel_x[2], panel_y0, panel_x[2] + panel_w, panel_y1),
        "03",
        "Execution Engine",
        "Target weights are reconciled into ordered buy/sell deltas with deterministic IDs and persisted account state.",
        ["sell reductions first", "idempotent order keys", "fills + positions"],
        CYAN,
    )
    product_panel(
        d,
        (panel_x[3], panel_y0, panel_x[3] + panel_w, panel_y1),
        "04",
        "Live Control Room",
        "The dashboard follows the agent through portfolio value, allocation, trades, alerts, runtime state, and live marks.",
        ["agent status", "execution log", "alerts"],
        YELLOW,
    )

    rounded(d, (120, 2240, 2360, 2635), 38, (14, 28, 45, 242), (42, 96, 118), 2)
    d.text((170, 2285), "HOW O(ALPHA) WORKS", font=F["body_med"](38), fill=WHITE)
    d.text((2310, 2292), "from user intent to autonomous execution", font=F["mono"](22), fill=CYAN, anchor="ra")
    flow = [
        ("User", "chooses risk\nand capital"),
        ("Backtest", "streams strategy\nperformance"),
        ("Launch", "creates agent\nrun"),
        ("Worker", "evaluates bars\nin background"),
        ("Router", "reconciles\nweights"),
        ("Ledger", "stores fills\nand snapshots"),
        ("Dashboard", "monitors live\nstate"),
    ]
    start, end, y = 250, 2230, 2470
    for i, (head, body) in enumerate(flow):
        cx = start + i * ((end - start) / (len(flow) - 1))
        color = CYAN if i % 2 == 0 else YELLOW
        d.ellipse((cx - 36, y - 36, cx + 36, y + 36), outline=color, width=6)
        d.ellipse((cx - 10, y - 10, cx + 10, y + 10), fill=color)
        d.text((cx, y + 54), head, font=F["body_med"](24), fill=WHITE, anchor="ma")
        d.multiline_text((cx, y + 88), body, font=F["body_med"](18), fill=(158, 181, 195), anchor="ma", align="center", spacing=3)
        if i < len(flow) - 1:
            nx = start + (i + 1) * ((end - start) / (len(flow) - 1))
            d.line((cx + 48, y, nx - 48, y), fill=(54, 96, 112), width=3)

    rounded(d, (120, 2715, 2360, 3015), 38, (10, 22, 35, 246), (42, 96, 118), 2)
    d.text((170, 2757), "TECHNICAL FOUNDATION", font=F["body_med"](36), fill=WHITE)
    foundations = [
        ("Backend orchestration", "Go/Gin API, agent manager, portfolio worker, runtime settings, heartbeat, restart recovery."),
        ("Market + account state", "Bars, orders, fills, cash ledger, positions, snapshots, alerts, and strategy metadata in Postgres."),
        ("Frontend experience", "Next.js onboarding, streamed backtest, launch/stop controls, allocation, trade log, alerts, and live price updates."),
    ]
    for i, (head, body) in enumerate(foundations):
        x = 170 + i * 705
        d.text((x, 2825), head.upper(), font=F["mono"](21), fill=CYAN if i % 2 == 0 else YELLOW)
        d.multiline_text((x, 2860), wrap_to_width(d, body, F["body_med"](22), 630), font=F["body_med"](22), fill=(178, 199, 211), spacing=5)

    tech_pills(d, 3175, "dark")
    draw_brand(img, 115, H - 185, 82, False)
    d.text((225, H - 154), "Built by Zhao Shizhen & Tan Jia Jun", font=F["body_med"](29), fill=WHITE)
    d.text((225, H - 113), "Team 6645  /  O(Alpha)", font=F["mono"](20), fill=(142, 164, 177))
    d.text((W - 115, H - 154), "An autonomous portfolio-agent platform for hands-free strategy execution.", font=F["body_med"](25), fill=(184, 206, 218), anchor="ra")
    d.text((W // 2, H - 66), "Code-grounded: portfolio_agent_handler.go  |  user_manager.go  |  db_execution_router.go  |  dashboard/page.tsx", font=F["mono"](14), fill=(92, 112, 127), anchor="ma")
    return save(img, "v6_autonomous_agent_product.png")


def draft_dark_fintech():
    img = Image.new("RGBA", (W, H), (6, 14, 24, 255))
    d = ImageDraw.Draw(img, "RGBA")
    for x in range(0, W, 92):
        d.line((x, 0, x, H), fill=(22, 50, 70, 38), width=1)
    for y in range(0, H, 92):
        d.line((0, y, W, y), fill=(22, 50, 70, 38), width=1)
    glow_circle(img, (W // 2, 650), 520, CYAN, 50, 140)
    glow_circle(img, (W // 2 + 360, 900), 360, YELLOW, 26, 180)
    add_noise(img, 4, 2)
    d.text((W // 2, 170), CONTENT["eyebrow"], font=F["mono"](26), fill=(154, 177, 190), anchor="ma")
    draw_brand(img, W // 2 - 170, 245, 340)
    d.text((W // 2, 642), CONTENT["title"], font=F["display_reg"](112), fill=WHITE, anchor="ma")
    d.text((W // 2, 760), CONTENT["tagline"], font=F["display"](70), fill=(209, 232, 240), anchor="ma")
    d.text((W // 2, 852), CONTENT["sub"], font=F["body_med"](32), fill=MUTED, anchor="ma")
    cw, gap = 500, 38
    sx = (W - (cw * 4 + gap * 3)) // 2
    for i, (num, title, body) in enumerate(CONTENT["cards"]):
        card(d, (sx + i * (cw + gap), 1038, sx + i * (cw + gap) + cw, 1328), title, body, num, accent=CYAN if i % 2 == 0 else YELLOW, body_size=23)
    dashboard_panel(d, (480, 1515, 1760, 2110), "dark")
    evidence_card(d, (1650, 1410, 2165, 1990), "dark")
    rounded(d, (330, 1865, 655, 2110), 30, (16, 31, 50, 232), (37, 100, 124), 2)
    d.text((382, 1910), "RUN SAFETY", font=F["mono"](22), fill=CYAN)
    for j, (k, v, c) in enumerate([("PAPER", "true", CYAN), ("ORDERS", "0", YELLOW), ("BROKER", "off", PINK)]):
        yy = 1965 + j * 42
        d.ellipse((386, yy + 8, 400, yy + 22), fill=c)
        d.text((422, yy), k, font=F["mono"](22), fill=(156, 178, 190))
        d.text((600, yy), v, font=F["mono"](22), fill=c, anchor="ra")
    pipeline(d, 180, 2660, W - 360, "dark")
    tech_pills(d, 3145, "dark")
    draw_footer(d, img)
    for segment in [(50, 50, 110, 50), (50, 50, 50, 110), (W - 50, 50, W - 110, 50), (W - 50, 50, W - 50, 110), (50, H - 50, 110, H - 50), (50, H - 50, 50, H - 110), (W - 50, H - 50, W - 110, H - 50), (W - 50, H - 50, W - 50, H - 110)]:
        d.line(segment, fill=(97, 126, 144), width=3)
    return save(img, "v1_dark_fintech.png")


def draft_light_exhibition():
    img = Image.new("RGBA", (W, H), (246, 250, 248, 255))
    d = ImageDraw.Draw(img, "RGBA")
    for y in range(0, H, 124):
        d.line((0, y, W, y), fill=(222, 232, 235, 90), width=1)
    for x in range(0, W, 124):
        d.line((x, 0, x, H), fill=(222, 232, 235, 70), width=1)
    glow_circle(img, (W - 430, 460), 480, YELLOW, 78, 160)
    glow_circle(img, (320, 1140), 440, CYAN, 64, 170)
    rounded(d, (80, 80, W - 80, 430), 46, (8, 18, 29, 255))
    draw_brand(img, 135, 120, 230, False)
    d.text((420, 155), CONTENT["eyebrow"], font=F["mono"](28), fill=(132, 160, 176))
    d.text((420, 205), CONTENT["title"], font=F["display_reg"](90), fill=WHITE)
    d.text((420, 320), "Validation-gated alpha research -> paper execution.", font=F["body_med"](38), fill=(188, 216, 226))
    d.text((W - 145, 184), "PAPER\nONLY", font=F["cond"](92), fill=YELLOW, anchor="ra", align="right", spacing=-8)
    d.multiline_text((120, 555), "Quant in\nyour pocket.", font=F["display"](170), fill=INK, spacing=-8)
    d.multiline_text((128, 960), "Autonomous portfolio research, validation gates,\nand auditable paper execution.", font=F["body_med"](34), fill=(75, 89, 103), spacing=8)
    rounded(d, (1280, 585, 2265, 1390), 56, (255, 255, 255, 245), (196, 211, 218), 3)
    dashboard_panel(d, (1340, 685, 2200, 1285), "light")
    mx, my = 120, 1180
    for k, v in CONTENT["metrics"]:
        rounded(d, (mx, my, mx + 245, my + 132), 28, (9, 17, 29, 255))
        d.text((mx + 24, my + 24), k, font=F["mono"](24), fill=(130, 154, 168))
        d.text((mx + 24, my + 65), v, font=F["body_med"](32), fill=YELLOW if k == "PBO" else WHITE)
        mx += 265
    sx, sy, cw, ch = 120, 1560, 520, 286
    for i, (num, title, body) in enumerate(CONTENT["cards"]):
        x = sx + (i % 2) * (cw + 40)
        y = sy + (i // 2) * (ch + 42)
        card(d, (x, y, x + cw, y + ch), title, body, num, accent=CYAN if i % 2 == 0 else YELLOW, theme="light", body_size=22)
    rounded(d, (120, 2390, W - 120, 2820), 44, (12, 25, 41, 255))
    d.text((180, 2445), "FROM RESEARCH HARNESS TO AUDITABLE STATE", font=F["mono"](28), fill=CYAN)
    pipeline(d, 250, 2630, W - 500, "dark")
    tech_pills(d, 3005, "light")
    d.text((W - 120, H - 270), "Low-risk catalog entries pass both primary and shifted windows.\nMedium/high variants remain diagnostics, not promoted choices.", font=F["body_med"](28), fill=(64, 78, 92), anchor="ra", align="right", spacing=8)
    draw_footer(d, img, "light")
    return save(img, "v2_light_exhibition.png")


def draft_terminal_dashboard():
    img = Image.new("RGBA", (W, H), (4, 8, 14, 255))
    d = ImageDraw.Draw(img, "RGBA")
    for i in range(-H, W, 150):
        d.line((i, 0, i + H, H), fill=(16, 50, 62, 62), width=2)
    glow_circle(img, (W // 2, 1200), 730, CYAN, 42, 220)
    glow_circle(img, (1900, 560), 420, PINK, 35, 180)
    glow_circle(img, (360, 2900), 420, YELLOW, 28, 180)
    add_noise(img, 5, 5)
    rounded(d, (96, 96, 290, H - 96), 48, (14, 29, 45, 238), (35, 85, 105), 2)
    d.text((193, 230), "O(Alpha)", font=F["cond"](56), fill=WHITE, anchor="ma")
    draw_brand(img, 126, 310, 134, False)
    for idx, label in enumerate(["VALIDATE", "CATALOG", "RUN", "AUDIT"]):
        y = 620 + idx * 220
        d.text((193, y), label, font=F["mono"](24), fill=CYAN if idx % 2 == 0 else YELLOW, anchor="ma")
        d.line((140, y + 42, 246, y + 42), fill=(50, 101, 118), width=2)
    d.text((193, H - 260), "APOLLO\nTEAM 6645", font=F["mono"](24), fill=(132, 154, 168), anchor="ma", align="center")
    x0 = 380
    d.text((x0, 160), CONTENT["eyebrow"], font=F["mono"](26), fill=(141, 162, 177))
    d.text((x0, 260), "Quant in your pocket.", font=F["display_reg"](132), fill=WHITE)
    d.text((x0, 405), "Validated alpha research, paper execution, and full audit trails in one system.", font=F["body_med"](36), fill=(168, 194, 207))
    rounded(d, (430, 600, 2260, 1680), 56, (12, 25, 41, 246), (32, 91, 115), 3)
    dashboard_panel(d, (520, 720, 1580, 1470), "dark")
    evidence_card(d, (1635, 720, 2180, 1468), "dark")
    rounded(d, (430, 1760, 2260, 2180), 40, (12, 25, 41, 238), (32, 91, 115), 2)
    d.text((500, 1820), "SYSTEM CAPABILITIES", font=F["mono"](30), fill=CYAN)
    cap_lines = [
        "Authenticated Go/Gin API and Next.js dashboard",
        "Portfolio catalog agent with one active paper run per user",
        "DB-backed fills, positions, cash ledger, snapshots, alerts",
        "Runtime HMM regime state, heartbeat, and restart resume",
        "Alpaca/Yahoo market data paths with Postgres bar storage",
    ]
    for i, line in enumerate(cap_lines):
        yy = 1890 + i * 48
        d.ellipse((502, yy + 11, 520, yy + 29), fill=YELLOW if i % 2 else CYAN)
        d.text((540, yy), line, font=F["body_med"](31), fill=WHITE)
    sx, y, cw, gap = 430, 2290, 430, 34
    for i, (num, title, body) in enumerate(CONTENT["cards"]):
        card(d, (sx + i * (cw + gap), y, sx + i * (cw + gap) + cw, y + 310), title, body, num, accent=CYAN if i % 2 == 0 else YELLOW, body_size=21)
    pipeline(d, 430, 2910, 1830, "dark")
    tech_pills(d, 3210, "dark")
    d.text((430, H - 180), "Built by Zhao Shizhen & Tan Jia Jun", font=F["body_med"](34), fill=WHITE)
    d.text((430, H - 132), "Paper-only research system. No live brokerage orders.", font=F["mono"](23), fill=YELLOW)
    d.text((W // 2, H - 64), CONTENT["footer"], font=F["mono"](14), fill=(92, 112, 127), anchor="ma")
    return save(img, "v3_terminal_dashboard.png")


def write_rundown():
    rundown = """# O(Alpha) Current Codebase Rundown

- Backend: Go/Gin API with auth, CORS, migrations, bars repository, backtests, portfolio catalog routes, single-symbol legacy agents, and portfolio catalog agents.
- Research harness: alpha-research, backtest, ml-meta-research, HMM exit research, ranker parity tools, paper ranker signal, Alpaca/Yahoo ingest.
- Strategy catalog: nine paper-only catalog strategies across low/medium/high risk, including LGBM h63 rankers, deterministic h63 proxy, low-vol sleeve, ranked sleeves, TSMOM, and composite momentum.
- Promotion evidence: official artifacts live under reports/batches; low-risk LGBM h63 and ranker proxy pass both primary 2015 and shifted 2016 catalog-bucket windows.
- Paper execution: PortfolioOrchestrator starts one active portfolio run per user, warms daily bars, evaluates the catalog strategy, applies runtime settings, and routes targets through DB or Alpaca-paper routers.
- DB state: orders, fills, positions, cash ledger, portfolio snapshots, system alerts, agent runs, settings, bars, strategy trials, model artifact metadata, universes, pair candidates.
- Execution safety: long-only DB router sells reductions before buys, uses deterministic client_order_id keys, emits rebalance/risk-exit alerts, updates snapshots, and keeps live broker execution distinct.
- Runtime resilience: active portfolio runs can resume after restart, runtime HMM state is saved into agent_runs parameters, and heartbeats update each evaluation loop.
- Frontend: Next.js/React dashboard with onboarding, risk profile selection, catalog strategy selection, streamed backtest acceptance, launch/stop, status polling, summary/history/positions/trades/alerts, and live portfolio stream updates.
- Do not market: live real-money trading, nonzero annual yield, or medium/high PBO-failing variants as promoted alpha.
"""
    path = ROOT / "output/poster/codebase_rundown.md"
    path.write_text(rundown)
    return str(path)


if __name__ == "__main__":
    manifest = {
        "drafts": [
            draft_dark_fintech(),
            draft_light_exhibition(),
            draft_terminal_dashboard(),
            draft_technical_blueprint(),
            draft_chief_marketer(),
            draft_autonomous_agent_product(),
            draft_clean_product_exhibition(),
        ],
        "rundown": write_rundown(),
    }
    (OUT / "manifest.json").write_text(json.dumps(manifest, indent=2))
    print(json.dumps(manifest, indent=2))
