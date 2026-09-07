Architecture proposé en énoncé

```
src/
│
├─ <subject>/ # ex. v_crop_rag, rom_rag, tutor_math, …
│ ├─ service/ # logique métier pure
│ │ ├─ __init__.py
│ │ ├─ core.py # fonctions exposées au UI
│ │ └─ utils.py
│ │
│ ├─ storage/ # accès aux DB / vector stores
│ │ ├─ __init__.py
│ │ └─ base.py
│ │
│ ├─ ollama_client/ # wrapper LLM/VLM/Embedding
│ │ ├─ __init__.py
│ │ ├─ base.py # interface Python
│ │ ├─ llm.py # implémentation pour LLM
│ │ ├─ vlm.py # implémentation pour VLM
│ │ └─ embedding.py
│ │
│ ├─ ui_gradio.py # point d’entrée Gradio → service
│ └─ config.py # pydantic-settings (env vars)
│
└─ shared/ # code partagé entre plusieurs sujets
└─ logging.py
docker/
└─ Dockerfile
docker-compose.yml
```