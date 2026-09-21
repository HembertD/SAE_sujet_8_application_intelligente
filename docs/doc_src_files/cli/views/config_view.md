# Vue Configuration

Fichier source associé : src/cli/views/config_view.go

## Description

Ce module permet de visualiser et de modifier les paramètres opérationnels de l'application (URL Ollama, modèle LLM, timeout, langue) et de diagnostiquer la connectivité réseau avec le serveur de l'IUT.

## Responsabilités

- afficher les paramètres actuellement appliqués issus du fichier d'environnement `.env` ;
- **Action [T] (Test Ping)** : exécuter un diagnostic réseau vers l'API Ollama, mesurer la latence en millisecondes et lister l'ensemble des modèles d'IA installés sur le serveur ;
- **Action [M] (Modification)** : permettre la saisie de nouvelles valeurs (y compris l'activation ou la désactivation à chaud du mock) et persister les modifications dans le fichier `.env` sans aucune recompilation du binaire ;
- **Action [R]** : retour au menu principal.

## Points clés

- **Conformité stricte aux exigences d'évaluation** : respect de la règle d'or "tout doit être dans un `.env`, aucune recompilation requise lors d'un changement de modèle ou d'adresse" ;
- **Diagnostic intégré** : permet de valider instantanément si le serveur de l'IUT (`10.22.28.190:11434`) répond avant de lancer des opérations de génération.

## Rôle dans l’architecture

Interface d'administration et de diagnostic du front-end, garantissant la configurabilité à chaud du système.
