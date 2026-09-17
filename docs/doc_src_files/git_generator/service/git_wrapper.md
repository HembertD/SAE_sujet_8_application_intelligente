# Git Wrapper

Fichier source associé : src/git_generator/service/git_wrapper.py

## Description

Ce module encapsule l’accès au dépôt Git et transforme les informations brutes en objets exploitables par l’application.

## Responsabilités

- ouvrir un dépôt Git valide ;
- extraire le diff indexé ;
- détecter les fichiers binaires ;
- masquage des secrets dans les patches ;
- structurer les données sous forme de DiffFile ;
- récupérer les commits entre deux tags ;
- réaliser un commit Git avec un message fourni.

## Points clés

- la méthode get_staged_diff retourne une liste de fichiers au format structuré ;
- la méthode sanitize_diff masque automatiquement les secrets détectés ;
- les objets DiffFile encapsulent les informations nécessaires au modèle LLM ;
- les erreurs sont remontées sous forme d’exception métier GitWrapperError.

## Rôle dans l’architecture

Le wrapper est la couche d’accès au dépôt Git. Il permet de garder le reste de l’application indépendant des détails de la commande Git et des objets bruts.
