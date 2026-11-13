# Tests Unitaires du Bridge Go

Fichier source associé : test/cli/bridge_test.go

## Description

Ce fichier contient la suite de tests automatisés validant le pont Go ↔ Python, la gestion du fichier `.env`, le basculement dynamique Mock/Réel, et l'isolation stricte des répertoires.

## Tests couverts

1. **`TestMockClientBasics`** : vérifie les réponses par défaut du `MockClient` (statut du dépôt, diff avec secrets masqués, message de commit, prise en compte des types imposés et détection d'intentions contradictoires ou sans mot-clé, release notes).
2. **`TestMockClientOverrides`** : valide l'injection de réponses personnalisées et d'erreurs sur mesure pour tester la robustesse des vues.
3. **`TestParseBoolFromString`** : vérifie le parsing des booléens textuels issus du fichier `.env` (`true`, `false`, `1`, `0`, `yes`, `oui`, etc.).
4. **`TestEnvFileReadWrite`** : valide l'écriture et la relecture d'un fichier `.env` temporaire.
5. **`TestBridgeClientSwitching`** : vérifie la bascule dynamique entre `MockClient` et `PythonClient` lors de l'appel à `SaveConfig`.
6. **`TestPythonClientRealIntegration`** : valide l'exécution réelle du sous-processus Python (`--action status`, `diff`, `ping-ollama`).
7. **`TestAppEnvIsolationFromTargetRepo`** : garantit que lors de l'analyse d'un dépôt cible externe, le `.env` est créé/modifié **uniquement** dans l'application et **jamais** dans le dépôt cible.
8. **`TestAppEnvNeverInSrc`** : vérifie formellement que l'application refuse d'écrire dans `src/` et que le fichier `.env` réside obligatoirement à la racine du projet.
9. **`TestSaveConfigBothMocks`** : valide l'enregistrement conjoint et la relecture distincte des réglages `MOCK_INTERFACE` et `MOCK_AI` dans le fichier `.env`.
10. **`TestResolveForcedType`** : valide la résolution des 9 types Conventional Commits par index numérique (1 à 9), par nom exact insensible à la casse et le rejet des saisies invalides.
11. **`TestBuildForcedTypeFeedback`** : valide la construction formatée du feedback (`type: <type>` ou `type: <type>. <consigne>`) conforme à l'extraction d'intention du backend.
12. **`TestReadWritePythonBinSetting`** : valide l'écriture de `PYTHON_BIN`, sa relecture dans `ConfigResponse`, et sa réinitialisation vers la détection automatique.
13. **`TestResolvePythonBinPriority`** : valide l'ordre de priorité strict de résolution (`PYTHON_BIN` surcharge `.venv`, et échec explicite sans repli silencieux si aucun environnement n'est configuré).
14. **`TestVerifyPythonEnvironmentReal`** : valide l'exécution de la sonde complète sur l'environnement applicatif réel (présence des dépendances, absence d'erreur d'import et succès global).
15. **`TestVerifyPythonEnvironmentInvalid`** : vérifie la détection d'erreur lorsqu'aucun environnement n'est configuré ou lorsqu'un binaire Python plante.
16. **`TestVerifyPythonEnvironmentMissingDep`** : vérifie la détection explicite d'une dépendance manquante (`missing_deps`) et le blocage immédiat par `CheckPythonEnvironment`.
17. **`TestVerifyPythonCandidate`** : vérifie la validation stricte d'un candidat avant modification de la configuration (rejet des binaires introuvables, des dépendances manquantes et validation des environnements intègres).

## Exécution des tests

```bash
cd test/cli
go test -v ./...
```
