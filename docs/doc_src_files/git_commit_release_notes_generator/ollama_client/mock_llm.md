# Module de Simulation de l'IA (Mock LLM)

Fichier source associé : `src/git_commit_release_notes_generator/ollama_client/mock_llm.py`

## Description

Ce module fournit une implémentation simulée (Mock) des fonctionnalités d'intelligence artificielle d'Ollama. Il permet de travailler sur l'application en mode hors-ligne sans serveur Ollama actif, tout en conservant l'analyse Git réelle, l'indexation et la persistance des fichiers.

## Responsabilités

- Générer un message de commit déterministe et statique respectant le format Conventional Commits (`feat(mock): commit mocker pour test hors ligne`) ;
- Supporter la prise en compte de consignes utilisateur (`feedback`) injectées dans le corps du message pour les tests de révision interactive ;
- Simuler la charge utile JSON d'un appel à l'API de chat Ollama (`mock_call_chat`) ;
- Générer une réponse fixe et prévisible pour les Release Notes (`mock_release_notes`) ;
- Simuler le ping de connectivité et la liste des modèles installés (`mock_ping`) avec une latence factice de 10 ms.

## Fonctions exposées

| Fonction | Arguments | Retour | Description |
|---|---|---|---|
| `mock_generate_commit_message` | `diff_files`, `feedback=""` | `CommitMessage` | Renvoie le commit statique mocké. Lève `ValueError` si la liste de diff est vide. |
| `mock_call_chat` | `messages` | `dict` | Simule la réponse brute de `/api/chat`. |
| `mock_release_notes` | `commits`, `from_tag`, `to_tag` | `str` | Renvoie une note de version Markdown statique déterministe. |
| `mock_ping` | — | `dict` | Renvoie un diagnostic de connectivité simulé réussi avec latence de 10 ms. |

## Points clés

- **Simplicité et prévisibilité** : Les réponses sont volontairement statiques et prévisibles afin de faciliter les tests unitaires et le développement hors-ligne.
- **Délégation automatique** : `llm.py` et `core.py` basculent automatiquement vers ce module dès que la variable `MOCK_AI` est activée dans la configuration.
