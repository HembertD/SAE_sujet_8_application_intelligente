import subprocess
from pathlib import Path


class GitService:

    @staticmethod
    def run_command(*args: str) -> str:
        """
        Execute une commande Git et retourne sa sortie standard.

        Args:
            *args: Arguments de la commande Git.

        Returns:
            La sortie standard de la commande.

        Raises:
            RuntimeError: Si la commande Git échoue.
        """
        try:
            result = subprocess.run(
                ["git", *args],
                capture_output=True,
                text=True,
                check=True
            )
            return result.stdout.strip()
        except FileNotFoundError as exception:
            raise RuntimeError("Git n'est pas installé ou n'est pas accessible.") from exception
        except subprocess.CalledProcessError as exception:
            raise RuntimeError(
                f"La commande Git a échoué : git {' '.join(args)}\n{exception.stderr.strip()}"
            ) from exception

    @staticmethod
    def is_git_repository() -> bool:
        """
        Vérifie si le répertoire courant est un dépôt Git.

        Returns:
            True si le répertoire courant appartient à un dépôt Git, sinon False.
        """
        try:
            GitService.run_command("rev-parse", "--is-inside-work-tree")
            return True
        except RuntimeError:
            return False

    @staticmethod
    def initialize_hooks() -> None:
        """
        Configure Git pour utiliser le dossier .githooks comme répertoire de hooks.
        """
        if not GitService.is_git_repository():
            raise RuntimeError("Le répertoire courant n'est pas un dépôt Git.")

        repository_root = Path(
            GitService.run_command("rev-parse", "--show-toplevel")
        )

        hooks_directory = repository_root / ".githooks"

        if not hooks_directory.is_dir():
            raise RuntimeError("Le dossier .githooks n'existe pas.")

        GitService.run_command(
            "config",
            "core.hooksPath",
            ".githooks"
        )