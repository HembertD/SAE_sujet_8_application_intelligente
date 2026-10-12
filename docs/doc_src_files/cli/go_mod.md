# Module Go CLI

Fichier source associé : src/cli/go.mod

## Description

Ce fichier définit le module Go standard (`sae-git-cli`) utilisé pour compiler l'interface terminal (TUI / CLI) de l'application.

## Responsabilités

- déclarer le nom canonique du module Go (`sae-git-cli`) ;
- fixer la version minimale requise du compilateur Go (`go 1.22`) ;
- garantir l'autonomie totale du module sans dépendances tierces externes.

## Points clés

- **Zéro dépendance externe** : l'interface utilise exclusivement les packages de la bibliothèque standard Go (`fmt`, `os`, `os/exec`, `bufio`, `encoding/json`, `time`, `os/signal`, `regexp`, `math`) ;
- compilation instantanée et portabilité native sur tout environnement Linux / macOS sans téléchargement via proxy ou connexion internet.

## Rôle dans l’architecture

Point d'ancrage pour le gestionnaire de paquets et l'outil de build Go (`go build`, `go run`).
