# Module de Version Applicative

Fichier source associé : `src/cli/version/version.go`

## Description

Ce module isole et centralise les métadonnées globales d'identification et de numéro de version de Gitelligence (`git-generator`). En découplant ces informations du module d'aide et de l'interface graphique, il garantit un point de vérité unique (SSOT - *Single Source of Truth*) accessible par l'ensemble des paquets Go et injectable lors de la compilation.

## Constantes et Fonctions

- **`Number`** : numéro sémantique de version de l'application (ex : `"1.0.0"`) ;
- **`AppName`** : nom officiel de l'application (`"Gitelligence"`) ;
- **`BinaryName`** : nom de l'exécutable (`"git-generator"`) ;
- **`String()`** : formate et renvoie l'en-tête d'identification complet au format :
  ```
  Gitelligence (git-generator) version 1.0.0
  ```

## Rôle dans l'architecture

Paquet transversal de base sans dépendance externe, importé par le module d'aide (`ui/help.go`), le point d'entrée (`main.go`) et les tests automatisés.
