# Tests unitaires du point d'entrée Core

Fichier source associé : test/test_core.py

## Description

Ce fichier de test valide le comportement du module `git_commit_release_notes_generator.service.core`, garantissant la conformité avec le contrat d'échange JSON du front-end Go.

## Cas de tests couverts

1. **`test_action_status_on_valid_repo`** : vérifie la conformité de l'objet d'état sur un vrai dépôt Git (branche, compteurs staged / unstaged).
2. **`test_action_status_on_non_git_repo`** : vérifie la gestion propre des erreurs sur un dossier hors-Git (`is_git_repo = False`).
3. **`test_action_diff_structure`** : vérifie la structure de chaque fichier de diff (`path`, `status`, `binary`, `patch`, `has_secrets_masked`).
4. **`test_action_generate_commit_no_staged_files`** : vérifie le retour d'erreur lorsqu'aucun fichier n'est indexé.
5. **`test_action_generate_commit_success`** : simule une inférence LLM réussie avec serveur joignable et intégration du `feedback` utilisateur.
6. **`test_action_generate_commit_ollama_unreachable`** : vérifie l'échec immédiat et explicite si Ollama est injoignable (`check_ollama_reachable = False`), sans tenter l'appel de génération.
7. **`test_action_generate_commit_network_error_fails_fast`** : vérifie le comportement de bout en bout lorsqu'une erreur réseau urllib survient lors du ping.
8. **`test_action_apply_commit_validation`** : rejette les messages de commit vides.
9. **`test_action_apply_commit_success`** : simule l'application d'un commit Git et vérifie la présence du SHA.
10. **`test_action_push_no_remote`** : gère l'absence de remote Git configuré sans plantage.
11. **`test_action_release_notes_success`** : simule une génération LLM réussie (`generate_release_notes` patché) et vérifie le Markdown renvoyé.
12. **`test_action_release_notes_fails_honestly_when_llm_call_errors`** : vérifie qu'un échec de l'appel LLM (ping OK mais erreur réseau lors du chat) remonte comme `success=False` avec un message d'erreur explicite, **jamais** comme un succès silencieusement rempli par un formateur Regex local présenté comme une sortie du modèle (anomalie corrigée, cf. `bug_relase_note.md`).
13. **`test_action_release_notes_ollama_unreachable`** : vérifie que la génération des release notes échoue immédiatement si le serveur Ollama est injoignable au ping (`check_ollama_reachable = False`), à l'identique des commits.
14. **`test_action_release_notes_empty`** : vérifie la réponse lors d'un historique vide entre deux bornes identiques.
15. **`test_unknown_action_exits_zero_with_error_json`** : vérifie que le process sort avec le code 0 et renvoie une erreur JSON propre sur action inconnue.
16. **`test_action_save_config`** : vérifie la prise en compte de la notification de configuration.
17. **`test_action_ping_ollama_reachable`** : simule une réponse positive d'Ollama avec latence et modèles installés.
18. **`test_action_ping_ollama_unreachable`** : vérifie le retour propre en cas de défaillance réseau (`reachable = False`).
