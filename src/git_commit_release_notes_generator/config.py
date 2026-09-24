"""Configuration du module C (client Ollama), lue directement depuis le fichier .env."""
import os
from pathlib import Path


def load_env() -> dict[str, str]:
    """Charge les paramètres de configuration depuis le fichier .env de l'application."""
    base_dir = Path(__file__).resolve().parent  # src/git_commit_release_notes_generator
    app_root = base_dir.parent.parent  # racine du dépôt de l'application

    candidates: list[Path] = []

    # 1. Chemin racine explicite transmis via l'environnement (ex: par le Go CLI)
    custom_root = os.environ.get("SMART_COMMIT_APP_ROOT")
    if custom_root:
        c_path = Path(custom_root)
        candidates.append(c_path / ".env")

    # 2. Emplacement standard : strictement à la racine du dépôt du projet
    candidates.append(app_root / ".env")

    env_data: dict[str, str] = {}
    seen = set()
    for candidate in candidates:
        cand_resolved = candidate.resolve()
        if cand_resolved in seen:
            continue
        seen.add(cand_resolved)

        if cand_resolved.is_file():
            try:
                with open(cand_resolved, "r", encoding="utf-8") as f:
                    for line in f:
                        line = line.strip()
                        if not line or line.startswith("#") or "=" not in line:
                            continue
                        k, v = line.split("=", 1)
                        k = k.strip()
                        v = v.strip().strip("'\"")
                        if k:
                            env_data[k] = v
                break
            except Exception:
                pass

    # Surcharge optionnelle par l'environnement système direct (si défini explicitement)
    for k in ("OLLAMA_BASE_URL", "OLLAMA_MODEL", "OLLAMA_TIMEOUT_S", "APP_LANGUAGE", "MOCK_INTERFACE"):
        if k in os.environ and os.environ[k]:
            env_data[k] = os.environ[k]

    return env_data


ENV = load_env()

if "OLLAMA_BASE_URL" not in ENV:
    ENV["OLLAMA_BASE_URL"] = "http://10.22.28.190:11434"
if "OLLAMA_MODEL" not in ENV:
    ENV["OLLAMA_MODEL"] = "gemma4:26b"
if "OLLAMA_TIMEOUT_S" not in ENV:
    ENV["OLLAMA_TIMEOUT_S"] = "300"
if "APP_LANGUAGE" not in ENV:
    ENV["APP_LANGUAGE"] = "fr"

OLLAMA_BASE_URL = ENV["OLLAMA_BASE_URL"]
OLLAMA_MODEL = ENV["OLLAMA_MODEL"]
OLLAMA_TIMEOUT_S = float(ENV["OLLAMA_TIMEOUT_S"])
APP_LANGUAGE = ENV["APP_LANGUAGE"]


