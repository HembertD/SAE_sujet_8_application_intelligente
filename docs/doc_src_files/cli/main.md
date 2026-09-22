# Point d'Entrée Principal CLI

Fichier source associé : src/cli/main.go

## Description

Ce fichier constitue le point d'entrée exécutable de l'interface en ligne de commande. Il initialise l'environnement, configure l'interception des signaux système, instancie le client de pontage vers Python et pilote la boucle d'événements principale du menu.

## Responsabilités

- analyser les arguments et flags de la ligne de commande (`--demo`, `--repo`) ;
- intercepter les signaux d'interruption système (`SIGINT` / `SIGTERM` / `Ctrl+C`) pour restaurer l'état du terminal (réaffichage du curseur et réinitialisation des couleurs) ;
- initialiser le `BackendClient` relié au répertoire du dépôt Git ;
- interroger l'état du dépôt auprès du backend Python à chaque itération ;
- aiguiller l'utilisateur vers les différentes vues (`commit`, `release-notes`, `config`, ou sortie propre).

## Points clés

- **Gestion des signaux OS** : mise en place d'une goroutine dédiée et d'un `signal.Notify` pour intercepter `Ctrl+C` et garantir que le terminal ne reste jamais dans un état dégradé (curseur masqué) ;
- **Boucle interactive robuste** : gestion de la fin de flux (EOF) et réaffichage automatique du contexte après chaque action.

## Rôle dans l’architecture

`main.go` est le chef d'orchestre du front-end. Il lie la couche de présentation (vues et utilitaires ANSI) au client de pontage backend sans intégrer la moindre règle métier.
