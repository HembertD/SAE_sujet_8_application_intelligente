import os
from pathlib import Path

# Convertit une valeur texte ou booléenne en booléen.
def parse_bool(val: str | bool | None) -> bool:
    if isinstance(val, bool):
        return val
    if not val:
        return False
    return str(val).strip().lower() in ("true", "1", "yes", "oui")

# Charge les variables depuis le fichier .env et l'environnement système.
def load_env() -> dict[str, str]:
    base_dir = Path(__file__).resolve().parent
    app_root = base_dir.parent.parent

    candidates: list[Path] = []

    custom_root = os.environ.get("SMART_COMMIT_APP_ROOT")
    if custom_root:
        c_path = Path(custom_root)
        candidates.append(c_path / ".env")

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

    for k in (
        "OLLAMA_BASE_URL",
        "OLLAMA_MODEL",
        "OLLAMA_TIMEOUT_S",
        "OLLAMA_NUM_CTX",
        "APP_LANGUAGE",
        "MOCK_INTERFACE",
        "MOCK_AI",
    ):
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
if "OLLAMA_NUM_CTX" not in ENV:
    ENV["OLLAMA_NUM_CTX"] = "32768"
if "APP_LANGUAGE" not in ENV:
    ENV["APP_LANGUAGE"] = "fr"
if "MOCK_AI" not in ENV:
    ENV["MOCK_AI"] = "false"

OLLAMA_BASE_URL = ENV["OLLAMA_BASE_URL"]
OLLAMA_MODEL = ENV["OLLAMA_MODEL"]
OLLAMA_TIMEOUT_S = float(ENV["OLLAMA_TIMEOUT_S"])
OLLAMA_NUM_CTX = int(ENV["OLLAMA_NUM_CTX"])
APP_LANGUAGE = ENV["APP_LANGUAGE"]
MOCK_AI = parse_bool(ENV.get("MOCK_AI", "false"))
