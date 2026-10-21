"""Crée un environnement virtuel Python autonome à la racine de l'application.

Outil de dev/installation, pas le module lui-même : celui-ci peut écrire sur
stdout. Résout la carte Trello "environnement Python" : plutôt que de
dépendre d'un interpréteur système deviné (python/python3) ou d'un chemin
d'environnement virtuel personnel codé en dur côté Go, ce script crée un
environnement dédié dans les fichiers de l'application elle-même
(<racine>/.venv), que le pont Go->Python sait retrouver automatiquement.

Usage : python scripts/setup_python_env.py
"""
from __future__ import annotations

import subprocess
import sys
import venv
from pathlib import Path

APP_ROOT = Path(__file__).parent.parent
VENV_DIR = APP_ROOT / ".venv"
REQUIREMENTS = APP_ROOT / "requirements.txt"


def _venv_python(venv_dir: Path) -> Path:
    if sys.platform == "win32":
        return venv_dir / "Scripts" / "python.exe"
    return venv_dir / "bin" / "python"


def main() -> None:
    if VENV_DIR.exists():
        print(f"Environnement déjà présent : {VENV_DIR}")
    else:
        print(f"Création de l'environnement virtuel dans {VENV_DIR}...")
        venv.create(VENV_DIR, with_pip=True)

    venv_python = _venv_python(VENV_DIR)
    if not venv_python.exists():
        raise SystemExit(f"Échec : {venv_python} introuvable après création de l'environnement.")

    print("Installation des dépendances (requirements.txt)...")
    subprocess.run(
        [str(venv_python), "-m", "pip", "install", "--upgrade", "pip"],
        check=True,
    )
    subprocess.run(
        [str(venv_python), "-m", "pip", "install", "-r", str(REQUIREMENTS)],
        check=True,
    )

    print(f"\nEnvironnement prêt : {venv_python}")
    print("Le pont Go détecte automatiquement cet environnement (.venv à la racine).")
    print("Pour forcer un autre interpréteur, renseigne PYTHON_BIN dans le .env.")


if __name__ == "__main__":
    main()
