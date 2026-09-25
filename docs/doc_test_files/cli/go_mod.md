# Module Go de Test

Fichier source associé : test/cli/go.mod

## Description

Ce fichier définit le module Go secondaire (`sae-git-cli-test`) utilisé pour exécuter les tests unitaires et d'intégration du front-end Go depuis le dossier `test/`.

## Responsabilités

- Isoler l'environnement de test Go dans le dossier `test/` conformément à l'architecture du projet.
- Référencer le module principal `sae-git-cli` via la directive `replace sae-git-cli => ../../src/cli`.
