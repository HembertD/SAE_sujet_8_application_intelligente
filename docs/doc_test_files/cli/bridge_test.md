# Tests Unitaires du Bridge Go

Fichier source associé : test/cli/bridge_test.go

## Description

Ce fichier contient la suite de tests automatisés validant le pont Go ↔ Python, la gestion du fichier `.env`, le basculement dynamique Mock/Réel, et l'isolation stricte des répertoires.

## Tests couverts

1. **`TestMockClientBasics`** : vérifie les réponses par défaut du `MockClient` (statut du dépôt, diff avec secrets masqués, message de commit, release notes).
2. **`TestMockClientOverrides`** : valide l'injection de réponses personnalisées et d'erreurs sur mesure pour tester la robustesse des vues.
3. **`TestParseBoolFromString`** : vérifie le parsing des booléens textuels issus du fichier `.env` (`true`, `false`, `1`, `0`, `yes`, `oui`, etc.).
4. **`TestEnvFileReadWrite`** : valide l'écriture et la relecture d'un fichier `.env` temporaire.
5. **`TestBridgeClientSwitching`** : vérifie la bascule dynamique entre `MockClient` et `PythonClient` lors de l'appel à `SaveConfig`.
6. **`TestPythonClientRealIntegration`** : valide l'exécution réelle du sous-processus Python (`--action status`, `diff`, `ping-ollama`).
7. **`TestAppEnvIsolationFromTargetRepo`** : garantit que lors de l'analyse d'un dépôt cible externe, le `.env` est créé/modifié **uniquement** dans l'application et **jamais** dans le dépôt cible.
8. **`TestAppEnvNeverInSrc`** : vérifie formellement que l'application refuse d'écrire dans `src/` et que le fichier `.env` réside obligatoirement à la racine du projet.

## Exécution des tests

```bash
cd test/cli
go test -v ./...
```
