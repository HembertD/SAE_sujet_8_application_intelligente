# Activation des situations de test

Fichier source associé : test/situation_test/init_test.py

## Description

Ce script réactive les éléments masqués pour les tests, en restaurant les fichiers et dossiers Git ou sensibles renommés avec le suffixe `_test`.

## Rôle fonctionnel

Il sert à remettre dans leur état nominal les éléments suivants :

- `.git` → `.git`
- `.gitignore` → `.gitignore`
- `.github` → `.github`
- `.gitattributes` → `.gitattributes`
- `.gitmodules` → `.gitmodules`
- `.env` → `.env`

Le suffixe `_test` est utilisé pour neutraliser les éléments sensibles sans les supprimer du dépôt.

## Comportement

Le script parcourt le dossier racine donné, recherche les entrées portant le suffixe `_test`, puis les renomme vers leur nom original. Il traite d’abord les éléments les plus profondément imbriqués pour éviter les conflits de renommage.

## Utilité dans le projet

Cette étape est nécessaire lorsque les situations de test doivent être exécutées dans un environnement proche du vrai dépôt, notamment pour tester les interactions Git ou les fichiers de configuration sensibles.
