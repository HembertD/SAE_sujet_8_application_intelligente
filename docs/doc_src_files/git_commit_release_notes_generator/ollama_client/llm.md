# Client LLM

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/llm.py

## Description

Ce module gère la génération du message de commit via Ollama. Il construit le résumé du diff, envoie une requête structurée à l’API de chat, valide la réponse JSON et la transforme en objet métier.

## Flux de travail

1. Rejeter immédiatement une liste de fichiers vide (`ValueError`), sans appel réseau.
2. Lire les fichiers différences et les convertir en résumé exploitable.
3. Construire le prompt système et le message utilisateur.
4. Appeler Ollama via une requête JSON structurée, avec une fenêtre de contexte (`num_ctx`) fixée explicitement.
5. Vérifier que la réponse est bien un JSON conforme.
6. En cas d’erreur, relancer une seconde fois avec le message invalidé en contexte.
7. Convertir la réponse validée en objet CommitMessage.

## Points clés

- format : le modèle répond avec un schéma JSON strict ;
- validation : le code refuse les messages trop longs, absents ou non conformes ;
- robustesse : un retry est tenté pour corriger une réponse invalide ;
- fenêtre de contexte : `options.num_ctx` est fixé explicitement dans la requête (`OLLAMA_NUM_CTX`, config.py) car Ollama utilise sinon une fenêtre par défaut bien plus petite que celle annoncée pour le modèle, ce qui tronquait silencieusement les diffs volumineux ;
- disponibilité : `check_ollama_reachable()` permet de vérifier rapidement (quelques secondes) si le serveur répond, avant de lancer une génération qui pourrait sinon attendre le timeout complet ;
- simulation hors-ligne : si `config.MOCK_AI` est activé, `generate_commit_message` et `check_ollama_reachable` délèguent immédiatement au module `mock_llm` sans aucune requête réseau.

## Fonctions exposées

- `generate_commit_message(diff_files, feedback="")` : fonction principale, décrite ci-dessus.
- `check_ollama_reachable(timeout=3.0, fetch_models=False)` (alias `ping_ollama`) : fonction unique et centralisée de diagnostic Ollama. Vérifie la joignabilité (`/api/version`), mesure la latence réseau en millisecondes et extrait optionnellement les modèles installés (`/api/tags` si `fetch_models=True`). Retourne un objet `PingResult(reachable, error, latency_ms, installed_models)` déballable directement en tuple `(reachable, error)` ou sérialisable via `.to_dict()`. Utilisée de manière transverse par `action_generate_commit`, `action_release_notes` et `action_ping_ollama` (écran de configuration Go).

## Dépendances

- [config.md](../config.md)
- [models.md](../models.md)
- [mock_llm.md](mock_llm.md)
- [exceptions.md](exceptions.md)
- [commit_message_system.md](prompts/commit_message_system.md)
