# Client de Pontage Go ↔ Python (Bridge)

Fichier source associé : src/cli/bridge/backend_client.go

## Description

Ce module implémente le pont de communication asynchrone entre l'interface Go et le backend Python. Il garantit la règle d'architecture **Zéro logique en Go** en déléguant l'intégralité des opérations Git, de sécurité et d'appel réseau au sous-processus Python.

## Responsabilités

- exécuter les commandes en sous-processus via `exec.Command` (`python3 -m git_commit_release_notes_generator.service.core`) ;
- capturer la sortie standard `stdout` pour la désérialisation JSON stricte via `encoding/json` ;
- isoler les logs et messages d'erreur émis sur `stderr` pour ne pas corrompre le flux de données ;
- exposer une API Go haut niveau et typée (`GetRepoStatus`, `GetDiff`, `StageAll`, `GenerateCommit`, `ApplyCommit`, `Push`, `GetReleaseNotes`, `GetConfig`, `SaveConfig`, `PingOllama`) ;
- fournir un mode simulation / démo automatique (`UseMockMode`) basé sur les situations de test de l'équipe lorsque le backend Python est en cours de développement ou indisponible.

## Points clés

- **Découplage strict** : le front-end n'a aucune connaissance des bibliothèques Python utilisées (`GitPython`, `requests`, `urllib`) ;
- **Résilience** : détection automatique de la disponibilité du backend au démarrage ;
- **Mode Démo intégré** : simule fidèlement les délais d'inférence LLM (1.4s), le masquage des secrets et les retours d'actions pour permettre des tests d'interface autonomes.

## Rôle dans l’architecture

Passerelle unique d'accès au domaine métier. Toute vue front-end interagit avec le système exclusivement par l'intermédiaire de cette structure.
