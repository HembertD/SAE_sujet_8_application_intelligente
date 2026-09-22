# Test réseau réel du client Ollama

Fichier source associé : test/test_ollama_llm_slow.py

## Description

Ce test vérifie que le client Ollama fonctionne réellement contre le serveur partagé, et pas seulement contre une réponse simulée. Il est marqué `slow` et n'est jamais exécuté par défaut.

## Ce qui est testé

- un appel complet `generate_commit_message()` sur un diff réaliste (ajout d'une fonction de connexion) ;
- que le serveur Ollama répond, que la réponse est bien un JSON conforme, et que le `CommitMessage` obtenu a un `type` et un `subject` non vides.

## Stratégie de test

Aucun mock : le test appelle vraiment `http://10.22.28.190:11434`. Il nécessite donc un accès réseau au serveur partagé et peut prendre 15 à 45 secondes selon que le modèle est déjà chargé en mémoire côté serveur.

## Cas importants

### génération contre le vrai modèle

Le test affiche (`-s`) le `type`, le `scope`, le `subject` et le `body` obtenus, pour une inspection visuelle rapide en plus des assertions automatiques.

## Points de vigilance

- Ce test doit rester exclu de l'exécution par défaut (`pytest.ini` filtre déjà `not slow`) pour ne pas ralentir ni faire échouer la suite rapide en cas de coupure réseau.
- Une latence de 15 à 45 secondes est normale, pas un signe de bug.
