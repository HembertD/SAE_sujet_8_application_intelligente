# Animation Spinner et Gestion Asynchrone

Fichier source associé : src/cli/ui/spinner.go

## Description

Ce module gère un indicateur de chargement animé (spinner) non-bloquant basé sur une goroutine Go, apportant un retour visuel en temps réel pendant les opérations asynchrones (inférence LLM, récupération Git, ping réseau).

## Responsabilités

- démarrer une goroutine indépendante affichant une séquence de caractères animés braille (`⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏`) à une fréquence de 80ms ;
- afficher un chronomètre dynamique en temps réel mesurant le temps d'attente (ex: `(3.4s)`) ;
- masquer le curseur du terminal pendant l'animation pour éviter les clignotements parasites (`HideCursor`) ;
- mettre à jour dynamiquement le texte d'attente (`UpdateLabel`) ;
- stopper proprement l'animation à l'aide d'un channel de synchronisation (`close(stopChan)`) et restaurer le curseur (`ShowCursor`).

## Points clés

- **Thread-safe** : protection des accès concurrents au libellé et à l'état actif via un verrou d'exclusion mutuelle (`sync.Mutex`) ;
- **Non-bloquant** : permet au thread principal d'attendre la réponse du backend Python sans figer l'interface ;
- **Nettoyage propre de la ligne** : réécriture en place via `\r\033[K` pour une transition sans saut de ligne indésirable.

## Rôle dans l’architecture

Composant d'attente visuelle fournissant le feedback indispensable à l'utilisateur lors des requêtes d'inférence LLM vers le serveur distant d'Ollama.
