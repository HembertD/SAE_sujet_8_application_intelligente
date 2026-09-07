from git_commit_release_notes_generator.service.git import GitService


def main() -> None:
    """
    Point d'entrée de l'application.
    """
    GitService.initialize_hooks()
    print("Git hooks initialisés avec succès.")


if __name__ == "__main__":
    main()