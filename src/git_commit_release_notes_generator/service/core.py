"""Point d'entrée du backend Python appelé par le CLI Go.

Ce fichier est exécuté en sous-processus par src/cli/bridge/python_client.go :
python -m git_commit_release_notes_generator.service.core --action <NOM> [args]

Règles impératives :
- Un unique objet JSON valide est écrit sur stdout à la fin.
- Tous les messages de logs et diagnostics passent sur stderr via logging.
"""
from __future__ import annotations

import argparse
import datetime
import json
import logging
import re
import sys
import time
import urllib.error
import urllib.request

import os
from pathlib import Path

# Chargement du fichier .env pour synchroniser la configuration (ex: OLLAMA_TIMEOUT_S)
for _candidate in (
    Path.cwd() / ".env",
    Path.cwd() / "src" / ".env",
    Path(__file__).resolve().parent.parent.parent.parent / ".env",
    Path(__file__).resolve().parent.parent.parent.parent / "src" / ".env",
):
    if _candidate.is_file():
        try:
            with open(_candidate, "r", encoding="utf-8") as _f:
                for _line in _f:
                    _line = _line.strip()
                    if _line and not _line.startswith("#") and "=" in _line:
                        _k, _v = _line.split("=", 1)
                        _k = _k.strip()
                        _v = _v.strip().strip("'\"")
                        if _k and _k not in os.environ:
                            os.environ[_k] = _v
            break
        except Exception:
            pass

from git_commit_release_notes_generator.config import (
    OLLAMA_BASE_URL,
    OLLAMA_MODEL,
    OLLAMA_TIMEOUT_S,
)
from git_commit_release_notes_generator.ollama_client.exceptions import (
    CommitMessageValidationError,
    OllamaCallError,
)
from git_commit_release_notes_generator.ollama_client.llm import generate_commit_message
from git_commit_release_notes_generator.service.git_wrapper import (
    GitWrapper,
    GitWrapperError,
)

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO, stream=sys.stderr, format="%(levelname)s: %(message)s")


def action_status(repo_path: str = ".") -> dict:
    """Renvoie l'état du dépôt Git pour le TUI Go.

    Format attendu côté Go :
    {"is_git_repo": bool, "repo_path": str, "branch": str, "staged_count": int, "unstaged_count": int}
    """
    try:
        wrapper = GitWrapper(repo_path)
        repo = wrapper.repo

        path_str = str(repo.working_tree_dir or repo.git_dir)

        # Récupération de la branche courante (supporte detached HEAD et initial commit)
        try:
            branch = repo.active_branch.name
        except TypeError:
            branch = repo.head.commit.hexsha[:7] if repo.head.is_valid() else "HEAD"
        except Exception:
            branch = "HEAD"

        # Nombre de fichiers indexés (staged)
        try:
            staged_lines = [l for l in repo.git.diff("--cached", "--name-only").splitlines() if l.strip()]
            staged_count = len(staged_lines)
        except Exception:
            staged_count = 0

        # Nombre de fichiers non indexés (unstaged + untracked)
        try:
            unstaged_lines = [l for l in repo.git.diff("--name-only").splitlines() if l.strip()]
            unstaged_count = len(unstaged_lines) + len(repo.untracked_files)
        except Exception:
            unstaged_count = 0

        return {
            "is_git_repo": True,
            "repo_path": path_str,
            "branch": branch,
            "staged_count": staged_count,
            "unstaged_count": unstaged_count,
        }
    except (GitWrapperError, Exception) as e:
        logger.debug(f"action_status non git repo: {e}")
        return {
            "is_git_repo": False,
            "repo_path": str(repo_path),
            "branch": "",
            "staged_count": 0,
            "unstaged_count": 0,
            "error": str(e),
        }


def action_diff(repo_path: str = ".") -> dict:
    """Renvoie le diff indexé analysé et assaini pour le TUI Go.

    Format attendu côté Go :
    {"success": bool, "files": [{"path", "status", "added", "removed", "binary", "patch", "has_secrets_masked"}], "staged_count": int}
    """
    try:
        wrapper = GitWrapper(repo_path)
        raw_diff = wrapper.repo.git.diff("--cached", "--no-color", "--diff-filter=ACMRTUXB")
        if not raw_diff.strip():
            return {
                "success": True,
                "files": [],
                "staged_count": 0,
            }

        parsed_files = wrapper._parse_diff(raw_diff)
        files = []
        for df in parsed_files:
            patch = df.patch
            # Tronquer le patch textuel volumineux pour éviter d'engorger la mémoire
            if not df.is_binary and len(patch) > 8000:
                patch = patch[:8000] + "\n\n... [DIFF TRUNCATED: TROP VOLUMINEUX] ..."

            files.append({
                "path": df.path,
                "status": df.status,
                "added": df.added,
                "removed": df.removed,
                "binary": bool(df.is_binary),
                "patch": patch,
                "has_secrets_masked": "[REDACTED_SECRET]" in patch,
            })

        return {
            "success": True,
            "files": files,
            "staged_count": len(files),
        }
    except (GitWrapperError, Exception) as e:
        logger.error(f"action_diff error: {e}")
        return {
            "success": False,
            "error": str(e),
            "files": [],
            "staged_count": 0,
        }


def action_stage_all(repo_path: str = ".") -> dict:
    """Indexe l'ensemble des modifications locales (git add -A).

    Format attendu côté Go :
    {"success": bool, "message": str}
    """
    try:
        wrapper = GitWrapper(repo_path)
        wrapper.repo.git.add("-A")
        return {
            "success": True,
            "message": "Tous les fichiers ont été indexés (git add -A)",
        }
    except (GitWrapperError, Exception) as e:
        logger.error(f"action_stage_all error: {e}")
        return {
            "success": False,
            "error": f"Erreur lors de l'indexation : {e}",
        }


def action_generate_commit(repo_path: str = ".") -> dict:
    """Génère une proposition de message de commit via le LLM.

    Format attendu côté Go :
    {"success": bool, "type": str, "scope": str, "subject": str, "body": str, "raw_formatted": str}
    """
    try:
        wrapper = GitWrapper(repo_path)
        diff_files = wrapper.get_staged_diff()
    except (GitWrapperError, Exception) as e:
        logger.error(f"action_generate_commit git error: {e}")
        return {"success": False, "error": str(e)}

    if not diff_files:
        return {
            "success": False,
            "error": "Aucune modification indexée (staged) exploitable pour le modèle LLM.",
        }

    try:
        msg = generate_commit_message(diff_files)
    except (OllamaCallError, CommitMessageValidationError, ValueError) as e:
        logger.warning(f"Génération commit échouée : {e}")
        return {"success": False, "error": str(e)}
    except Exception as e:
        logger.error(f"Erreur inattendue génération commit : {e}")
        return {"success": False, "error": f"Erreur inattendue : {e}"}

    raw_formatted = f"{msg.type}" + (f"({msg.scope})" if msg.scope else "") + f": {msg.subject}"
    if msg.body:
        raw_formatted += f"\n\n{msg.body}"

    return {
        "success": True,
        "type": msg.type,
        "scope": msg.scope or "",
        "subject": msg.subject,
        "body": msg.body or "",
        "raw_formatted": raw_formatted,
    }


def action_apply_commit(repo_path: str = ".", message: str = "") -> dict:
    """Applique le commit Git sur les modifications indexées.

    Format attendu côté Go :
    {"success": bool, "sha": str, "message": str}
    """
    if not message or not message.strip():
        return {
            "success": False,
            "error": "Le message de commit ne peut pas être vide.",
        }

    try:
        wrapper = GitWrapper(repo_path)
        sha = wrapper.commit(message.strip())
        return {
            "success": True,
            "sha": sha,
            "message": f"Commit appliqué avec succès ({sha}).",
        }
    except (GitWrapperError, Exception) as e:
        logger.error(f"action_apply_commit error: {e}")
        return {
            "success": False,
            "error": str(e),
        }


def action_push(repo_path: str = ".") -> dict:
    """Pousse les modifications vers le dépôt distant.

    Format attendu côté Go :
    {"success": bool, "message": str}
    """
    try:
        wrapper = GitWrapper(repo_path)
        repo = wrapper.repo
        if not repo.remotes:
            return {
                "success": False,
                "error": "Aucun dépôt distant (remote) configuré pour ce dépôt Git local.",
            }
        _ = repo.git.push()
        return {
            "success": True,
            "message": "Modifications poussées avec succès vers le dépôt distant.",
        }
    except (GitWrapperError, Exception) as e:
        logger.error(f"action_push error: {e}")
        return {
            "success": False,
            "error": f"Échec lors du git push : {e}",
        }


def _format_release_notes_fallback(commits: list, from_tag: str, to_tag: str) -> str:
    """Formate et catégorise les commits en Markdown structuré selon Conventional Commits."""
    categories: dict[str, list[str]] = {
        "feat": [],
        "fix": [],
        "docs": [],
        "refactor": [],
        "perf": [],
        "other": [],
    }

    pattern = re.compile(
        r"^(?P<type>feat|fix|docs|refactor|perf|chore|test|ci|build|revert)(?:\((?P<scope>[^)]+)\))?:\s*(?P<subject>.+)$",
        re.IGNORECASE,
    )

    for c in commits:
        first_line = c.message.splitlines()[0].strip() if c.message else ""
        match = pattern.match(first_line)
        if match:
            ctype = match.group("type").lower()
            scope = match.group("scope")
            subject = match.group("subject")
            scope_part = f"({scope})" if scope else ""
            line = f"- **{ctype}{scope_part}**: {subject} (`{c.sha}`)"
            if ctype in categories:
                categories[ctype].append(line)
            else:
                categories["other"].append(line)
        else:
            categories["other"].append(f"- {first_line} (`{c.sha}`)")

    today = datetime.date.today().strftime("%Y-%m-%d")
    lines = [
        f"# Release Notes ({from_tag} ➔ {to_tag})",
        f"Date : {today}",
        "",
    ]

    sections = [
        ("feat", "### 🚀 Nouveautés (Features)"),
        ("fix", "### 🐛 Corrections de bogues (Fixes)"),
        ("refactor", "### ⚡ Refactoring & Améliorations"),
        ("perf", "### 🏎️ Performances"),
        ("docs", "### 📚 Documentation"),
        ("other", "### 🛠️ Autres changements"),
    ]

    has_any = False
    for key, title in sections:
        items = categories.get(key, [])
        if items:
            has_any = True
            lines.append(title)
            lines.extend(items)
            lines.append("")

    if not has_any:
        lines.append("Aucun changement notable.")

    return "\n".join(lines).strip()


def action_release_notes(repo_path: str = ".", from_tag: str = "", to_tag: str = "") -> dict:
    """Génère les notes de version entre deux tags.

    Format attendu côté Go :
    {"success": bool, "from_tag": str, "to_tag": str, "markdown": str, "commit_count": int}
    """
    if not from_tag:
        from_tag = "v0.1.0"
    if not to_tag:
        to_tag = "HEAD"

    try:
        wrapper = GitWrapper(repo_path)
        commits = wrapper.get_commits_between_tags(from_tag, to_tag)
    except (GitWrapperError, Exception) as e:
        logger.error(f"action_release_notes git error: {e}")
        return {
            "success": False,
            "error": str(e),
            "from_tag": from_tag,
            "to_tag": to_tag,
            "markdown": "",
            "commit_count": 0,
        }

    if not commits:
        return {
            "success": True,
            "from_tag": from_tag,
            "to_tag": to_tag,
            "markdown": f"# Release Notes ({from_tag} ➔ {to_tag})\n\nAucun commit trouvé dans cet intervalle.",
            "commit_count": 0,
        }

    # Tentative d'inférence LLM si Ollama est joignable
    md_content: str | None = None
    try:
        # Test rapide de connectivité avant inférence
        v_req = urllib.request.Request(f"{OLLAMA_BASE_URL.rstrip('/')}/api/version", headers={"User-Agent": "SmartCommit/1.0"})
        with urllib.request.urlopen(v_req, timeout=1.5):
            pass

        commits_summary = "\n".join([f"- [{c.sha}] {c.message.splitlines()[0]}" for c in commits])
        prompt = (
            f"Rédige des Release Notes professionnelles au format Markdown en français pour la version {from_tag} -> {to_tag}.\n"
            f"Voici la liste des commits :\n{commits_summary}\n\n"
            "Organise la réponse avec les sections Markdown :\n"
            f"# Release Notes ({from_tag} ➔ {to_tag})\n"
            "### 🚀 Nouveautés (Features)\n"
            "### 🐛 Corrections de bogues (Fixes)\n"
            "### 📚 Documentation & Maintenance\n"
            "Ne renvoie que le Markdown pur, sans balises ```markdown englobantes."
        )
        body = {
            "model": OLLAMA_MODEL,
            "messages": [
                {"role": "system", "content": "Tu es un expert Git qui rédige des release notes claires et concises en Markdown."},
                {"role": "user", "content": prompt},
            ],
            "stream": False,
        }
        req = urllib.request.Request(
            url=f"{OLLAMA_BASE_URL.rstrip('/')}/api/chat",
            data=json.dumps(body).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with urllib.request.urlopen(req, timeout=min(OLLAMA_TIMEOUT_S, 20.0)) as resp:
            payload = json.loads(resp.read().decode("utf-8"))
            content = payload.get("message", {}).get("content")
            if content and isinstance(content, str) and content.strip():
                md_content = content.strip()
    except Exception as e:
        logger.info(f"Ollama indisponible pour release-notes ({e}), repli sur le formateur déterministe.")

    if not md_content:
        md_content = _format_release_notes_fallback(commits, from_tag, to_tag)

    return {
        "success": True,
        "from_tag": from_tag,
        "to_tag": to_tag,
        "markdown": md_content,
        "commit_count": len(commits),
    }


def action_save_config(base_url: str = "", model: str = "", timeout: str = "", language: str = "") -> dict:
    """Action de notification de la sauvegarde de configuration.

    Format attendu côté Go :
    {"success": bool, "message": str}
    """
    return {
        "success": True,
        "message": "Configuration notifiée avec succès au backend Python.",
    }


def action_ping_ollama() -> dict:
    """Teste la connexion avec le serveur Ollama de l'IUT.

    Format attendu côté Go :
    {"success": bool, "reachable": bool, "latency_ms": int, "installed_models": [str]}
    """
    start_time = time.perf_counter()
    installed_models: list[str] = []

    try:
        version_url = f"{OLLAMA_BASE_URL.rstrip('/')}/api/version"
        req_v = urllib.request.Request(version_url, headers={"User-Agent": "SmartCommit/1.0"})
        with urllib.request.urlopen(req_v, timeout=3.0) as resp:
            _ = resp.read()

        latency_ms = int((time.perf_counter() - start_time) * 1000)

        tags_url = f"{OLLAMA_BASE_URL.rstrip('/')}/api/tags"
        req_t = urllib.request.Request(tags_url, headers={"User-Agent": "SmartCommit/1.0"})
        with urllib.request.urlopen(req_t, timeout=3.0) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            raw_models = data.get("models", [])
            for m in raw_models:
                if isinstance(m, dict) and "name" in m:
                    installed_models.append(str(m["name"]))

        return {
            "success": True,
            "reachable": True,
            "latency_ms": latency_ms,
            "installed_models": installed_models,
        }
    except Exception as e:
        logger.debug(f"ping-ollama failed: {e}")
        return {
            "success": True,
            "reachable": False,
            "latency_ms": 0,
            "installed_models": [],
            "error": str(e),
        }


def main() -> None:
    parser = argparse.ArgumentParser(description="Service Core CLI pour Smart Commit Generator")
    parser.add_argument("--action", required=True, help="Action à exécuter")
    parser.add_argument("--repo", default=".", help="Chemin vers le dépôt Git cible")
    parser.add_argument("--feedback", default="", help="Consigne utilisateur pour réviser le commit")
    parser.add_argument("--message", default="", help="Message de commit à appliquer")
    parser.add_argument("--from", dest="from_tag", default="", help="Tag ou révision de départ")
    parser.add_argument("--to", dest="to_tag", default="", help="Tag ou révision d'arrivée")
    parser.add_argument("--base-url", default="", help="URL du serveur Ollama")
    parser.add_argument("--model", default="", help="Nom du modèle Ollama")
    parser.add_argument("--timeout", default="", help="Timeout en secondes")
    parser.add_argument("--language", default="", help="Langue de l'application")

    args = parser.parse_args()

    handlers = {
        "status": lambda: action_status(args.repo),
        "diff": lambda: action_diff(args.repo),
        "stage-all": lambda: action_stage_all(args.repo),
        "generate-commit": lambda: action_generate_commit(args.repo),
        "apply-commit": lambda: action_apply_commit(args.repo, message=args.message),
        "push": lambda: action_push(args.repo),
        "release-notes": lambda: action_release_notes(args.repo, from_tag=args.from_tag, to_tag=args.to_tag),
        "save-config": lambda: action_save_config(args.base_url, args.model, args.timeout, args.language),
        "ping-ollama": action_ping_ollama,
    }

    handler = handlers.get(args.action)
    if handler is None:
        result = {"success": False, "error": f"action inconnue: {args.action}"}
        print(json.dumps(result, ensure_ascii=False))
        sys.exit(1)

    try:
        result = handler()
    except Exception as e:
        logger.exception(f"Erreur non rattrapée dans le handler {args.action}: {e}")
        result = {"success": False, "error": str(e)}

    print(json.dumps(result, ensure_ascii=False))


if __name__ == "__main__":
    main()
