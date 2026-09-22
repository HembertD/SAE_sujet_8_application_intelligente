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

Ce job vérifie les tests Python rapides :

1. récupération du dépôt
2. installation de Python 3.12
3. installation des dépendances depuis `requirements.txt`
4. installation de `pytest`
5. exécution de :

```bash
python -m pytest -m "not slow" -q
```

Il ne lance que les tests non marqués comme `slow` pour garder la validation rapide.

#### `go-tests`

Ce job exécute les tests Go du projet :

1. configuration de Go à partir de `src/cli/go.mod`
2. exécution de :

```bash
go test ./...
```

Il est lancé depuis le dossier `src/cli` pour vérifier le sous-projet Go.

#### `slow-tests`

Ce job est conditionné par :

```yaml
if: github.event_name == 'workflow_dispatch'
```

Cela signifie qu’il ne s’exécute que lorsqu’un workflow est déclenché manuellement. Il lance les tests Python marqués `slow` :

```bash
python -m pytest -m slow -q
```

## Rôle des marqueurs de tests

Le projet utilise les marqueurs `slow` pour distinguer les tests lourds ou plus longs des validations rapides.

Cela permet :

- d’avoir une CI courte et réactive pour les PR
- de garder les tests longs hors du flux standard
- d’exécuter les validations lourdes à la demande

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
