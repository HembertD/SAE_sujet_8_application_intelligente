# Tests du wrapper Git

Fichier source associé : test/test_git_wrapper.py

## Description

Ce module valide le comportement du wrapper Git utilisé pour analyser les diffs et préparer les données envoyées au système de génération de commit.

## Ce qui est testé

- le parsing d’un diff Git en fichiers structurés ;
- la détection des modifications par fichier ;
- la suppression des données sensibles dans les patches ;
- la création d’un payload JSON exploitable par le modèle ou par l’orchestrateur applicatif.

## Cas de test principaux

### 1. Parsing d’un diff multi-fichiers

Le test vérifie que le diff est découpé correctement en plusieurs objets, avec pour chaque fichier :

- le chemin du fichier ;
- le statut de modification ;
- les lignes ajoutées et supprimées ;
- le patch complet associé.

### 2. Sanitisation des secrets

Le test simule un contenu contenant une clé ou une valeur sensible et vérifie que le mécanisme de redaction remplace bien les informations sensibles par une valeur de type `[REDACTED_SECRET]`.

### 3. Structure du payload JSON

Le test s’assure que les objets renvoyés par le wrapper sont convertibles en payloads normalisés, sans informations parasites ou inutiles. Il vérifie notamment la présence des champs attendus comme `path`, `status`, `patch` et `is_binary`.

## Points de vigilance

- La logique de parsing doit rester tolérante aux variations de format de diff Git.
- La redaction des secrets doit être systématique avant toute transmission hors du contexte local.
- Les payloads doivent rester structurés et légers pour éviter d’encombrer le modèle LLM.
