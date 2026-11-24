# Point d'Entrée Principal CLI

Fichier source associé : `src/cli/main.go`

## Description

Ce fichier constitue le point d'entrée exécutable de l'interface en ligne de commande. Il intercepte immédiatement les demandes d'aide (`--help`, `-h`) et de version (`--version`, `-v`), consomme les drapeaux CLI définis dans `flags/flags.go`, instancie le client de pontage vers Python et pilote la boucle interactive du menu TUI principal.

## Responsabilités

- **Interception prioritaire d'aide et de version** : détecte `-h`, `--help`, `help` ou `-v`, `--version` pour afficher la documentation appropriée et quitter immédiatement avec le code de succès `0`, sans charger Python ni la TUI ;
- **Analyse des drapeaux de la ligne de commande** :
  - `--repo <chemin>` : dépôt Git cible à analyser (défaut : `.`) ;
  - `--app-dir <chemin>` : racine de l'application Gitelligence contenant le `.env` (auto-détecté) ;
  - `--dev` : active le mode développeur (débloque les options de mock pour développer l'application) ;
  - `--mockInterface` : active le mockage complet de l'interface pour le développement (sans Git ni Python, requiert `--dev`) ;
  - `--mockIA` : active le mockage de l'IA pour le développement (sans serveur Ollama, requiert `--dev`) ;
  - `--docker-llm` : force l'usage du serveur Ollama local via conteneur Docker ;
  - `--python-bin <chemin>` : chemin explicite vers l'interpréteur Python (surcharge le `.env` et le `.venv`) ;
  - `-h, --help` : affiche l'aide détaillée ;
  - `-v, --version` : affiche la version du programme ;
- **Contrôle d'accès au mode mock** : vérifie que l'utilisation de `--mockInterface` ou `--mockIA` s'accompagne bien de `--dev` sous peine d'interruption immédiate ;
- **Vérification automatique et approfondie de l'environnement Python** : s'assure qu'un environnement valide est configuré (`bridge.VerifyPythonEnvironment`) avant toute initialisation du backend, sauf en mode mock-interface ;
- **Interception des signaux OS** : capture `SIGINT` / `SIGTERM` (`Ctrl+C`) pour restaurer la visibilité du curseur et réinitialiser les attributs du terminal ;
- **Pilote de la boucle interactive TUI** : initialise le `BackendClient`, interroge l'état du dépôt Git et dirige l'utilisateur vers les différentes vues (`views.ShowMainMenu`).

## Arguments de la ligne de commande

| Option | Valeur par défaut | Description |
|---|---|---|
| `-h, --help` | `-` | Affiche l'aide détaillée et quitte avec le code 0 |
| `-v, --version` | `-` | Affiche la version de Gitelligence et quitte avec le code 0 |
| `--repo <chemin>` | `.` | Chemin vers le dépôt Git cible à analyser |
| `--app-dir <chemin>` | Auto-détecté | Racine de l'application Gitelligence contenant le `.env` |
| `--dev` | `false` | Active le mode développeur (débloque les options de mock pour développer l'application) |
| `--mockInterface` | `false` | Active le mockage complet de l'interface pour le développement (mode sans Git ni Python, requiert `--dev`) |
| `--mockIA` | `false` | Active le mockage de l'IA pour le développement (sans serveur Ollama, requiert `--dev`) |
| `--docker-llm` | `false` | Force l'usage du serveur Ollama local via conteneur Docker |
| `--python-bin <chemin>` | Auto-détecté | Chemin explicite vers l'interpréteur Python à utiliser |

## Points clés

- **Code de sortie 0 garanti sur l'aide et la version** : conformité stricte avec les standards Unix ;
- **Restauration propre du terminal** : gestion des signaux pour réactiver le curseur ANSI en toute circonstance.

## Rôle dans l'architecture

`main.go` est le chef d'orchestre du front-end. Il lie la couche de présentation (vues et utilitaires ANSI) au client de pontage backend sans intégrer la moindre règle métier.

## Détails déplacés du code source

Précisions issues des commentaires retirés de `main.go` et non redondantes avec les sections ci-dessus :

- En cas d'erreur de `GetRepoStatus`, le statut est dégradé à `nil` (`status = nil`) et le menu principal est quand même affiché via `views.ShowMainMenu`.
- Toute option non reconnue dans le `switch` (`default`) provoque un retour direct au menu, sans message d'erreur ni sortie.
- Le contrôle de l'environnement Python (`VerifyPythonEnvironment`) vérifie à la fois la résolution de l'interpréteur, la présence de toutes les dépendances indispensables (`GitPython`, `PyYAML`) et le chargement du backend applicatif. Il arrête immédiatement l'application avec `ui.PrintErrorCard` et des conseils de remédiation clairs en cas de manquement. Il est volontairement sauté si le mode mock-interface est actif (flag `--mockInterface` ou `.env`), puisqu'aucun sous-processus Python n'est alors jamais lancé.

## Commentaires de fonction (dupliqués du code source)

- `main()` : Point d'entrée initialisant l'interface et pilotant le menu principal.
