# Implémentation Simulée MockClient

Fichier source associé : src/cli/bridge/mock_client.go

## Description

Ce module implémente l'interface `BackendClient` en fournissant des réponses synthétiques immédiates pour le mode démo hors-ligne et les tests automatisés du front-end.

## Responsabilités

- Simuler un dépôt Git actif avec des modifications réalistes (détection de secrets, fichiers binaires, ajouts/suppressions).
- Générer des messages de commit simulés respectant la convention Conventional Commits.
- Produire des Release Notes formatées en Markdown sans appel réseau.
- Simuler la connectivité et la latence avec un serveur Ollama.
- Permettre la surcharge (`CustomError`, `CustomDiff`, `SimulateDelay = false`) pour les tests unitaires Go.

## Points clés

- **Autonomie totale** : Permet de présenter l'application ou d'exécuter des tests d'interface sans nécessiter de serveur Ollama actif ni de backend Python installé.
- **Conformité aux règles de sécurité** : Démontre la détection des clés API sensibles et l'exclusion des fichiers binaires.

## Détails déplacés du code source

Précisions issues des commentaires retirés de `mock_client.go` :

- `MockClient` fournit une implémentation de `BackendClient` pour les tests et le mode hors-ligne ; les champs `CustomRepoStatus`, `CustomDiff`, `CustomCommitProposal`, `CustomReleaseNotes`, `CustomPing` et `CustomError` permettent la surcharge pour les tests unitaires.
- `IsMock` retourne toujours `true` ; `SimulateDelay` (via `sleep`) simule les latences réseau.
- Données types : branche `interface-cli` avec 3 fichiers indexés et 1 non indexé ; `src/auth/token_vault.py` avec données sensibles masquées (`HasSecretsMasked`) ; `assets/banner.png` comme fichier binaire exclu.
- `GenerateCommit` renvoie une proposition Conventional Commits adaptée au `feedback` éventuel ; `ApplyCommit` applique le commit sur la branche courante ; `Push` envoie vers la branche distante ; `GetReleaseNotes` synthétise le Markdown entre deux tags ; `PingOllama` vérifie la joignabilité d'Ollama ; `SaveMarkdownFile` enregistre localement le Markdown exporté ; `GetConfig` / `SaveConfig` lisent et mettent à jour le `.env` applicatif.

## Commentaires de fonction (dupliqués du code source)

- `NewMockClient()` : Instancie le client simulé avec la même racine pour dépôt et application.
- `NewMockClientWithAppRoot()` : Instancie le client simulé avec racines distinctes.
- `MockClient.sleep()` : Simule une latence réseau lorsque activée.
- `MockClient.IsMock()` : Indique que le client est toujours en mode simulation.
- `MockClient.GetRepoStatus()` : Retourne un statut de dépôt simulé.
- `MockClient.GetDiff()` : Retourne un diff simulé avec secrets et binaires.
- `MockClient.StageAll()` : Simule l'indexation de tous les fichiers.
- `MockClient.GenerateCommit()` : Génère une proposition de commit simulée adaptée au retour (avec détection déterministe d'un type forcé via `extractForcedType` si une intention explicite est formulée).
- `MockClient.ApplyCommit()` : Simule l'application du commit sur la branche courante.
- `MockClient.Push()` : Simule la poussée vers la branche distante.
- `MockClient.GetReleaseNotes()` : Génère des notes de version simulées entre deux tags.
- `MockClient.GetConfig()` : Charge la configuration simulée depuis l'environnement.
- `MockClient.SaveConfig()` : Persiste la configuration simulée dans le .env applicatif.
- `MockClient.PingOllama()` : Retourne un diagnostic Ollama simulé.
- `MockClient.SaveMarkdownFile()` : Sauvegarde le contenu Markdown dans le dépôt simulé.
- `extractForcedType()` : Détecte un type forcé dans la consigne de feedback via regex d'intention explicite (aligné sur `_extract_forced_type` de `llm.py`).

