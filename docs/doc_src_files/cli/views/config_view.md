# Vue Configuration

Fichier source associé : src/cli/views/config_view.go

## Description

Ce module permet de visualiser et de modifier les paramètres opérationnels de l'application (URL Ollama, modèle LLM, timeout, interpréteur Python) et de diagnostiquer la connectivité réseau avec le serveur de l'IUT ainsi que l'état de l'environnement Python.

## Responsabilités

- afficher les paramètres actuellement appliqués issus du fichier d'environnement `.env` situé à la racine du projet (URL, Modèle, Timeout, Interpréteur Python résolu avec indication du mode `.venv` ou `PYTHON_BIN`, ainsi que les statuts distincts `Mock Interface` et `Mock IA` exclusivement lorsque le mode `--dev` est actif) ;
- **Action [T] (Test Ping)** : exécuter un diagnostic réseau vers l'API Ollama (ou simulation instantanée si Mock IA actif), mesurer la latence en millisecondes et lister l'ensemble des modèles d'IA installés sur le serveur ;
- **Action [P] (Vérification Python)** : tester en profondeur l'environnement Python, valider la présence de toutes les dépendances requises (GitPython, PyYAML), confirmer le chargement effectif du backend applicatif et afficher la version exacte (`python --version`) ainsi que le mode d'attribution (`.venv` ou `PYTHON_BIN`) ;
- **Action [M] (Modification)** : permettre la saisie de nouvelles valeurs et persister les modifications dans le fichier `.env` à la racine de l'application sans aucune recompilation du binaire (y compris la variable `PYTHON_BIN` avec vérification préalable systématique du candidat : si la sonde échoue ou que des dépendances manquent, la modification est refusée et l'ancienne valeur est conservée sans modification du `.env`) ;
- **Action [R]** : retour au menu principal.

## Points clés

- **Conformité stricte aux exigences d'évaluation** : respect de la règle d'or "tout doit être dans un `.env`, aucune recompilation requise lors d'un changement de modèle ou d'adresse" ;
- **Diagnostic intégré complet** : permet de valider instantanément si le serveur de l'IUT (`10.22.28.190:11434`) répond et si l'environnement d'exécution Python autonome ou personnalisé est opérationnel ;
- **Saisie sécurisée** : utilisation de `ReadChoice` pour éviter toute sortie accidentelle lors d'un appui sur Entrée à vide (retour exigeant explicitement la touche `[R]`).

## Rôle dans l’architecture

Interface d'administration et de diagnostic du front-end, garantissant la configurabilité à chaud du système.

## Commentaires de fonction (dupliqués du code source)

- `ShowConfigFlow()` : Affiche et gère le flux interactif de configuration.
