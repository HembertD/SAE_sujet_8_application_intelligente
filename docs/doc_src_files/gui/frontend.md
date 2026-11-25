# Interface Utilisateur Web (Frontend Wails)

Fichiers sources associés : `src/gui/frontend/index.html`, `src/gui/frontend/src/style.css`, `src/gui/frontend/src/main.js`

## Description

Le frontend de l'application graphique est conçu selon les standards du Web moderne (HTML5 sémantique, CSS3 Glassmorphism natif, JavaScript ES Modules, bundler Vite). Il communique de manière bidirectionnelle avec le backend Go via le pont IPC de Wails v2 sans aucun rechargement de page.

## Vues et Fonctionnalités

1. **Dashboard & État Git (`#view-dashboard`)** :
   - Affichage synthétique des 4 KPIs principaux : branche active, fichiers modifiés (staged vs unstaged), modèle LLM Ollama et état du conteneur Docker ;
   - Accès rapide aux actions clés (génération de commit, inspection diff, indexation complète, push) ;
   - Liste des derniers commits avec identifiants SHA courts, messages, auteurs et badges de tags ;
   - Moniteur de ressources système en temps réel alimenté par Go (`runtime.MemStats`, goroutines, CPU).

2. **Inspecteur de Diffs (`#view-diff`)** :
   - Liste latérale des fichiers modifiés avec pastilles d'état (M, A, D, R) et différentiel de lignes (`+` vert, `-` rouge) ;
   - Visualiseur de patch centralisé avec coloration syntaxique complète des ajouts, suppressions et en-têtes de hunks ;
   - Détection visuelle des avertissements de sécurité (fichiers contenant des secrets masqués ou fichiers binaires).

3. **Générateur de Commits IA (`#view-commit`)** :
   - Proposition découpée selon la norme Conventional Commits (Type, Scope, Sujet, Corps) ;
   - Barre de sélection de type imposé (`[T]` du CLI) : permet de forcer en un clic `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `build` ;
   - Éditeur de message en texte brut permettant d'ajuster ou enrichir le texte avant validation ;
   - Champ de feedback utilisateur interactif permettant d'ajuster et affiner la proposition du LLM (`[R]` du CLI) ;
   - Boutons pour copier dans le presse-papier, valider le commit (`git commit`) ;
   - Proposition post-commit automatique pour pousser les modifications vers le remote (`git push`) comme dans le CLI.

4. **Rédacteur de Release Notes (`#view-release-notes`)** :
   - **Double sélecteur interactif de commits** (copie conforme de l'expérience CLI `release_notes_view.go`) :
     - Deux colonnes côte-à-côte présentant l'ensemble de l'historique Git (`GetRecentCommits(0)`) avec SHA court, badges de tags (ex. `v0.1.0`), date relative, auteur et message complet ;
     - Barre de recherche instantanée par colonne pour filtrer par message, tag, auteur ou hash de commit ;
     - Cartouche récapitulatif interactif montrant en temps réel l'intervalle sélectionné (`[Commit Début] ➔ [Commit Fin]`) et le nombre exact de commits inclus dans la plage ;
     - Boutons de raccourcis rapides (Presets) : *Tous l'historique*, *Depuis le dernier tag*, *5 derniers commits*, *10 derniers commits* ;
   - Génération assistée par IA avec compteur de commits analysés ;
   - Double prévisualisation : vue Markdown brut et vue HTML mise en forme avec typographie soignée ;
   - Export et sauvegarde directe du fichier `.md` sur le système de fichiers.

5. **Configuration, Python & Diagnostics (`#view-config`)** :
   - **Section Environnement & Interpréteur Python** (copie conforme de la CLI) :
     - Statut de l'environnement, version de Python et exécutable résolu ;
     - Configuration du chemin personnalisé `PYTHON_BIN` avec boutons rapides (`~/monenv`, `Auto (.venv)`) ;
     - Test d'un interpréteur candidat en direct (`VerifyPythonCandidate`) avant enregistrement ;
     - Outil de diagnostic approfondi (`[P]` du CLI) avec rapport détaillé (dépendances GitPython/PyYAML et module backend) ;
   - Réglages complets du fichier `.env` (URL Ollama, modèle, délai d'expiration) ;
   - Outil de test de connexion (Ping Ollama, option `[T]` du CLI) avec détection et affichage des modèles installés sur le serveur ;
   - **Gestion avancée du conteneur Docker Ollama & Console de logs** :
     - Indicateur d'état dynamique avec pastille pulsante (vert = prêt/healthy, bleu = téléchargement du modèle, jaune = démarrage, rouge = anomalie, gris = arrêté) ;
     - Jauge de progression animée lors du téléchargement (`ollama pull`) avec affichage du dernier log de transfert ;
     - Barre d'actions complète : Démarrer (`up -d --build`), Redémarrer (`restart ollama`), Arrêter (`down`) et Test HTTP local ;
     - **Visualiseur de journaux d'exécution en direct** (`docker compose logs`) : terminal intégré avec coloration syntaxique des logs (téléchargements, erreurs, serveur prêt), auto-refresh réglable (5s à la demande), défilement automatique, choix du nombre de lignes (50 à 500) et copie en un clic dans le presse-papier ;
   - Interrupteurs pour le mode développeur (Mock Interface, Mock IA) et notifications de bureau ;
   - Bannière d'alerte Python en haut de page si l'environnement présente une anomalie.

## Design & Ergonomie

- **Cyber Glassmorphism Haute Performance** : Arrière-plans sombres profonds (`#090d16` / `#0d121c`), bordures subtiles et orbes d'ambiance lumineuses (`glow-orb` à gradients radiaux statiques sans filtre blur logiciel CPU) garantissant une fluidité à 60 FPS même en rendu logiciel GTK/Wayland ;
- **Défilement isolé par vue** : Chaque onglet (`.tab-view`) dispose de son propre conteneur de scroll isolé avec réinitialisation systématique des positions de défilement lors du changement de vue, évitant toute propagation de scroll entre onglets sous WebKitGTK ;
- **Thèmes & Accents dynamiques** : Prise en charge complète du thème sombre et du thème clair, avec sélecteur de palette d'accents en direct (Cyan, Violet, Émeraude, Sunset, Rose) conservé dans le `localStorage` ;
- **Notifications Toast** : Système de bulles d'information contextuelles animées (succès, avertissement, erreur).
