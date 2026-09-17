# Architecture de documentation du projet
## Principe

La Documentation technique est placée dans :

- docs/doc_src_files

C’est un miroir de :

- src

La règle est simple :

- un fichier Python dans src a son équivalent .md dans docs/doc_src_files ;
- les noms de fichiers sont identiques, avec l’extension .md ;
- la structure des dossiers est conservée autant que possible.

## Convention de nommage

Exemples :

- src/git_commit_release_notes_generator/config.py
- docs/doc_src_files/git_commit_release_notes_generator/config.md

- src/git_generator/service/git_wrapper.py
- docs/doc_src_files/git_generator/service/git_wrapper.md

## Arborescence actuelle

```text
.
├─ README.md
├─ contributing.md
├─ doc_architecture.md
├─ pytest.ini
├─ .githooks/
│  └─ prepare-commit-msg
├─ src/
│  ├─ git_commit_release_notes_generator/
│  │  ├─ __init__.py
│  │  ├─ config.py
│  │  ├─ models.py
│  │  ├─ ollama_client/
│  │  │  ├─ __init__.py
│  │  │  ├─ base.py
│  │  │  ├─ embedding.py
│  │  │  ├─ exceptions.py
│  │  │  ├─ llm.py
│  │  │  └─ prompts/
│  │  │     └─ commit_message_system.txt
│  │  └─ service/
│  │     ├─ __init__.py
│  │     ├─ core.py
│  │     └─ utils.py
│  └─ git_generator/
│     ├─ __init__.py
│     └─ service/
│        ├─ __init__.py
│        └─ git_wrapper.py
├─ docs/
│  ├─ monitoring/
│  │  └─ sprint-00.md
│  └─ doc_src_files/
│     ├─ README.md
│     ├─ git_commit_release_notes_generator/
│     │  ├─ __init__.md
│     │  ├─ config.md
│     │  ├─ models.md
│     │  ├─ ollama_client/
│     │  │  ├─ __init__.md
│     │  │  ├─ base.md
│     │  │  ├─ embedding.md
│     │  │  ├─ exceptions.md
│     │  │  ├─ llm.md
│     │  │  └─ prompts/
│     │  │     └─ commit_message_system.md
│     │  └─ service/
│     │     ├─ __init__.md
│     │     ├─ core.md
│     │     └─ utils.md
│     └─ git_generator/
│        ├─ __init__.md
│        └─ service/
│           ├─ __init__.md
│           └─ git_wrapper.md
├─ test/
│  ├─ ...
└─ scripts/
   └─ run_metrics.py
```

## Rôle de chaque dossier

### src/
Le code applicatif du projet.

### docs/doc_src_files/
La documentation technique détaillant les fichiers de src, avec le même arbre que le code et des fichiers .md.

### docs/monitoring/
La gestion des sprints, de l’avancement et du suivi projet.

## Règles de maintenance

1. Ajouter les fichiers .md dans docs/doc_src_files en gardant le même chemin que src.
2. Ne pas mélanger la documentation technique avec la gestion de sprint.
3. Mettre à jour la doc lorsqu’un module est ajouté, renommé ou refactoré.
4. Conserver la séparation claire entre :
   - architecture technique : docs/doc_src_files
   - planification / sprints : docs/monitoring

## Pourquoi cette organisation ?

Cette structure rend le projet plus lisible et extensible :

- il est possible de retrouver rapidement le code et sa documentation correspondante ;
- les sprints restent séparés de l’architecture technique ;
- l’arborescence peut évoluer sans casser la logique documentaire.
