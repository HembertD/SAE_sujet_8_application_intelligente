Architecture à adopter (à modifier pour ajouter les fichiers précis) (à maintenir à jour le plus possible)

```
src/
│
├─ Git Commit & Release Notes Generator
│ ├─ service/ # logique métier pure
│ │ ├─ __init__.py
│ │ ├─ core.py # fonctions exposées au UI
│ │ └─ utils.py
│ │
│ ├─ ollama_client/ # wrapper LLM/Embedding
│ │ ├─ __init__.py
│ │ ├─ base.py # interface Python
│ │ ├─ llm.py # implémentation pour LLM
│ │ └─ embedding.py
│ │
│ └─ config.py # pydantic-settings (env vars)
|
└─ logging.py

.githooks/
└── prepare-commit-msg

test/
|
|- test # permet de tester l'application
```