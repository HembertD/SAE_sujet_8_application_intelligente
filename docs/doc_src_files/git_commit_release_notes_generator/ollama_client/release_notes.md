# Génération de release notes

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/release_notes.py

## Description

Ce module génère le Markdown des release notes entre deux tags, via Ollama, en réutilisant le même socle HTTP et la même rigueur que `generate_commit_message()` et `suggest_conflict_resolution()` : sortie JSON contrainte côté serveur, retry unique sur contenu invalide. Il remplace un appel HTTP Ollama précédemment dupliqué directement dans `service/core.py`.

## Périmètre

Ce module ne lit ni n'écrit aucun fichier, ne résout pas les tags Git (rôle du `GitWrapper`, module B) et n'écrit rien sur `stdout`. Le mode hors-ligne (`MOCK_AI`) et le repli déterministe en cas d'échec de génération restent gérés par l'appelant (`service/core.py::action_release_notes`), pas ici.

## Flux de travail

1. Rejeter immédiatement une liste de commits vide (`ValueError`), sans appel réseau.
2. Construire le prompt système et le résumé des commits (sha + première ligne du message).
3. Appeler Ollama via le socle partagé `llm.call_chat_json()` avec un schéma JSON dédié (`{"markdown": str}`).
4. Vérifier que la réponse est bien un JSON conforme et que le `markdown` n'est pas vide.
5. En cas d'erreur (validation ou contenu non-JSON), relancer une seconde fois avec l'erreur en contexte.
6. Retourner le Markdown validé (ou lever `ReleaseNotesError` après le second échec).

## Points clés

- format : le modèle répond avec un schéma JSON strict (`markdown`) plutôt qu'une chaîne déjà formatée ;
- validation : le code rejette tout `markdown` manquant ou vide ;
- robustesse : même mécanisme de retry unique que les autres générateurs du module (validation métier ET contenu non-JSON) ;
- réutilisation : s'appuie sur `llm.call_chat_json()` (socle HTTP partagé) plutôt que de dupliquer un troisième appel Ollama indépendant ;
- simulation hors-ligne : si `config.MOCK_AI` est activé, délègue à `mock_llm.mock_release_notes()` sans requête réseau.

## Fonctions exposées

- `generate_release_notes(commits: list[Any], from_tag: str, to_tag: str) -> str` : fonction principale, décrite ci-dessus.

## Dépendances

- [config.md](../config.md)
- [mock_llm.md](mock_llm.md)
- [exceptions.md](exceptions.md)
- [llm.md](llm.md)
- [release_notes_system.md](prompts/release_notes_system.md)

## Détails déplacés du code source

- `commits` : liste d'objets quelconques exposant `.sha` et `.message` (produits par le `GitWrapper`, module B) ; typé `list[Any]` pour ne pas créer de dépendance d'import vers un autre module.
- `generate_release_notes()` : un seul retry est tenté ; la réponse invalide est réinjectée comme message `assistant` (JSON sérialisé, ou contenu brut si la réponse n'était pas du JSON), suivie d'un message `user` signalant l'erreur et exigeant une proposition conforme.
- `_build_commits_prompt()` : assemble le résumé textuel des commits (sha + première ligne du message) et les tags.
- `_validate(raw)` : retourne `None` si valide, sinon un message explicite (`markdown` manquant/vide).
