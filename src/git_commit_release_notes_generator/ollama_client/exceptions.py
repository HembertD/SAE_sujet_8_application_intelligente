"""Exceptions du client Ollama (module C)."""


class OllamaCallError(RuntimeError):
    """Erreur lors de l'appel à Ollama (réseau, réponse incomplète ou illisible)."""


class MalformedResponseError(OllamaCallError):
    """Le contenu renvoyé par Ollama n'est pas du JSON valide, malgré le `format` contraint.

    Sous-classe de OllamaCallError (donc toujours attrapable comme tel), mais
    distinguée pour permettre à generate_commit_message() de la traiter comme
    un échec de validation retryable plutôt qu'une erreur fatale immédiate.
    """

    def __init__(self, content: str, message: str) -> None:
        super().__init__(message)
        self.content = content


class CommitMessageValidationError(RuntimeError):
    """Le message de commit généré ne respecte pas le format attendu, même après retry."""


class ConflictResolutionError(RuntimeError):
    """La proposition de résolution de conflit de merge est invalide, même après retry
    (marqueurs de conflit non résolus, contenu vide, ou réponse non exploitable)."""
