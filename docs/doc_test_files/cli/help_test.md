# Tests Automatisés de l'Option --help et Version CLI

Fichier source associé : `test/cli/help_test.go`

## Description

Ce fichier contient la suite complète de tests automatisés validant le bon fonctionnement de l'option `--help`, `-h`, de la version (`--version`, `-v`) et du code de retour système `0`.

## Tests couverts

1. **`TestVersionPackage`** : vérifie les constantes du package `version` (`Number`, `AppName`, `BinaryName`) et le formattage de `version.String()`.
2. **`TestGeneralHelpContent`** : vérifie que l'aide générée par `ui.GetGeneralHelp()` contient l'ensemble des éléments requis :
   - Nom de l'application, nom du binaire et version ;
   - Description concise du rôle de Gitelligence en français ;
   - Liste des sections standardisées au format Linux (`Utilisation :`, `Options :`, `Initialiser l'environnement Python (.venv) :`, `Variables d'environnement (.env) :`) ;
   - Intégration directe du tutoriel de préparation de l'environnement Python (.venv) ;
   - Documentation de chaque drapeau CLI avec sa valeur par défaut ;
   - Documentation de chaque variable d'environnement `.env` avec sa valeur par défaut.
3. **`TestPrintHelpOutput`** : valide l'écriture de l'aide et de la version sur un flux de sortie (`io.Writer`).
4. **`TestHelpAndVersionDetection`** : teste la robustesse de détection de `ui.IsHelpRequested` et `ui.IsVersionRequested` sur de multiples combinaisons d'arguments (`--help`, `-h`, `help`, `--version`, `-v`, etc.) et garantit que les valeurs d'options (`--repo help`, `--app-dir help`, `--repo -v`) ne déclenchent pas indûment l'aide ou la version.
5. **`TestBinaryHelpEndToEnd`** : test d'intégration de bout en bout exécutant le binaire CLI Go dans un sous-processus pour vérifier :
   - Code de retour `0` strict sur `--help`, `-h`, `help`, `--version`, `-v` ;
   - Non-déclenchement de l'interface interactive TUI ;
   - Présence effective des chaînes clés dans la sortie standard.
6. **`TestDynamicFlagRegistration`** : enregistre un nouveau drapeau à la volée sur `flag.CommandLine` pour vérifier qu'il est instantanément découvert par `GetGeneralHelp()` et pris en compte dans le saut d'arguments de `IsHelpRequested()`.

## Exécution des tests

```bash
cd test/cli
go test -v -run TestHelp ./...
```
