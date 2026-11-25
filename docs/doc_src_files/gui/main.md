# Point d'Entrée Principal GUI (Wails v2)

Fichier source associé : `src/gui/main.go`

## Description

Ce fichier constitue le point d'entrée exécutable de l'interface graphique native développée avec **Wails v2**. Il parse les arguments de ligne de commande, contrôle la cohérence des options de simulation/développement, instancie l'orchestrateur `App`, et lance la fenêtre applicative Webview intégrée avec WebKitGTK 4.1.

## Responsabilités

- **Analyse des drapeaux CLI** :
  - `--repo <chemin>` : dépôt Git cible à analyser (défaut : `.`) ;
  - `--app-dir <chemin>` : racine de l'application Gitelligence hébergeant le fichier `.env` ;
  - `--dev` : active le mode développement (débloque les simulations d'interface et d'IA) ;
  - `--mockInterface` : active le mockage complet de l'interface (mode autonome sans Git ni Python, requiert `--dev`) ;
  - `--mockIA` : active la simulation de l'IA sans serveur Ollama (requiert `--dev`) ;
  - `--docker-llm` : force l'utilisation du conteneur Ollama local piloté par Docker Compose ;
  - `--python-bin <chemin>` : spécifie explicitement l'interpréteur Python à utiliser (surcharge le `.env` et `.venv`) ;
- **Contrôle d'accès au mode mock** : vérifie que l'utilisation de `--mockInterface` ou `--mockIA` s'accompagne bien de `--dev` sous peine d'interruption immédiate ;
- **Validation stricte de l'environnement Python** : exécute `bridge.VerifyPythonEnvironment` au démarrage (exactement comme le CLI) pour alerter immédiatement sur les dépendances manquantes ou l'interpréteur invalide ;
- **Localisation de la racine applicative** : fait appel à `bridge.FindAppRoot` pour localiser le `.env` de manière robuste ;
- **Configuration et démarrage de Wails** :
  - Intègre les assets Web statiques compilés (`dist/`) via le système `embed.FS` natif de Go ;
  - Définit les dimensions de la fenêtre (1240x840, minimum 980x680) et la couleur d'arrière-plan cyber glassmorphism `#090D16` ;
  - Lie l'instance `App` (`Bind: []interface{}{ app }`) pour exposer automatiquement ses méthodes publiques vers JavaScript (`window.go.main.App.*`).

## Compilation et Build Tags

Wails v2 sous Linux requiert des build tags explicites pour lier le runtime de production et la bibliothèque WebKitGTK 4.1 (`webkit2gtk-4.1`) :
- **Via Wails CLI** (recommandé) :
  ```bash
  cd src/gui
  ~/go/bin/wails build
  ```
  *(La balise `"build:tags": "webkit2_41"` configurée dans `wails.json` est transmise automatiquement)*.
- **Via Go directement** :
  ```bash
  cd src/gui
  go build -tags "production,webkit2_41" -o ../../git-gui .
  ```
