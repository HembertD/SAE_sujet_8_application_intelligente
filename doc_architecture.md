# Architecture de documentation du projet
## Principe

La Documentation technique est placée dans :

- docs/doc_src_files : documentation des fichiers de src
- docs/doc_test_files : documentation des fichiers de test

Les deux dossiers sont des miroirs de :

- src
- test

La règle est simple :

- un fichier Python dans src a son équivalent .md dans docs/doc_src_files ;
- un fichier Python dans test a son équivalent .md dans docs/doc_test_files ;
- les noms de fichiers sont identiques, avec l’extension .md ;
- la structure des dossiers est conservée autant que possible.

## Convention de nommage

Exemples :

- src/git_commit_release_notes_generator/config.py
- docs/doc_src_files/git_commit_release_notes_generator/config.md

- src/cli/bridge/env.go
- docs/doc_src_files/cli/bridge/env.md

- src/git_commit_release_notes_generator/service/git_wrapper.py
- docs/doc_src_files/git_commit_release_notes_generator/service/git_wrapper.md

- test/cli/bridge_test.go
- docs/doc_test_files/cli/bridge_test.md

## Arborescence actuelle

```text
.
├─ .env.exemple
├─ .gitignore
├─ pytest.ini
├─ requirements.txt
├─ README.md
├─ contributing.md
├─ doc_architecture.md
├─ .githooks/
│  └─ prepare-commit-msg
├─ src/
│  ├─ cli/
│  │  ├─ go.mod
│  │  ├─ main.go
│  │  ├─ bridge/
│  │  │  ├─ client.go
│  │  │  ├─ python_client.go
│  │  │  ├─ mock_client.go
│  │  │  └─ env.go
│  │  ├─ models/
│  │  │  └─ types.go
│  │  ├─ ui/
│  │  │  ├─ box.go
│  │  │  ├─ spinner.go
│  │  │  └─ styles.go
│  │  └─ views/
│  │     ├─ commit_view.go
│  │     ├─ config_view.go
│  │     ├─ menu_view.go
│  │     └─ release_notes_view.go
│  └─ git_commit_release_notes_generator/
│     ├─ __init__.py
│     ├─ config.py
│     ├─ models.py
│     ├─ ollama_client/
│     │  ├─ __init__.py
│     │  ├─ base.py
│     │  ├─ embedding.py
│     │  ├─ exceptions.py
│     │  ├─ llm.py
│     │  └─ prompts/
│     │     └─ commit_message_system.txt
│     └─ service/
│        ├─ __init__.py
│        ├─ core.py
│        ├─ git_wrapper.py
│        └─ utils.py
├─ docs/
│  ├─ explication_go_cli.md
│  ├─ monitoring/
│  │  ├─ sprint-00.md
│  │  └─ sprint-01.md
│  ├─ doc_scripts_files/
│  │  └─ run_metrics.md
│  ├─ doc_src_files/
│  │  ├─ cli/
│  │  │  ├─ go_mod.md
│  │  │  ├─ main.md
│  │  │  ├─ bridge/
│  │  │  │  ├─ backend_client.md
│  │  │  │  ├─ client.md
│  │  │  │  ├─ python_client.md
│  │  │  │  ├─ mock_client.md
│  │  │  │  └─ env.md
│  │  │  ├─ models/
│  │  │  │  └─ types.md
│  │  │  ├─ ui/
│  │  │  │  ├─ box.md
│  │  │  │  ├─ spinner.md
│  │  │  │  └─ styles.md
│  │  │  └─ views/
│  │  │     ├─ commit_view.md
│  │  │     ├─ config_view.md
│  │  │     ├─ menu_view.md
│  │  │     └─ release_notes_view.md
│  │  └─ git_commit_release_notes_generator/
│  │     ├─ __init__.md
│  │     ├─ config.md
│  │     ├─ models.md
│  │     ├─ ollama_client/
│  │     │  ├─ __init__.md
│  │     │  ├─ base.md
│  │     │  ├─ embedding.md
│  │     │  ├─ exceptions.md
│  │     │  ├─ llm.md
│  │     │  └─ prompts/
│  │     │     └─ commit_message_system.md
│  │     └─ service/
│  │        ├─ __init__.md
│  │        ├─ core.md
│  │        ├─ git_wrapper.md
│  │        └─ utils.md
│  └─ doc_test_files/
│     ├─ cli/
│     │  ├─ go_mod.md
│     │  └─ bridge_test.md
│     ├─ test_core.md
│     ├─ test_git_wrapper.md
│     ├─ test_ollama_llm.md
│     ├─ test_golden_diffs_slow.md
│     ├─ test_integration_git_wrapper_slow.md
│     ├─ test_ollama_llm_slow.md
│     ├─ fixtures/
│     │  └─ golden_diffs.md
│     └─ situation_test/
│        ├─ init_test.md
│        └─ deinit_test.md
├─ test/
│  ├─ __init__.py
│  ├─ README_situation_test.md
│  ├─ cli/
│  │  ├─ go.mod
│  │  └─ bridge_test.go
│  ├─ test_core.py
│  ├─ test_git_wrapper.py
│  ├─ test_ollama_llm.py
│  ├─ test_golden_diffs_slow.py
│  ├─ test_integration_git_wrapper_slow.py
│  ├─ test_ollama_llm_slow.py
│  ├─ fixtures/
│  │  ├─ __init__.py
│  │  └─ golden_diffs.py
│  └─ situation_test/
│     ├─ init_test.py
│     ├─ deinit_test.py
│     ├─ fake_project/
│     └─ commit/
└─ scripts/
   └─ run_metrics.py
```

## Rôle de chaque dossier

### src/
Le code applicatif du projet (Python pour le backend IA et git wrapper, Go pour l'interface TUI/CLI).

### docs/doc_src_files/
La documentation technique détaillant les fichiers de src, avec le même arbre que le code et des fichiers .md.

### docs/doc_test_files/
La documentation technique détaillant les fichiers de test, avec le même arbre que les tests et des fichiers .md.

### docs/doc_scripts_files/
La documentation technique des scripts utilitaires (mesures de métriques, benchmarks).

### docs/monitoring/
La gestion des sprints, de l’avancement et du suivi projet.

## Règles de maintenance

1. Ajouter les fichiers .md dans docs/doc_src_files en gardant le même chemin que src.
2. Ajouter les fichiers .md dans docs/doc_test_files en gardant le même chemin que test.
3. Ne pas mélanger la documentation technique avec la gestion de sprint.
4. Mettre à jour la doc lorsqu’un module est ajouté, renommé ou refactoré.
5. Conserver la séparation claire entre :
   - architecture technique : docs/doc_src_files
   - tests et scénarios : docs/doc_test_files
   - planification / sprints : docs/monitoring

## Pourquoi cette organisation ?

Cette structure rend le projet plus lisible et extensible :

- il est possible de retrouver rapidement le code et sa documentation correspondante ;
- les tests disposent aussi d’une documentation explicite, sans mélange avec le code applicatif ;
- les sprints restent séparés de l’architecture technique ;
- l’arborescence peut évoluer sans casser la logique documentaire.
