# Résolution de conflits de merge

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/conflict_resolver.py

## Description

Ce module propose une résolution pour une section de conflit de merge Git, via Ollama. Il construit le prompt à partir des deux côtés du conflit, envoie une requête structurée à l'API de chat, valide la réponse JSON et la transforme en objet métier.

## Périmètre

Ce module ne lit ni n'écrit aucun fichier, et ne détecte pas les conflits sur le disque (rôle du `GitWrapper`, module B). Il prend en entrée le texte déjà extrait des deux côtés du conflit (`ConflictSection`), et sa sortie (`ConflictResolution`) n'est **jamais appliquée automatiquement** : c'est toujours à l'utilisateur de valider avant toute écriture sur disque, comme pour le reste de l'application.

## Flux de travail

1. Rejeter immédiatement un conflit dont les deux côtés sont vides (`ValueError`), sans appel réseau.
2. Construire le prompt système et le résumé du conflit (fichier, base si disponible, ours, theirs).
3. Appeler Ollama via le socle partagé `llm.call_chat_json()` avec un schéma JSON dédié.
4. Vérifier que la réponse est bien un JSON conforme et qu'elle ne contient plus de marqueur de conflit (`<<<<<<<`, `=======`, `>>>>>>>`).
5. En cas d'erreur (validation ou contenu non-JSON), relancer une seconde fois avec l'erreur en contexte.
6. Convertir la réponse validée en objet `ConflictResolution`.

## Points clés

- format : le modèle répond avec un schéma JSON strict (`resolved_code`, `explanation`) ;
- validation : le code rejette toute proposition vide ou qui contient encore des marqueurs de conflit non résolus ;
- robustesse : même mécanisme de retry unique que `generate_commit_message()` (validation métier ET contenu non-JSON) ;
- réutilisation : s'appuie sur `llm.call_chat_json()` (socle HTTP partagé) plutôt que de dupliquer un troisième appel Ollama indépendant ;
- simulation hors-ligne : si `config.MOCK_AI` est activé, délègue à `mock_llm.mock_conflict_resolution()` sans requête réseau.

## Fonctions exposées

- `suggest_conflict_resolution(conflict: ConflictSection) -> ConflictResolution` : fonction principale, décrite ci-dessus.

## Dépendances

- [config.md](../config.md)
- [models.md](../models.md)
- [mock_llm.md](mock_llm.md)
- [exceptions.md](exceptions.md)
- [llm.md](llm.md)
- [conflict_resolution_system.md](prompts/conflict_resolution_system.md)

## Détails déplacés du code source

- `ConflictSection`/`ConflictResolution` : définies dans `models.py` comme contrat partagé (voir sa doc miroir), pas dans ce fichier.
- `suggest_conflict_resolution()` : un seul retry est tenté ; la réponse invalide est réinjectée comme message `assistant` (JSON sérialisé, ou contenu brut si la réponse n'était pas du JSON), suivie d'un message `user` signalant l'erreur et exigeant une proposition conforme sans marqueur de conflit restant.
- `_build_conflict_prompt()` : assemble le résumé textuel du conflit (fichier, base optionnelle, ours, theirs).
- `_validate(raw)` : retourne `None` si valide, sinon un message explicite (`resolved_code` manquant/vide, marqueur de conflit résiduel, `explanation` manquante/vide).
