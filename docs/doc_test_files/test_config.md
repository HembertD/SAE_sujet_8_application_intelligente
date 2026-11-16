# Tests de la configuration Python

Fichier source associé : test/test_config.py

## Description

Ce module vérifie la résolution de la configuration Ollama côté Python, en
particulier l'articulation entre le serveur de référence de l'IUT et le serveur
local lancé par Docker.

## Ce qui est testé

- `USE_DOCKER_LLM=true` force `OLLAMA_BASE_URL` sur `http://localhost:11434`,
  même si le `.env` contient l'URL du serveur IUT ;
- `USE_DOCKER_LLM=false` conserve l'URL du serveur IUT du `.env` ;
- le modèle par défaut reste `gemma4:26b`, modèle de référence du projet et du
  serveur IUT.

## Stratégie de test

Le module `config` lit le `.env` à l'import. Les tests créent un `.env`
temporaire, positionnent `GITELLIGENCE_APP_ROOT` dessus, puis rechargent le
module avec `importlib.reload`. La configuration d'origine est restaurée à la
fin de chaque test.

## Points de vigilance

- Le mode Docker ne doit pas modifier les valeurs par défaut du projet : il ne
  s'active que lorsque `USE_DOCKER_LLM=true`.
- Le backend Python doit appliquer la bascule même lorsqu'il est exécuté seul,
  sans passer par le CLI Go.
