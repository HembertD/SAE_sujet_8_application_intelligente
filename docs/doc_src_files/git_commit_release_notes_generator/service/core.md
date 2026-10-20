# Core du service (Point d'entrée du pont Go ↔ Python)

Fichier source associé : src/git_commit_release_notes_generator/service/core.py

## Description

Ce module constitue le point d'entrée exécutable sous forme de sous-processus par le front-end Go (`src/cli/bridge/python_client.go`). Il orchestre l'ensemble des fonctionnalités du back-end Python :
- interaction et extraction des statuts et diffs Git via `GitWrapper` ;
- filtrage des fichiers et masquage des secrets ;
- interrogation du LLM via `generate_commit_message` ;
- synthèse des release notes au format Markdown via `generate_release_notes` ;
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
- `action_generate_commit(repo_path, feedback)` : extrait le diff staged textuel, teste la joignabilité d'Ollama via `check_ollama_reachable()` pour échouer rapidement en cas de problème réseau, et le transmet au module LLM avec la consigne optionnelle utilisateur (ou renvoie le commit statique si `MOCK_AI=True`).
- `action_release_notes(repo_path, from_tag, to_tag)` : extrait l'historique, teste la joignabilité d'Ollama via `check_ollama_reachable()` (échoue immédiatement avec `success=False` si injoignable, à l'identique de `action_generate_commit`) et délègue la génération à `ollama_client.release_notes.generate_release_notes()` (moteur unifié avec `action_generate_commit` : même socle `call_chat_json`, mêmes règles de retry/validation, mode `MOCK_AI` géré en interne par cette fonction), avec repli sur le formateur déterministe si l'appel échoue pour quelque raison que ce soit.
- `action_ping_ollama()` : teste la joignabilité, mesure la latence et liste les modèles installés en déléguant à la fonction unifiée `check_ollama_reachable(fetch_models=True)`.

## Détails déplacés du code source

- Exécution en sous-processus : `python -m git_commit_release_notes_generator.service.core --action <NOM> [args]`, appelé par `src/cli/bridge/python_client.go`. Le processus sort toujours en code 0, même pour une action inconnue, car le client Go ignore entièrement `stdout` si le code de sortie est non nul.
- `action_status` : la branche courante gère le `detached HEAD` et le commit initial (repli sur les 7 premiers caractères du SHA, sinon `"HEAD"`). `staged_count` compte les lignes non vides de `git diff --cached --name-only`, `unstaged_count` additionne `git diff --name-only` et `untracked_files`.
- `action_diff` : chaque fichier expose `has_secrets_masked` (vrai si le `patch` contient `[REDACTED_SECRET]`).
- `action_release_notes` : valeurs par défaut `from_tag="v0.1.0"`, `to_tag="HEAD"`. Vérification stricte de joignabilité avant toute génération (échec immédiat si injoignable), puis tentative d'inférence LLM via `generate_release_notes()` (module `ollama_client`, plus de second appel HTTP Ollama dupliqué localement), et repli sur le formateur déterministe `_format_release_notes_fallback` si cet appel lève une exception quelconque.
- `_format_release_notes_fallback` : catégorise chaque première ligne de commit via une expression Conventional Commits (`feat|fix|docs|refactor|perf|chore|test|ci|build|revert`, `scope` optionnel) en sections `feat`, `fix`, `refactor`, `perf`, `docs`, `other`, avec le SHA court entre backticks ; ajoute la date du jour et `Aucun changement notable.` si aucune section n'est remplie.
- `action_save_config` : simple notification, renvoie toujours un succès sans écrire de fichier.

## Commentaires de fonction (dupliqués du code source)

- `action_status()` : Calcule l'état courant du dépôt Git.
- `action_diff()` : Expose les fichiers stagés avec patchs assainis.
- `action_stage_all()` : Indexe toutes les modifications du dépôt.
- `action_generate_commit()` : Génère un message de commit depuis le diff stagé.
- `action_apply_commit()` : Applique le message de commit validé au dépôt.
- `action_push()` : Pousse les commits locaux vers le dépôt distant.
- `_format_release_notes_fallback()` : Formate les release notes sans appel au LLM.
- `action_release_notes()` : Génère les release notes Markdown entre deux révisions.
- `action_save_config()` : Notifie la prise en compte de la configuration.
- `action_ping_ollama()` : Diagnostique la joignabilité d'Ollama avec modèles installés.
- `main()` : Exécute l'action demandée et affiche le JSON résultat.
