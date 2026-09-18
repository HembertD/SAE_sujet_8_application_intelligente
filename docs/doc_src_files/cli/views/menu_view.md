# Vue Menu Principal

Fichier source associé : src/cli/views/menu_view.go

## Description

Ce module gère l'affichage de la page d'accueil et la capture du premier niveau d'interaction de l'utilisateur.

## Responsabilités

- effacer l'écran et afficher la bannière contextuelle du projet ;
- présenter les 4 choix d'actions numérotés :
  - `[1]` Générer un message de commit (IA Conventional Commits) ;
  - `[2]` Générer les Release Notes (Synthèse entre tags) ;
  - `[3]` Configuration (Modèle, URL Ollama, Langue) ;
  - `[4]` Quitter ;
- signaler si le mode simulation/démo est actif ;
- capturer la saisie utilisateur et gérer les signaux de fin de flux (EOF).

## Points clés

- **Simplicité d'utilisation** : navigation rapide par touche directe (`1`, `2`, `3`, `4` ou `q`) ;
- **Contexte temps réel** : actualisation de la branche et du nombre de fichiers modifiés à chaque retour sur le menu.

## Rôle dans l’architecture

Écran racine du CLI, point de départ de toutes les fonctionnalités applicatives.
