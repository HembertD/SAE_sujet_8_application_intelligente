# Module des Options et Drapeaux CLI

Fichier source associé : `src/cli/flags/flags.go`

## Description

Ce module centralise la déclaration de l'ensemble des options et drapeaux de la ligne de commande de Gitelligence (`--repo`, `--app-dir`, `--dev`, `--mockInterface`, `--mockIA`, `--docker-llm`, `--python-bin`). En découplant ces définitions du moteur d'affichage (`ui/help.go`) et du point d'entrée exécutable (`main.go`), il garantit une source unique de vérité (*Single Source of Truth*) pour les paramètres CLI.

## Variables et Drapeaux

Toutes les options sont enregistrées directement dans le registre standard `flag.CommandLine` de Go :

- **`Dev`** (`*bool`) : active le mode développeur, autorisant les drapeaux de simulation et débloquant l'accès à la modification des réglages de mock dans la configuration (défaut : `false`) ;
- **`MockInterface`** (`*bool`) : force la simulation complète de l'interface en mode autonome, sans solliciter Git ni exécuter le backend Python (requiert `--dev`, défaut : `false`) ;
- **`MockIA`** (`*bool`) : active la simulation déterministe des réponses du LLM hors-ligne sans serveur Ollama (requiert `--dev`, défaut : `false`) ;
- **`DockerLLM`** (`*bool`) : force l'utilisation du serveur Ollama local conteneurisé sous Docker (défaut : `false`) ;
- **`Repo`** (`*string`) : chemin vers le répertoire du dépôt Git cible à analyser (défaut : `"."`) ;
- **`AppDir`** (`*string`) : chemin explicite vers la racine de l'application Gitelligence hébergeant le code Python et le fichier `.env` applicatif (auto-détecté par défaut) ;
- **`PythonBin`** (`*string`) : chemin explicite vers l'interpréteur Python à utiliser, surchargeant le `.env` et le `.venv` (auto-détecté par défaut).

## Guide Développeur : Ajouter un nouveau drapeau CLI

Pour ajouter une nouvelle option en `--` à l'application :

1. Ajouter la variable dans `src/cli/flags/flags.go` :
   ```go
   MonOption = flag.String("mon-option", "valeurDefaut", "Description claire en français")
   // ou pour un booléen :
   // MonBooleen = flag.Bool("mon-booleen", false, "Description du booléen")
   ```
2. **C'est tout !** Grâce au moteur dynamique de `ui/help.go` :
   - L'option apparaît immédiatement dans le texte de `./git-generator --help` avec sa signature (`--mon-option <valeur>`) et sa valeur par défaut.
   - Les arguments sont automatiquement protégés contre les collisions d'aide (ex: `--mon-option help` ne déclenchera pas l'aide).
3. Utiliser la valeur déréférencée `*flags.MonOption` dans `src/cli/main.go`.

## Rôle dans l'architecture

Paquet fondamental sans dépendance interne, importé par le point d'entrée (`main.go`), le module d'aide (`ui/help.go`) et les suites de tests unitaires (`test/cli/help_test.go`).
