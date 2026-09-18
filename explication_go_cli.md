## Pourquoi avoir choisi Go pour le Front-end ?
* **Vitesse et instantanéité** : Go se compile directement en code machine natif. L'exécutable se lance en moins de 5 millisecondes, sans aucun temps de démarrage d'interpréteur (comme Python ou Node.js).
* **Binaire unique et autonome** : Aucun besoin d'installer de runtime Go sur la machine cible. Le binaire embarque tout.
* **Maîtrise complète du Terminal** : Possibilité de créer une interface TUI (Terminal User Interface) élégante, fluide et animée sans aucune bibliothèque tierce lourde.


## Comment Go communique avec Python (Le Pont / Bridge)

Toute la communication repose sur le fichier [`src/cli/bridge/backend_client.go`](file:///home/maxence/Documents/DATA/Ecole/but3_info/sae_applicationIntelligente/SAE_sujet_8_application_intelligente/src/cli/bridge/backend_client.go).

### Le mécanisme d'appel système (`exec.Command`)
Go n'utilise pas de socket réseau ou de serveur HTTP local pour parler à Python (ce qui serait lourd et risquerait des conflits de ports).  
Go lance simplement Python en sous-processus système :

```go
// Exemple simplifié de ce que fait Go sous le capot :
cmd := exec.Command("python3", "-m", "git_commit_release_notes_generator.service.core", "--action", "diff")
cmd.Dir = repoPath // S'exécute à la racine du dépôt

var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

err := cmd.Run()
```

### Le protocole JSON et la règle de `stdout` vs `stderr`
Pour que le pont fonctionne sans faille, nous avons fixé une règle universelle :
1. **`stdout` (Sortie Standard)** : Réservé **exclusivement** au JSON de données. Python ne doit JAMAIS faire de `print("chargement...")` sur `stdout`.
2. **`stderr` (Sortie d'Erreur)** : Utilisé pour les logs, les traces de débogage ou les messages d'avertissement de Python.

Dès que la commande Python se termine, Go désérialise le JSON brut reçu sur `stdout` directement dans une structure Go typée grâce à `json.Unmarshal` :

```go
var diffResp models.DiffResponse
err := json.Unmarshal(stdout.Bytes(), &diffResp)
// Maintenant, Go manipule des variables fortement typées : diffResp.Files, etc.
```

### Mock

#### Implémentation du mock
Dans la structure `BackendClient`, nous avons défini un drapeau booléen :
```go
type BackendClient struct {
    pythonBin   string
    moduleName  string
    repoRoot    string
    UseMockMode bool   // <--- Le drapeau qui active ou désactive la simulation
}
```

Au démarrage, Go teste si le backend Python répond :
```go
func (b *BackendClient) isBackendAvailable() bool {
    cmd := exec.Command(b.pythonBin, "-m", b.moduleName, "--action", "ping-backend")
    cmd.Dir = b.repoRoot
    err := cmd.Run()
    return err == nil // Si Python plante ou n'existe pas -> renvoie false !
}
```
Si Python ne répond pas (ou si on passe le flag `./git-generator --demo`), `b.UseMockMode` passe à `true`.

#### Les fausses données codées en dur dans les fonctions
Dans chaque fonction du client (`GetDiff`, `GenerateCommit`, `PingOllama`, etc.), une bifurcation `if b.UseMockMode` intercepte l'appel et renvoie des données factices mais réalistes :

```go
func (b *BackendClient) GenerateCommit(feedback string) (*models.CommitProposal, error) {
    // BRANCHE SIMULATION (MOCK) :
    if b.UseMockMode {
        // Pause artificielle de 1.4s pour imiter le temps de réflexion de l'IA
        time.Sleep(1400 * time.Millisecond)

        return &models.CommitProposal{
            Success: true,
            Type:    "feat",
            Scope:   "auth",
            Subject: "ajout de l'authentification OAuth2 sécurisée",
            Body:    "Intègre la gestion des jetons d'accès et filtre automatiquement les variables d'environnement sensibles.",
            RawFormatted: "feat(auth): ajout de l'authentification OAuth2 sécurisée...",
        }, nil
    }

    // BRANCHE RÉELLE (Appel sous-processus Python) :
    out, err := b.runPythonCommand("--action", "generate-commit")
    // Désérialisation du vrai JSON renvoyé par Python...
}
```

#### La bascule automatique vers le réel
Dès que on aura le `core.py` dans `Dev` :
1. `isBackendAvailable()` renverra `true`.
2. `UseMockMode` passera à `false`.
3. Le bloc `if b.UseMockMode` sera ignoré, et Go exécutera le vrai Python sans nécessiter **la moindre modification de code côté Go**.

---

## Architecture Interne du Code Go (`src/cli/`)

### Arborescence détaillée
```text
src/cli/
├── go.mod                     # Définition du module Go (go 1.22, 0 dépendance)
├── main.go                    # Point d'entrée, capture Ctrl+C, boucle du menu
├── models/
│   └── types.go               # Structs Go typées mappées sur les schémas JSON
├── bridge/
│   └── backend_client.go      # Exécution sous-processus Python & parsing JSON
├── ui/
│   ├── styles.go              # Codes ANSI, TrueColor, calcul VisualLen, StripANSI
│   ├── box.go                 # Tracé des fenêtres, bordures Unicode, tableaux
│   └── spinner.go             # Goroutine de chargement dynamique avec chronomètre
└── views/
    ├── menu_view.go           # Écran 1 : Accueil et affichage du contexte dépôt
    ├── commit_view.go         # Écran 2 : Diff, validation, feedback LLM, git push
    ├── release_notes_view.go  # Écran 3 : Sélection tags et rendu Markdown
    └── config_view.go         # Écran 4 : Paramètres .env et test ping Ollama
```

### Rôle de chaque paquet :
* **`models`** : Ne contient aucune logique. Seulement les définitions de structures (`RepoStatus`, `DiffFile`, `CommitProposal`, `ConfigResponse`, etc.).
* **`bridge`** : Le seul composant qui a le droit d'exécuter des commandes système pour joindre Python.
* **`ui`** : La boîte à outils graphique. Gère l'affichage à l'écran, les couleurs, les bordures et les animations.
* **`views`** : Contient les 4 écrans interactifs. Chaque vue prend en paramètre le `bufio.Reader` (pour lire le clavier), le `BackendClient` (pour demander des données) et le `RepoStatus`.


## Fonctionnement Écran par Écran

### Écran 1 : Menu Principal & Contexte Dépôt
* Dès l'ouverture, Go appelle `client.GetRepoStatus()`.
* Python lui répond en JSON :
  ```json
  { "is_git_repo": true, "branch": "interface-cli", "staged_count": 3, "unstaged_count": 1 }
  ```
* Go génère la bannière d'en-tête avec les pastilles de couleur :
  * Si `staged_count > 0` : affiché en vert.
  * Si `staged_count == 0` : affiché en jaune avec avertissement.
* L'utilisateur tape `1`, `2`, `3`, `4` ou `q` pour naviguer.

---

### Écran 2 : Génération de Commit & Boucle de Décision
C'est le cœur de l'outil :

1. **Vérification du diff** : Go demande les fichiers modifiés (`client.GetDiff()`).
2. **Auto-stage** : Si aucun fichier n'est indexé, Go demande :  
   `Voulez-vous indexer automatiquement toutes les modifications (git add -A) ? [O/n]`  
   Si oui, Go demande à Python d'exécuter l'ajout, puis réactualise le diff.
3. **Tableau visuel** : Affiche les fichiers modifiés :
   * `[M]` en cyan pour modifié.
   * `[A]` en vert pour ajouté.
   * `[D]` en rouge pour supprimé.
   * Badge spécial `🔒 [Secret Masqué]` si Dorian a détecté une clé API ou un token dans le patch.
   * Badge `[Binaire]` pour les images ou exécutables exclus.
4. **Appel IA** : Le spinner tourne pendant que le modèle `gemma4:12b` réfléchit.
5. **Carte Conventional Commits** : Affichage encadré du résultat structuré :
   * `Type` (`feat`, `fix`, `docs`, etc.)
   * `Scope` (`auth`, `cli`, etc.)
   * `Subject` (la phrase de commit)
   * `Body` (explications complémentaires)
6. **Boucle d'actions** :
   * **`[V] Valider`** : Go demande à Python d'appliquer le commit (`client.ApplyCommit`). Python exécute `git commit` et renvoie le SHA (ex: `7a9e2f4`). Go demande ensuite si l'utilisateur souhaite faire un `git push`.
   * **`[R] Régénérer avec consigne`** : Go demande un texte à l'utilisateur (ex: *"Sois plus concis"* ou *"C'est plutôt un fix"*). Go renvoie ce feedback à Python pour réinterroger le LLM.
   * **`[M] Modifier manuellement`** : Permet de retaper directement le sujet ou le body dans la console avant de valider.
   * **`[A] Annuler`** : Revient au menu sans rien toucher.

---

### Écran 3 : Génération des Release Notes
1. Go demande à l'utilisateur :
   * Le tag de départ (par défaut `v0.1.0`).
   * Le tag d'arrivée (par défaut `HEAD`).
2. Python extrait les commits entre ces deux bornes et demande au LLM de rédiger un changelog structuré.
3. Go affiche le document Markdown dans une boîte dédiée.
4. L'utilisateur peut appuyer sur `[S]` pour enregistrer automatiquement le résultat dans un fichier `RELEASE_NOTES.md` à la racine du dépôt.

---

### Écran 4 : Configuration (.env) & Diagnostic Ping Ollama
Cet écran répond directement aux **consignes impératives de contrôle de l'IUT** :
* Affiche les variables lues par Python :
  * `OLLAMA_BASE_URL` (par défaut `http://10.22.28.190:11434`)
  * `OLLAMA_MODEL` (par défaut `gemma4:12b`)
  * `OLLAMA_TIMEOUT_S` (60s)
  * `APP_LANGUAGE` (fr)
* **Action `[T]` (Test Ping)** : Go demande à Python d'effectuer une requête rapide vers le serveur Ollama. S'il répond, l'écran affiche en vert la latence en millisecondes et la liste réelle des modèles installés sur le serveur !
* **Action `[M]` (Modification)** : Permet de modifier l'adresse IP, le modèle ou le timeout. Python sauvegarde immédiatement les valeurs dans le fichier `.env`.  
  👉 **Aucune recompilation du binaire Go n'est nécessaire.**

---

## Guide Pratique : Compiler, Tester et Présenter

### Compiler le projet :
Dans un terminal, placez-vous dans le dossier `src/cli` :
```bash
cd SAE_sujet_8_application_intelligente/src/cli
go build -o git-generator main.go
```

### Lancer l'application :
```bash
# Lancement normal (détecte le backend Python s'il est présent) :
./git-generator

# Lancement forcé en mode simulation (idéal pour démo ou oral si le serveur IUT est éteint) :
./git-generator --demo

# Lancement en ciblant un autre dépôt Git :
./git-generator --repo /chemin/vers/un/autre/projet
```
