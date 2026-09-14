"""Exceptions du client Ollama (module C)."""


class OllamaCallError(RuntimeError):
    """Erreur lors de l'appel à Ollama (réseau, réponse incomplète ou illisible)."""


class CommitMessageValidationError(RuntimeError):
    """Le message de commit généré ne respecte pas le format attendu, même après retry."""
