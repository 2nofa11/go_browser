#!/usr/bin/env python3
"""chapters/*/ja.md を GitHub Pages 用の静的HTMLに変換する。

章を追加するときは chapters/<番号>-<スラッグ>/ja.md を置くだけでよい。
原文（source.md）は再配布を避けるため公開しない。

使い方: python3 scripts/build_site.py [出力先=_site]
環境変数 GITHUB_REPOSITORY (owner/repo) があれば、.go などへのリンクを GitHub の blob URL に張り替える。
"""
import html
import os
import re
import shutil
import sys
from pathlib import Path

import markdown
from pygments.formatters import HtmlFormatter

ROOT = Path(__file__).resolve().parent.parent
OUT = ROOT / (sys.argv[1] if len(sys.argv) > 1 else "_site")
REPO = os.environ.get("GITHUB_REPOSITORY")
BRANCH = os.environ.get("GITHUB_REF_NAME", "main")

EXTENSIONS = ["fenced_code", "tables", "footnotes", "toc", "sane_lists", "codehilite"]
EXT_CONFIG = {"codehilite": {"guess_lang": False, "css_class": "codehilite"}}

CSS = """
:root{--bg:#fff;--fg:#1f2328;--muted:#656d76;--border:#d0d7de;--code-bg:#f6f8fa;--link:#0969da}
@media(prefers-color-scheme:dark){:root{--bg:#0d1117;--fg:#e6edf3;--muted:#8d96a0;--border:#30363d;--code-bg:#161b22;--link:#58a6ff}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font:16px/1.8 -apple-system,BlinkMacSystemFont,"Hiragino Sans","Noto Sans JP",sans-serif}
main,header,footer{max-width:46rem;margin:0 auto;padding:0 16px}
header{padding-top:12px;padding-bottom:12px;border-bottom:1px solid var(--border);font-size:.9rem}
header a{margin-right:1rem}
a{color:var(--link)}
h1,h2,h3{line-height:1.4}
h2{border-bottom:1px solid var(--border);padding-bottom:.3em;margin-top:2em}
img{max-width:100%;height:auto}
blockquote{margin:1em 0;padding:0 1em;color:var(--muted);border-left:.25em solid var(--border)}
pre,code{font:.875rem/1.5 ui-monospace,SFMono-Regular,Menlo,monospace}
code{background:var(--code-bg);padding:.15em .35em;border-radius:4px}
pre{background:var(--code-bg);padding:12px 14px;border-radius:6px;overflow-x:auto}
pre code{background:none;padding:0}
.codehilite{background:var(--code-bg);border-radius:6px;margin:1em 0;overflow-x:auto}
.codehilite pre{margin:0}
table{border-collapse:collapse;display:block;overflow-x:auto}
th,td{border:1px solid var(--border);padding:6px 12px}
.footnote{font-size:.9rem;color:var(--muted)}
nav.pager{display:flex;justify-content:space-between;gap:1rem;margin:3rem 0 2rem;font-size:.9rem}
ul.chapters{list-style:none;padding:0}
ul.chapters li{padding:.6rem 0;border-bottom:1px solid var(--border)}
ul.chapters small{color:var(--muted);margin-left:.5rem}
footer{padding-bottom:2rem;color:var(--muted);font-size:.85rem}
"""


def pygments_css() -> str:
    light = HtmlFormatter(style="default").get_style_defs(".codehilite")
    dark = HtmlFormatter(style="monokai").get_style_defs(".codehilite")
    return f"{light}\n@media(prefers-color-scheme:dark){{\n{dark}\n}}"


def render_md(text: str) -> tuple[str, str]:
    md = markdown.Markdown(extensions=EXTENSIONS, extension_configs=EXT_CONFIG)
    body = md.convert(text)
    m = re.search(r"^#\s+(.+)$", text, re.M)
    return body, (m.group(1).strip() if m else "")


def fix_links(body: str, src_dir: Path) -> str:
    """相対リンクを張り替える。.md は .html へ、それ以外は GitHub の blob URL へ。"""

    def repl(m: re.Match) -> str:
        href = m.group(1)
        if re.match(r"^(https?:|mailto:|#|img/)", href):
            return m.group(0)
        path, _, frag = href.partition("#")
        suffix = f"#{frag}" if frag else ""
        if path.endswith(".md"):
            return f'href="{path[:-3]}.html{suffix}"'
        if REPO:
            rel = (src_dir / path).resolve().relative_to(ROOT).as_posix()
            return f'href="https://github.com/{REPO}/blob/{BRANCH}/{rel}{suffix}"'
        return m.group(0)

    return re.sub(r'href="([^"]+)"', repl, body)


def page(title: str, body: str, root: str, pager: str = "") -> str:
    return f"""<!doctype html>
<html lang="ja">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{html.escape(title)} | go_browser</title>
<link rel="stylesheet" href="{root}style.css">
</head>
<body>
<header><a href="{root}index.html">go_browser</a><a href="{root}glossary.html">用語集</a></header>
<main>
{body}
{pager}
</main>
<footer>Web Browser Engineering の日本語訳とGo実装（学習用）</footer>
</body>
</html>
"""


def main() -> None:
    chapters = sorted(p for p in (ROOT / "chapters").iterdir() if (p / "ja.md").exists())

    if OUT.exists():
        shutil.rmtree(OUT)
    OUT.mkdir(parents=True)
    (OUT / "style.css").write_text(CSS + pygments_css(), encoding="utf-8")
    (OUT / ".nojekyll").touch()

    entries = []
    for i, d in enumerate(chapters):
        dest = OUT / "chapters" / d.name
        dest.mkdir(parents=True)
        if (d / "img").is_dir():
            shutil.copytree(d / "img", dest / "img")

        body, title = render_md((d / "ja.md").read_text(encoding="utf-8"))
        pager_parts = []
        if i > 0:
            pager_parts.append(f'<a href="../{chapters[i-1].name}/ja.html">← 前の章</a>')
        else:
            pager_parts.append("<span></span>")
        if i + 1 < len(chapters):
            pager_parts.append(f'<a href="../{chapters[i+1].name}/ja.html">次の章 →</a>')
        pager = f'<nav class="pager">{"".join(pager_parts)}</nav>'
        (dest / "ja.html").write_text(
            page(title, fix_links(body, d), "../../", pager), encoding="utf-8"
        )

        has_ex = (d / "exercises.md").exists()
        if has_ex:
            ex_body, ex_title = render_md((d / "exercises.md").read_text(encoding="utf-8"))
            (dest / "exercises.html").write_text(
                page(ex_title, fix_links(ex_body, d), "../../"), encoding="utf-8"
            )
        entries.append((d.name, title, has_ex))

    items = "\n".join(
        f'<li><a href="chapters/{n}/ja.html">{html.escape(t)}</a>'
        + (f'<small><a href="chapters/{n}/exercises.html">演習</a></small>' if ex else "")
        + "</li>"
        for n, t, ex in entries
    )
    intro = (
        '<h1>go_browser</h1>'
        '<p><a href="https://browser.engineering/">Web Browser Engineering</a> '
        "の日本語訳（学習用）とGo実装。</p>"
    )
    (OUT / "index.html").write_text(
        page("目次", f'{intro}<ul class="chapters">\n{items}\n</ul>', ""), encoding="utf-8"
    )

    gloss, gtitle = render_md((ROOT / "GLOSSARY.md").read_text(encoding="utf-8"))
    (OUT / "glossary.html").write_text(page(gtitle, gloss, ""), encoding="utf-8")

    print(f"built {len(entries)} chapters -> {OUT}")


if __name__ == "__main__":
    main()
