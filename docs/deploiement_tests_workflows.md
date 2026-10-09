# Tests automatisés via GitHub Actions

Cette page décrit comment les tests du projet sont lancés automatiquement par les workflows GitHub Actions.

## Objectif

Le projet utilise l’intégration continue pour vérifier automatiquement la qualité du code à chaque pull request et pour exécuter les tests importants dans un environnement standardisé.

## Workflows disponibles

Un seul workflow est présent dans le dossier `.github/workflows/` :

- `ci.yml` : tests Go, tests Python et build/déploiement de la documentation (jobs `go-tests`, `python-tests`, `documentation-build`, `deploy-docs`)

## Workflow CI

Le fichier `.github/workflows/ci.yml` est déclenché sur les `push` (toutes branches), les `pull_request` ciblant `main` et `dev`, et manuellement via `workflow_dispatch`.

### Jobs

#### `python-tests`

Ce job exécute la suite de tests Python (module C) :

1. installation des dépendances depuis `requirements.txt` (dont `pytest`)
2. exécution de :

```bash
python -m pytest -q
```

`pytest.ini` exclut par défaut les tests marqués `slow` (ceux qui appellent le vrai serveur Ollama de l'IUT), via `addopts = -m "not slow"`. Le job CI ne lance donc que les tests mockés, qui ne dépendent d'aucune connexion réseau — c'est précisément pour ça que la distinction `slow` / non-`slow` existe.

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
