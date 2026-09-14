"""
Active l'environnement de test en restaurant tous les éléments Git et sensibles (suffixe '_test' retiré).
Restaure :
- .git_test -> .git
- .gitignore_test -> .gitignore
- .github_test -> .github
- .gitattributes_test -> .gitattributes
- .gitmodules_test -> .gitmodules
- .env_test -> .env

Compatible multi-OS (Linux, macOS, Windows) via pathlib.
"""

from pathlib import Path
import shutil
from typing import List, Tuple, Optional

# Éléments ayant un impact avec Git ou sécurité (.env) à réactiver
GIT_TARGETS = [".git", ".gitignore", ".github", ".gitattributes", ".gitmodules", ".env"]


def init_test(base_dir: Optional[Path] = None) -> List[Tuple[Path, Path]]:
    """
    Restaure tous les fichiers et dossiers désactivés (*_test -> *) dans base_dir.
    Ne produit aucun affichage dans la console.

    :param base_dir: Répertoire racine de recherche (par défaut le dossier du script).
    :return: Liste de tuples (chemin_source, chemin_cible) pour chaque élément renommé.
    """
    if base_dir is None:
        base_dir = Path(__file__).resolve().parent
    else:
        base_dir = Path(base_dir).resolve()

    renamed: List[Tuple[Path, Path]] = []

    for name in GIT_TARGETS:
        test_name = f"{name}_test"
        items = [p for p in base_dir.rglob(test_name)]
        # Traitement du plus profond au plus superficiel
        items.sort(key=lambda p: len(p.parts), reverse=True)

        for item in items:
            target = item.with_name(name)
            if target.is_dir():
                shutil.rmtree(target)
            elif target.exists():
                target.unlink()
            item.rename(target)
            renamed.append((item, target))

    return renamed


if __name__ == "__main__":
    print("=== Initialisation de l'environnement de test (*_test -> actifs) ===")
    results = init_test()
    if not results:
        print("ℹ️  Aucun élément désactivé (*_test) trouvé.")
    else:
        for src, dst in results:
            print(f"✔ Activé : {src.name} -> {dst.name} ({dst.parent.name})")
        print(f"🎉 {len(results)} élément(s) réactivé(s) avec succès !")
