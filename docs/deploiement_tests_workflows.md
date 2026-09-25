# Tests automatisés via GitHub Actions

Cette page décrit comment les tests du projet sont lancés automatiquement par les workflows GitHub Actions.

## Objectif

Le projet utilise l’intégration continue pour vérifier automatiquement la qualité du code à chaque pull request et pour exécuter les tests importants dans un environnement standardisé.

## Workflows disponibles

Deux workflows principaux sont présents dans le dossier `.github/workflows/` :

- `ci.yml` : exécution des tests automatisés
- `documentation.yml` : build et publication de la documentation

## Workflow CI

Le fichier `.github/workflows/ci.yml` est déclenché sur les pull requests ciblant les branches `main` et `Dev`.

### Jobs

#### `python-tests`

Ce job a été retirer car les tests pythons ont besoin de la connection à l'IUT pour questionner l'IA local à cet connection, il est donc impossible que le workflows vérifie les tests

#### `go-tests`

Ce job exécute les tests Go du projet :

1. configuration de Go à partir de `src/cli/go.mod`
2. exécution de :

```bash
go test ./...
```

Il est lancé depuis le dossier `src/cli` pour vérifier le sous-projet Go.

## Vérification locale

Pour lancer les tests rapides localement :

```bash
python -m pytest -m "not slow" -q
```

Pour lancer les tests lents :

```bash
python -m pytest -m slow -q
```

Pour lancer les tests Go :

```bash
cd src/cli
go test ./...
```

## Intérêt du workflow CI

Le workflow CI sert à :

- détecter les régressions rapidement
- vérifier les modifications avant fusion
- garantir une certaine qualité de code sur les branches principales
- conserver une exécution des tests lourds dans un cadre contrôlé
