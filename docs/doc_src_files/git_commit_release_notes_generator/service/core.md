# Core du service (Point d'entrée du pont Go ↔ Python)

Fichier source associé : src/git_commit_release_notes_generator/service/core.py

## Description

Ce module constitue le point d'entrée exécutable sous forme de sous-processus par le front-end Go (`src/cli/bridge/python_client.go`). Il orchestre l'ensemble des fonctionnalités du back-end Python :
- interaction et extraction des statuts et diffs Git via `GitWrapper` ;
- filtrage des fichiers et masquage des secrets ;
- interrogation du LLM via `generate_commit_message` ;
- synthèse des release notes au format Markdown ;
- diagnostic et joignabilité du serveur Ollama.

## Contrat d'interface (Règle d'or Go / Python)

1. **Sortie JSON stricte sur `stdout`** : seul un objet JSON unique est imprimé sur la sortie standard.
2. **Logs et diagnostics sur `stderr`** : toutes les traces, messages d'information et erreurs passent exclusivement par le module `logging` sur `sys.stderr`.
3. **Résilience sans crash** : toutes les exceptions sont interceptées pour renvoyer un JSON d'erreur propre `{"success": false, "error": "..."}`.

## Les 9 actions gérées (`--action <NOM>`)

| `--action` | Arguments | Format de sortie JSON | Rôle et source des données |
|---|---|---|---|
| `status` | `--repo` | `{"is_git_repo", "repo_path", "branch", "staged_count", "unstaged_count"}` | Détermine l'état du dépôt Git via `GitWrapper` |
| `diff` | `--repo` | `{"success", "files": [...], "staged_count"}` | Analyse et assainit les fichiers indexés (`GitWrapper._parse_diff`) |
| `stage-all` | `--repo` | `{"success", "message"}` | Indexe l'ensemble des fichiers modifiés (`git add -A`) |
| `generate-commit` | `--repo`, `--feedback` | `{"success", "type", "scope", "subject", "body", "raw_formatted"}` | Inférence du commit Conventional Commits avec support du feedback |
| `apply-commit` | `--repo`, `--message` | `{"success", "sha", "message"}` | Applique le commit sur la branche courante via `GitWrapper.commit()` |
| `push` | `--repo` | `{"success", "message"}` | Pousse vers le dépôt distant configuré via `git push` |
| `release-notes` | `--repo`, `--from`, `--to` | `{"success", "from_tag", "to_tag", "markdown", "commit_count"}` | Synthèse Markdown des commits entre deux révisions (LLM + repli déterministe) |
| `save-config` | `--base-url`, `--model`, `--timeout`, `--language` | `{"success", "message"}` | Notification de sauvegarde de configuration |
| `ping-ollama` | — | `{"success", "reachable", "latency_ms", "installed_models"}` | Vérification HTTP de joignabilité et des modèles disponibles sur Ollama |

## Fonctions clés

- `action_status(repo_path)` : vérifie la validité du dépôt, gère les branches orphelines (detached HEAD) et dénombre les modifications.
- `action_diff(repo_path)` : expose chaque fichier avec son patch, ses compteurs `added`/`removed`, le drapeau `binary` et la détection `has_secrets_masked`.
- `action_generate_commit(repo_path, feedback)` : extrait le diff staged textuel et le transmet au module LLM avec la consigne optionnelle utilisateur.
- `action_release_notes(repo_path, from_tag, to_tag)` : extrait l'historique et génère une synthèse catégorisée (Features, Fixes, Refactoring, Docs) avec repli déterministe en cas de déconnexion d'Ollama.
- `action_ping_ollama()` : teste `/api/version` et `/api/tags` avec mesure précise de la latence réseau en millisecondes.
