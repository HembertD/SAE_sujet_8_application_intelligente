from pathlib import Path
import shutil
import subprocess

import yaml


ROOT = Path(__file__).resolve().parents[1]


def collect_pages(entry, pages, seen):
    if isinstance(entry, dict):
        for title, value in entry.items():
            if isinstance(value, str):
                relative_path = value.strip()
                if relative_path not in seen:
                    pages.append((title, relative_path))
                    seen.add(relative_path)
            elif isinstance(value, (list, dict)):
                collect_pages(value, pages, seen)
    elif isinstance(entry, list):
        for item in entry:
            collect_pages(item, pages, seen)


def build_markdown():
    mkdocs_path = ROOT / "mkdocs.yml"
    config = yaml.safe_load(mkdocs_path.read_text(encoding="utf-8"))
    pages = []
    collect_pages(config.get("nav", []), pages, set())

    content = [
        "# Documentation technique - SAE Application intelligente",
        "",
        "---",
        "",
        "## Table des matières",
        "",
        r"\tableofcontents",
        "",
        "---",
        "",
    ]

    for title, relative_path in pages:
        file_path = ROOT / relative_path
        if not file_path.exists():
            file_path = ROOT / "docs" / relative_path
        if not file_path.exists():
            raise FileNotFoundError(
                f"Missing Markdown file referenced in mkdocs.yml: {relative_path}"
            )

        text = file_path.read_text(encoding="utf-8")
        clean_title = title.strip() if title else file_path.stem.replace("_", " ").title()
        content.extend((f"# {clean_title}", "", text.strip(), "", "---", ""))

    output = ROOT / ".tmp-doc" / "documentation.md"
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text("\n".join(content).strip() + "\n", encoding="utf-8")
    print(f"Generated {output} with {len(pages)} documentation pages.")
    return output


def main():
    required_tools = ("pandoc", "xelatex")
    missing_tools = [tool for tool in required_tools if shutil.which(tool) is None]
    if missing_tools:
        missing = ", ".join(missing_tools)
        raise SystemExit(
            f"Dependance manquante : {missing}. Installez-la avec :\n"
            "sudo apt-get update && sudo apt-get install -y pandoc "
            "texlive-xetex texlive-latex-recommended "
            "texlive-fonts-recommended lmodern"
        )

    markdown_path = build_markdown()
    subprocess.run(
        [
            "pandoc",
            str(markdown_path.relative_to(ROOT)),
            "--from=markdown+raw_tex",
            "--standalone",
            "--toc",
            "--number-sections",
            "--pdf-engine=xelatex",
            "-V",
            "geometry:margin=1in",
            "-V",
            "colorlinks=true",
            "-V",
            "linkcolor=blue",
            "-V",
            "urlcolor=blue",
            "-V",
            "mainfont=DejaVu Sans",
            "-V",
            "monofont=DejaVu Sans Mono",
            "-o",
            "documentation.pdf",
        ],
        check=True,
        cwd=ROOT,
    )


if __name__ == "__main__":
    main()