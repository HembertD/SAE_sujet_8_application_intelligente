# Description d'image par un modèle de vision

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/vision.py

## Description

Ce module décrit le contenu d'une image ajoutée/supprimée, ou le changement visuel entre deux versions d'une image modifiée, via le modèle de vision Ollama (`OLLAMA_MODEL`, qui annonce la capacité `vision` dans `/api/tags`). Il permet aux fichiers image de contribuer à la génération d'un message de commit, au lieu d'être simplement ignorés comme n'importe quel autre binaire.

## Périmètre

Ce module ne lit ni n'écrit aucun fichier (les octets d'image lui sont fournis par l'appelant, `GitWrapper.get_staged_image_changes()`, module B) et n'écrit rien sur `stdout`. Il ne décide pas non plus **quand** il doit être appelé : `action_generate_commit()` (`service/core.py`) choisit de l'invoquer uniquement au moment de générer un message de commit, jamais lors d'une simple extraction de diff (`action_diff`/`get_staged_diff()`, qui doivent rester rapides et sans réseau).

## Flux de travail

1. Rejeter immédiatement un appel sans aucune image (`ValueError`), sans appel réseau.
2. Choisir, selon `status`, les images à envoyer et le message d'intention :
   - `"A"` (ajout) : seule la nouvelle image, décrite comme un nouveau contenu ;
   - `"D"` (suppression) : seule l'ancienne image, décrite comme un contenu retiré ;
   - sinon (modification/renommage) : les deux images (avant/après) envoyées ensemble, pour une comparaison explicite.
3. Encoder les images en base64 et les joindre au message utilisateur (champ `images` d'Ollama `/api/chat`, pris en charge tel quel par `llm.call_chat_json()` sans modification de son socle).
4. Appeler Ollama via `llm.call_chat_json()` avec un schéma JSON dédié.
5. Vérifier que la réponse est bien un JSON conforme et que la description n'est pas vide.
6. En cas d'erreur (validation ou contenu non-JSON), relancer une seconde fois avec l'erreur en contexte.
7. Retourner la description validée (ou lever `ImageDescriptionError` après le second échec).

## Points clés

- format : le modèle répond avec un schéma JSON strict (`description`) ;
- validation : le code rejette toute description vide ;
- robustesse : même mécanisme de retry unique que les autres générateurs du module (validation métier ET contenu non-JSON) ;
- réutilisation : s'appuie sur `llm.call_chat_json()` (socle HTTP partagé) — le champ `images` d'un message passe directement à travers, aucune modification du socle n'a été nécessaire ;
- simulation hors-ligne : si `config.MOCK_AI` est activé, délègue à `mock_llm.mock_describe_image()` sans requête réseau ;
- vérifié en conditions réelles contre le serveur Ollama partagé (`qwen3-vl:8b-instruct`) : une comparaison de deux images (fond blanc → fond gris) a produit une description correcte du changement.

## Fonctions exposées

- `describe_image_change(file_path, status, new_image, old_image=None) -> str` : fonction principale, décrite ci-dessus.

## Dépendances

- [config.md](../config.md)
- [mock_llm.md](mock_llm.md)
- [exceptions.md](exceptions.md)
- [llm.md](llm.md)
- [image_description_system.md](prompts/image_description_system.md)

## Détails déplacés du code source

- `new_image`/`old_image` : octets bruts (`bytes`), produits par `GitWrapper.get_staged_image_changes()` ; `None` pour le côté absent (ajout ou suppression).
- `describe_image_change()` : un seul retry est tenté ; la réponse invalide est réinjectée comme message `assistant` (JSON sérialisé, ou contenu brut si la réponse n'était pas du JSON), suivie d'un message `user` signalant l'erreur et exigeant une description conforme.
- `_validate(raw)` : retourne `None` si valide, sinon un message explicite (`description` manquante/vide).
