# Système d'Aide CLI et Documentation Intégrée

Fichier source associé : `src/cli/ui/help.go`

## Description

Ce module implémente le système complet d'aide en ligne de commande de Gitelligence (`--help`, `-h`). Conforme aux critères d'acceptation de la User Story #90, il fournit un format standard Linux concis et direct détaillant la syntaxe, les options, l'initialisation de l'environnement Python et la configuration `.env`.

## Responsabilités

- **Identification de l'application** : référence les constantes centralisées du package `version` (`AppName`, `BinaryName`, `AppVersion`) ;
- **Intégration du module de drapeaux** :
  - Découplage de la définition des drapeaux via le paquet `sae-git-cli/flags` ;
  - `EnvVariables` : documente toutes les variables configurables du `.env` (`OLLAMA_BASE_URL`, `OLLAMA_MODEL`, `OLLAMA_TIMEOUT_S`, `OLLAMA_NUM_CTX`, `ENABLE_NOTIFICATIONS`, `PYTHON_BIN`, `MOCK_INTERFACE`, `MOCK_AI`, `USE_DOCKER_LLM`) ;
- **Génération et rendu dynamique** :
  - `GetGeneralHelp()` : découvre dynamiquement l'ensemble des drapeaux enregistrés dans `flag.CommandLine` (provenant de `sae-git-cli/flags`) via `VisitAll()`, détecte automatiquement le typage (booléen ou avec paramètre `<chemin>` / `<valeur>`), extrait les valeurs par défaut et met en page le rendu avec coloration ANSI et alignement optimal ;
  - `PrintHelp(w)` : écrit l'aide sur le flux de sortie fourni (`os.Stdout`) ;
  - `PrintVersion(w)` et `GetVersionText()` : affichent la version applicative ;
- **Détection robuste des requêtes CLI** :
  - `IsHelpRequested(args)` : analyse les arguments et drapeaux pour identifier si `-h`, `--help`, `-help` ou `help` est invoqué, tout en interrogeant dynamiquement `flag.CommandLine.Lookup()` pour sauter les valeurs des options avec argument (évitant tout faux positif sur `--repo help`) ;
  - `IsVersionRequested(args)` : détecte la présence des drapeaux `-v` ou `--version` avec le même niveau d'isolation dynamique.

## Guide Développeur : Ajouter une nouvelle option CLI

Grâce à la détection dynamique, l'ajout d'une nouvelle option se fait désormais dans le paquet dédié :
1. Déclarer l'option dans `src/cli/flags/flags.go` :
   ```go
   MonOption = flag.String("mon-option", "defaut", "Description de l'option")
   ```
2. Le drapeau apparaît immédiatement dans `./git-generator --help`, son type (`<chemin>` ou `<valeur>`) et sa valeur par défaut sont inférés, et `IsHelpRequested` protège automatiquement sa valeur contre les faux positifs.

## Guide Développeur : Ajouter une variable d'environnement (.env)

Pour documenter une nouvelle variable de configuration applicative dans l'aide CLI :
1. Déclarer la variable et sa valeur par défaut dans `.env.exemple` (et dans le backend Python `config.py` si nécessaire) ;
2. Ajouter une entrée dans le tableau `EnvVariables` de `src/cli/ui/help.go` :
   ```go
   {
       Name:        "MA_NOUVELLE_VARIABLE",
       Default:     "valeur_par_defaut",
       Description: "Description courte de son rôle en français",
   },
   ```
3. Elle s'affichera immédiatement dans la section `Variables d'environnement (.env) :` de `./git-generator --help` avec coloration et alignement automatique.

## Points clés

- **Code de sortie 0** : l'affichage de l'aide et de la version termine immédiatement le programme avec le code de succès `0` sans charger le moteur TUI ni le pont Python ;
- **Harmonie visuelle** : mise en valeur par coloration ANSI (`Cyan`, `Green`, `Magenta`, `Dim`, `Bold`, `Reset`) ;
- **Style Linux standard** : présentation concise et directe calquée sur les commandes Unix usuelles.

## Rôle dans l'architecture

Composant d'affichage consommé directement par le point d'entrée `main.go` avant l'initialisation de Bubble Tea et du pont Python.
