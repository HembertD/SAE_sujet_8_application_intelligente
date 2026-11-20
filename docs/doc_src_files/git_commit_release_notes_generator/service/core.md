# Core du service (Point d'entrée du pont Go ↔ Python)

Fichier source associé : src/git_commit_release_notes_generator/service/core.py

## Description

Ce module constitue le point d'entrée exécutable sous forme de sous-processus par le front-end Go (`src/cli/bridge/python_client.go`). Il orchestre l'ensemble des fonctionnalités du back-end Python :
- interaction et extraction des statuts et diffs Git via `GitWrapper` ;
- filtrage des fichiers et masquage des secrets ;
- interrogation du LLM via `generate_commit_message`, enrichie par la description des images modifiées via `describe_image_change` (module de vision) ;
- synthèse des release notes au format Markdown via `generate_release_notes` ;
- diagnostic et joignabilité du serveur Ollama, incluant la validation que le modèle configuré y est bien installé.

## Contrat d'interface (Règle d'or Go / Python)

1. **Sortie JSON stricte sur `stdout`** : seul un objet JSON unique est imprimé sur la sortie standard.
2. **Logs et diagnostics sur `stderr`** : toutes les traces, messages d'information et erreurs passent exclusivement par le module `logging` sur `sys.stderr`.
3. **Résilience sans crash** : toutes les exceptions sont interceptées pour renvoyer un JSON d'erreur propre `{"success": false, "error": "..."}`.

## Les 10 actions gérées (`--action <NOM>`)

| `--action` | Arguments | Format de sortie JSON | Rôle et source des données |
|---|---|---|---|
| `status` | `--repo` | `{"is_git_repo", "repo_path", "branch", "staged_count", "unstaged_count"}` | Détermine l'état du dépôt Git via `GitWrapper` |
| `diff` | `--repo` | `{"success", "files": [...], "staged_count"}` | Analyse et assainit les fichiers indexés (`GitWrapper._parse_diff`) |
| `stage-all` | `--repo` | `{"success", "message"}` | Indexe l'ensemble des fichiers modifiés (`git add -A`) |
| `generate-commit` | `--repo`, `--feedback` | `{"success", "type", "scope", "subject", "body", "raw_formatted"}` | Inférence du commit Conventional Commits avec support du feedback ; décrit aussi les images modifiées via un modèle de vision |
| `apply-commit` | `--repo`, `--message` | `{"success", "sha", "message"}` | Applique le commit sur la branche courante via `GitWrapper.commit()` |
| `push` | `--repo` | `{"success", "message"}` | Pousse vers le dépôt distant configuré via `git push` |
| `commits` | `--repo`, `--limit` | `{"success", "commits": [{"sha", "message", "author", "date", "tag"}]}` | Extrait l'historique complet ou récent des commits pour le sélecteur TUI |
| `release-notes` | `--repo`, `--from`, `--to` | `{"success", "from_tag", "to_tag", "markdown", "commit_count"}` | Synthèse Markdown des commits entre deux révisions via le LLM ; échec explicite (`success=False`) si l'appel échoue, aucun repli silencieux |
| `save-config` | `--base-url`, `--model`, `--timeout`, `--language` | `{"success", "message"}` | Notification de sauvegarde de configuration |
| `ping-ollama` | — | `{"success", "reachable", "latency_ms", "installed_models"}` | Vérification HTTP de joignabilité et des modèles disponibles sur Ollama |
| `check-env` | — | `{"success", "version", "executable", "status"}` | Vérification de l'intégrité de l'environnement Python et du module applicatif |

## Fonctions clés

- `action_status(repo_path)` : vérifie la validité du dépôt, gère les branches orphelines (detached HEAD) et dénombre les modifications.
- `action_diff(repo_path)` : expose chaque fichier avec son patch, ses compteurs `added`/`removed`, le drapeau `binary` et la détection `has_secrets_masked`.
- `action_generate_commit(repo_path, feedback)` : extrait le diff staged textuel (`get_staged_diff`) et les images modifiées (`get_staged_image_changes`), teste la joignabilité d'Ollama via `check_ollama_reachable(validate_model=True)` pour échouer rapidement en cas de problème réseau **ou si le modèle configuré n'est pas installé sur le serveur**, décrit chaque image modifiée via `describe_image_change()` (une image dont la description échoue est simplement ignorée, pas bloquante pour le reste), puis transmet l'ensemble (diffs textuels + descriptions d'images, sous forme de `DiffFile` uniformes) au module LLM avec la consigne optionnelle utilisateur (ou renvoie le commit statique si `MOCK_AI=True`). Échoue si ni diff textuel ni image n'est exploitable.
- `action_release_notes(repo_path, from_tag, to_tag)` : extrait l'historique, teste la joignabilité d'Ollama via `check_ollama_reachable(validate_model=True)` (échoue immédiatement avec `success=False` si injoignable **ou si le modèle n'est pas installé**, à l'identique de `action_generate_commit`) et délègue la génération à `ollama_client.release_notes.generate_release_notes()` (moteur unifié avec `action_generate_commit` : même socle `call_chat_json`, mêmes règles de retry/validation, mode `MOCK_AI` géré en interne par cette fonction). Si cet appel échoue (timeout, réponse invalide après retry), l'action renvoie `success=False` avec un message d'erreur explicite — **aucun repli silencieux** vers un formateur Regex qui simulerait une sortie LLM (cf. `bug_relase_note.md`, anomalie corrigée : avant ce fix, un échec était masqué par un Markdown généré localement renvoyé avec `success=True`).
- `action_ping_ollama()` : teste la joignabilité, mesure la latence et liste les modèles installés en déléguant à la fonction unifiée `check_ollama_reachable(fetch_models=True)`.

## Détails déplacés du code source

- Exécution en sous-processus : `python -m git_commit_release_notes_generator.service.core --action <NOM> [args]`, appelé par `src/cli/bridge/python_client.go`. Le processus sort toujours en code 0, même pour une action inconnue, car le client Go ignore entièrement `stdout` si le code de sortie est non nul.
- `action_status` : la branche courante gère le `detached HEAD` et le commit initial (repli sur les 7 premiers caractères du SHA, sinon `"HEAD"`). `staged_count` compte les lignes non vides de `git diff --cached --name-only`, `unstaged_count` additionne `git diff --name-only` et `untracked_files`.
- `action_diff` : chaque fichier expose `has_secrets_masked` (vrai si le `patch` contient `[REDACTED_SECRET]`).
- `action_generate_commit` / `action_release_notes` : `validate_model=True` fait échouer `check_ollama_reachable()` *avant* toute tentative de génération si `OLLAMA_MODEL` n'est pas dans la liste des modèles installés sur le serveur — au lieu de découvrir l'absence du modèle après coup via l'erreur HTTP 404 réactive de `call_chat_json()` (qui reste en place comme filet de sécurité si la liste n'a pas pu être récupérée).
- `action_generate_commit` : chaque image décrite avec succès devient un `DiffFile` texte (`is_binary=False`, `added=0`, `removed=0`, `patch="[Analyse visuelle] <description>"`) ajouté à la liste transmise à `generate_commit_message()`, au même titre qu'un fichier texte ou document modifié.
- `action_release_notes` : valeurs par défaut `from_tag="v0.1.0"`, `to_tag="HEAD"`. Vérification stricte de joignabilité avant toute génération (échec immédiat si injoignable), puis tentative d'inférence LLM via `generate_release_notes()` (module `ollama_client`, plus de second appel HTTP Ollama dupliqué localement). Toute exception levée par cet appel (timeout, retry épuisé, etc.) produit un retour `success=False` avec `markdown=""` et un message d'erreur explicite ; il n'existe plus de formateur de repli local (`_format_release_notes_fallback` a été supprimé) qui aurait masqué l'échec sous un `success=True` trompeur.
- `action_save_config` : simple notification, renvoie toujours un succès sans écrire de fichier.

## Commentaires de fonction (dupliqués du code source)

- `action_status()` : Calcule l'état courant du dépôt Git.
- `action_diff()` : Expose les fichiers stagés avec patchs assainis.
- `action_stage_all()` : Indexe toutes les modifications du dépôt.
- `action_generate_commit()` : Génère un message de commit depuis le diff stagé.
- `action_apply_commit()` : Applique le message de commit validé au dépôt.
- `action_push()` : Pousse les commits locaux vers le dépôt distant.
- `action_release_notes()` : Génère les release notes Markdown entre deux révisions.
- `action_save_config()` : Notifie la prise en compte de la configuration.
- `action_ping_ollama()` : Diagnostique la joignabilité d'Ollama avec modèles installés.
- `main()` : Exécute l'action demandée et affiche le JSON résultat.
