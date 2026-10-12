# Tests du Mock IA (Simulation Hors-Ligne)

Fichier source associé : `test/test_mock_ai.py`

## Description

Ce fichier contient la suite de tests automatisés validant le module de simulation de l'intelligence artificielle (`mock_llm.py`), l'interprétation des variables de configuration d'environnement, et la bascule dynamique via la variable `MOCK_AI`.

## Ce qui est testé

1. **`test_parse_bool`** : vérifie la robustesse du parsing des valeurs booléennes issues de l'environnement (`true`, `TRUE`, `1`, `yes`, `oui`, `false`, `0`, etc.).
2. **`test_mock_generate_commit_empty_diff_raises`** : valide qu'une exception `ValueError` est levée lorsqu'aucun diff n'est fourni.
3. **`test_mock_generate_commit_static_message`** : vérifie la conformité Conventional Commits du message statique généré (`feat(mock): commit mocker pour test hors ligne`).
4. **`test_mock_generate_commit_with_user_feedback`** : contrôle l'injection de la consigne utilisateur dans le corps du message lors d'une révision.
5. **`test_mock_call_chat`** : vérifie la structure de la réponse JSON simulée de l'API de chat Ollama.
6. **`test_mock_release_notes`** : confirme la génération d'une note de version Markdown statique déterministe.
7. **`test_mock_ping`** : contrôle la simulation d'un ping réussi vers Ollama avec latence factice et liste de modèles installés.
8. **`test_llm_check_ollama_reachable_when_mock_ai`** : valide que le ping rapide répond immédiatement sans requête HTTP lorsque `MOCK_AI=True`.
9. **`test_llm_generate_commit_delegates_to_mock_when_mock_ai`** : valide l'aiguillage automatique de `generate_commit_message` vers le mock.
10. **`test_core_action_generate_commit_with_mock_ai`** : valide le fonctionnement de l'action `generate-commit` du service core sous mode mock.
11. **`test_core_action_release_notes_with_mock_ai`** : valide l'action `release-notes` du service core sous mode mock.
12. **`test_core_action_ping_ollama_with_mock_ai`** : valide l'action `ping-ollama` du service core sous mode mock.

## Stratégie de test

Les tests s'exécutent entièrement hors-ligne sans nécessiter de serveur Ollama ni d'accès réseau. Ils s'appuient sur `unittest.mock.patch` pour isoler `config.MOCK_AI` et valider les deux branches d'exécution.

## Exécution des tests

```bash
pytest test/test_mock_ai.py
```
