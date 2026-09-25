# Point d'Entrée Principal CLI

Fichier source associé : src/cli/main.go

## Description

Ce fichier constitue le point d'entrée exécutable de l'interface en ligne de commande. Il initialise l'environnement, configure l'interception des signaux système, instancie le client de pontage vers Python et pilote la boucle d'événements principale du menu.

## Responsabilités

- analyser les arguments et flags de la ligne de commande (`--demo`, `--repo`, `--app-dir`) ;
- localiser la racine applicative via `FindAppRoot` pour pointer sur le `.env` de l'application indépendamment du dépôt cible analysé ;
- intercepter les signaux d'interruption système (`SIGINT` / `SIGTERM` / `Ctrl+C`) pour restaurer l'état du terminal (réaffichage du curseur et réinitialisation des couleurs) ;
- initialiser le `BackendClient` relié au répertoire du dépôt Git et à la racine de l'application ;
- interroger l'état du dépôt auprès du backend Python à chaque itération ;
- aiguiller l'utilisateur vers les différentes vues (`commit`, `release-notes`, `config`, ou sortie propre).

## Arguments de la ligne de commande

- `--repo <chemin>` : chemin vers le dépôt Git cible à analyser (défaut : `.`).
- `--demo` : activation forcée du mode simulation hors-ligne (`MockClient`) sans dépendances externes.
- `--app-dir <chemin>` : chemin explicite vers la racine de l'application Smart Commit (où se trouvent le code et le `.env` applicatif). Permet d'isoler à 100% la configuration applicative lorsque le binaire est exécuté sur un dépôt tiers.

## Points clés

- **Gestion des signaux OS** : mise en place d'une goroutine dédiée et d'un `signal.Notify` pour intercepter `Ctrl+C` et garantir que le terminal ne reste jamais dans un état dégradé (curseur masqué) ;
- **Boucle interactive robuste** : gestion de la fin de flux (EOF) et réaffichage automatique du contexte après chaque action ;
- **Séparation stricte dépôt cible vs application** : l'application dissocie totalement le dépôt analysé (`repoRoot`) de l'arborescence de l'outil (`appRoot`), empêchant toute écriture de `.env` dans le projet inspecté.

## Rôle dans l’architecture

`main.go` est le chef d'orchestre du front-end. Il lie la couche de présentation (vues et utilitaires ANSI) au client de pontage backend sans intégrer la moindre règle métier.
