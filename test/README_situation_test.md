# Situations de test pour l'application intelligente

Ce dossier regroupe un ensemble de **situations de test** réalistes pour l'application intelligente. L'objectif est de mettre l'application en situation face à différents contextes de modifications de code (`git diff`) afin d'évaluer et de valider ses fonctionnalités :
- **Analyse de code** : détection et extraction des modifications locales non commitées.
- **Classification et génération** : proposition automatique d'un message respectant la norme **Conventional Commits** (`feat`, `fix`, `docs`, etc.).
- **Sécurité et filtrage** : exclusion des données sensibles (fichiers `.env`, clés d'API).
- **Synthèse de version** : extraction de l'historique entre deux tags Git pour générer des release notes.

Pour cela, un faux projet de base (`fake_project/` : calculatrice Python dotée d'un historique Git local) a été décliné dans le sous-dossier `commit/` en plusieurs situations de test, chacune présentant des modifications locales non encore commitées.

## Activation et désactivation des dossiers de test (* / *_test)

Pour éviter que Git, GitHub ou des scanners d'environnement n'interfèrent avec ces dossiers de test (sous-modules, règles d'exclusion, workflows GitHub Actions, secrets .env), **tous les éléments sensibles et liés à Git sont neutralisés en de simples fichiers/dossiers avec le suffixe `_test`** par défaut :
- `.git` ➔ `.git_test`
- `.gitignore` ➔ `.gitignore_test`
- `.github` ➔ `.github_test`
- `.env` ➔ `.env_test`

- **Activer les situations de test** (`*_test` ➔ actifs) :
  ```bash
  python3 situation_test/init_test.py
  ```
  *(ou importable en Python : `from situation_test.init_test import init_test; init_test()`)*

- **Désactiver / Masquer les dossiers** (actifs ➔ `*_test`) :
  ```bash
  python3 situation_test/deinit_test.py
  ```
  *(ou importable en Python : `from situation_test.deinit_test import deinit_test; deinit_test()`)*

## Liste des situations de test

| Dossier | Type | Modifications non commitées (`git diff`) | Résultat attendu pour le LLM |
| :--- | :--- | :--- | :--- |
| `fake_project/` | *Base* | Dépôt propre avec tags `v0.1.0` et `v1.0.0` | Référence de base, release notes |
| `commit/commit_feat/` | **feat** | Ajout des fonctions `power(a, b)` et `modulo(a, b)` | `feat: ajout des opérations puissance et modulo` |
| `commit/commit_fix/` | **fix** | Correction d'arrondi flottant et gestion 0/0 | `fix: correction des imprécisions flottantes et division 0/0` |
| `commit/commit_docs/` | **docs** | Enrichissement du `README.md` (guide de contribution, gestion d'erreurs) | `docs: mise à jour de la documentation et guide d'utilisation` |
| `commit/commit_style/` | **style** | Reformatage PEP 8, typage et docstrings sans changement logique | `style: formatage du code selon les conventions PEP 8` |
| `commit/commit_refactor/` | **refactor** | Encapsulation de la calculatrice dans une classe `Calculator` | `refactor: encapsulation de la calculatrice dans une classe Calculator` |
| `commit/commit_test/` | **test** | Ajout de 4 nouvelles méthodes de tests unitaires dans `test_calculator.py` | `test: ajout de tests unitaires pour les cas limites` |
| `commit/commit_chore/` | **chore** | Ajout d'un `Makefile` pour automatiser les tâches (`test`, `clean`) | `chore: ajout du Makefile pour les tâches de maintenance` |
| `commit/commit_perf/` | **perf** | Mise en cache avec `@lru_cache` et table constante d'opérations | `perf: optimisation des calculs via mise en cache lru_cache` |
| `commit/commit_ci/` | **ci** | Matrice multi-versions Python et étape `flake8` dans `.github/workflows/ci.yml` | `ci: ajout de la matrice multi-versions et linting dans GitHub Actions` |
| `commit/commit_build/` | **build** | Ajout de `pyproject.toml` et mise à jour de `requirements.txt` | `build: configuration de build pyproject.toml et dépendances` |
| `commit/commit_revert/` | **revert** | Annulation des modifications introduites par le commit précédent | `revert: annulation du commit de la fonction inverse` |
| `commit/commit_sensitive/` | **security** | Ajout d'un fichier `.env` avec clés API et token dans le code | Détection et filtrage de sécurité des données sensibles |
