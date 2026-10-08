#!/usr/bin/env python3
"""Install or update the West Monroe presentation skill from a trusted checkout."""

import argparse
import hashlib
import os
from pathlib import Path
import shutil
import sys
import time
import uuid

SKILL_NAME = "west-monroe-presentations"


def absolute_path(value):
    return Path(os.path.abspath(os.path.expanduser(str(value))))


def canonical_leaf(path):
    """Resolve parent aliases while preserving a possibly symlinked leaf."""
    path = absolute_path(path)
    return path.parent.resolve(strict=False) / path.name


def overlaps(first, second):
    try:
        first.relative_to(second)
        return True
    except ValueError:
        pass
    try:
        second.relative_to(first)
        return True
    except ValueError:
        return False


def source_files(source):
    if source.is_symlink() or not source.is_dir():
        raise ValueError("source must be a real directory, not a symlink")
    files = []
    for path in sorted(source.rglob("*")):
        if path.is_symlink():
            raise ValueError("source contains a symlink: {}".format(path.relative_to(source)))
        if path.is_dir():
            continue
        if not path.is_file():
            raise ValueError("source contains a non-regular file: {}".format(path.relative_to(source)))
        files.append(path)
    skill_file = source / "SKILL.md"
    if not skill_file.is_file():
        raise ValueError("source folder must contain SKILL.md")
    lines = skill_file.read_text(encoding="utf-8").splitlines()
    if not lines or lines[0].strip() != "---":
        raise ValueError("SKILL.md must have a YAML frontmatter name")
    try:
        end = next(i for i, line in enumerate(lines[1:], 1) if line.strip() == "---")
    except StopIteration:
        raise ValueError("SKILL.md frontmatter is not closed")
    names = [line.partition(":") for line in lines[1:end] if line.partition(":")[0].strip() == "name"]
    if len(names) != 1 or names[0][1] != ":" or names[0][2].strip().strip("\"'") != SKILL_NAME:
        raise ValueError("source SKILL.md must declare name: {}".format(SKILL_NAME))
    return files


def digest_tree(root):
    """Hash relative names and bytes; symlinks are distinct from regular files."""
    if root.is_symlink():
        try:
            resolved = root.resolve(strict=True)
        except OSError:
            return None
        if not resolved.is_dir():
            return None
        root = resolved
    if not root.is_dir():
        return None
    digest = hashlib.sha256()
    for path in sorted(root.rglob("*")):
        rel = path.relative_to(root).as_posix().encode("utf-8")
        if path.is_symlink():
            digest.update(b"L\0" + rel + b"\0" + os.readlink(path).encode("utf-8") + b"\0")
        elif path.is_file():
            digest.update(b"F\0" + rel + b"\0")
            with path.open("rb") as stream:
                for block in iter(lambda: stream.read(1024 * 1024), b""):
                    digest.update(block)
            digest.update(b"\0")
        elif not path.is_dir():
            return None
    return digest.hexdigest()


def copy_previous(path, backup):
    """Copy the destination itself; a top-level symlink is copied as a symlink."""
    backup.parent.mkdir(parents=True, exist_ok=True)
    if path.is_symlink():
        # Keep a usable reference to the same target even though backup lives
        # outside the original parent (relative symlinks would otherwise drift).
        target = str(path.resolve(strict=False))
        os.symlink(target, str(backup), target_is_directory=True)
    elif path.is_dir():
        shutil.copytree(str(path), str(backup), symlinks=True)
    else:
        raise ValueError("existing destination is not a directory or symlink: {}".format(path))


def remove_path(path):
    if path.is_symlink() or path.is_file():
        path.unlink()
    elif path.exists():
        shutil.rmtree(str(path))


def safe_backup_root(value, source, destinations, discovery_roots):
    root = canonical_leaf(absolute_path(value))
    if root.exists() and (root.is_symlink() or not root.is_dir()):
        raise ValueError("backup directory must be a real directory: {}".format(root))
    protected = [source] + destinations + discovery_roots
    if any(overlaps(root, other) for other in protected):
        raise ValueError("backup directory overlaps source, destination, or skill discovery location: {}".format(root))
    return root


def codex_global_destination(home):
    canonical = canonical_leaf(home / ".agents" / "skills" / SKILL_NAME)
    codex_home = absolute_path(os.environ.get("CODEX_HOME", home / ".codex"))
    legacy = canonical_leaf(codex_home / "skills" / SKILL_NAME)
    canonical_exists = canonical.exists() or canonical.is_symlink()
    legacy_exists = legacy.exists() or legacy.is_symlink()
    if canonical == legacy:
        return canonical
    if canonical_exists and legacy_exists:
        raise ValueError("Codex skill exists in both {} and {}; choose one explicitly with --dest".format(canonical, legacy))
    if legacy_exists:
        return legacy
    return canonical


def destination_paths(agent, project, dest):
    home = Path.home().resolve()
    if dest is not None:
        return [(agent, canonical_leaf(absolute_path(dest)))]
    selected = ("codex", "claude") if agent == "both" else (agent,)
    if project:
        base = canonical_leaf(absolute_path(project)).resolve()
        roots = {
            "codex": base / ".agents" / "skills" / SKILL_NAME,
            "claude": base / ".claude" / "skills" / SKILL_NAME,
        }
        return [(name, canonical_leaf(roots[name])) for name in selected]
    result = []
    for name in selected:
        if name == "codex":
            result.append((name, canonical_leaf(codex_global_destination(home))))
        else:
            result.append((name, canonical_leaf(home / ".claude" / "skills" / SKILL_NAME)))
    return result


def discovery_roots(project):
    home = Path.home().resolve()
    roots = [home / ".agents" / "skills", home / ".claude" / "skills"]
    roots.append(home / ".codex" / "skills")
    codex_home = os.environ.get("CODEX_HOME")
    if codex_home:
        roots.append(absolute_path(codex_home) / "skills")
    if project:
        project_root = absolute_path(project).resolve(strict=False)
        roots.extend([project_root / ".agents" / "skills", project_root / ".claude" / "skills"])
    resolved = []
    for root in roots:
        root = absolute_path(root)
        if root.exists() or root.is_symlink():
            resolved.append(root.resolve(strict=False))
        else:
            resolved.append(canonical_leaf(root))
    return resolved


def lock_destinations(destinations):
    locks = []
    try:
        for _, dest in sorted(destinations, key=lambda item: str(item[1])):
            dest.parent.mkdir(parents=True, exist_ok=True)
            lock = dest.parent / ("." + dest.name + ".install-lock")
            try:
                lock.mkdir()
            except FileExistsError:
                raise ValueError("another installer may be running; lock exists: {}".format(lock))
            locks.append(lock)
            (lock / "owner").write_text("pid={}\n".format(os.getpid()), encoding="utf-8")
        return locks
    except Exception:
        for lock in reversed(locks):
            remove_path(lock)
        raise


def install(args):
    script_root = Path(__file__).resolve().parent.parent
    source_arg = absolute_path(args.source or script_root / "skills" / SKILL_NAME)
    if source_arg.is_symlink():
        raise ValueError("source must not be a symlink")
    source = source_arg.resolve(strict=True)
    files = source_files(source)
    source_digest = digest_tree(source)
    destinations = destination_paths(args.agent, args.project, args.dest)
    dest_paths = [dest for _, dest in destinations]
    discoveries = discovery_roots(args.project)

    resolved_destinations = [dest.resolve(strict=False) for dest in dest_paths]
    for index, dest in enumerate(dest_paths):
        if overlaps(source, dest):
            raise ValueError("source and destination overlap: {}".format(dest))
        try:
            actual = dest.resolve(strict=True)
        except OSError:
            actual = dest
        if overlaps(source, actual):
            raise ValueError("destination resolves into the source tree: {}".format(dest))
        for other in dest_paths[:index]:
            if (overlaps(dest, other)
                    or overlaps(resolved_destinations[index], resolved_destinations[dest_paths.index(other)])):
                raise ValueError("selected destinations alias or overlap; select one explicitly with --dest: {} and {}".format(other, dest))

    default_backup = Path(os.environ.get("XDG_DATA_HOME", str(Path.home() / ".local" / "share"))) / "pptxgengo" / "skill-backups"
    backup_root = safe_backup_root(args.backup_dir or default_backup, source, dest_paths, discoveries)

    plans = []
    for agent, dest in destinations:
        if dest.exists() and not dest.is_dir() and not dest.is_symlink():
            raise ValueError("destination exists and is not a directory or symlink: {}".format(dest))
        exists = dest.exists() or dest.is_symlink()
        current = digest_tree(dest) if exists else None
        plans.append({"agent": agent, "dest": dest, "current": current, "exists": exists, "changed": current != source_digest})

    if args.dry_run:
        for plan in plans:
            action = "no change" if not plan["changed"] else ("install" if not plan["exists"] else "update with backup")
            print("{}: {} {}".format(plan["agent"], action, plan["dest"]))
        if any(plan["changed"] and plan["exists"] for plan in plans):
            print("backup directory: {}".format(backup_root))
        print("source SHA-256: {}".format(source_digest))
        return

    locks = lock_destinations(destinations)
    staged = {}
    backups = {}
    held = {}
    installed = []
    committed = False
    try:
        # Recheck after acquiring every lock so the plan cannot race another installer.
        for plan in plans:
            dest = plan["dest"]
            exists_now = dest.exists() or dest.is_symlink()
            now = digest_tree(dest) if exists_now else None
            if exists_now != plan["exists"] or now != plan["current"]:
                raise ValueError("destination changed while waiting for its lock: {}".format(dest))
        for plan in plans:
            if not plan["changed"]:
                continue
            dest = plan["dest"]
            stage = dest.parent / ("." + dest.name + ".stage-" + uuid.uuid4().hex)
            staged[dest] = stage
            shutil.copytree(str(source), str(stage), symlinks=False)
            if digest_tree(stage) != source_digest:
                raise ValueError("source changed while staging; destination left untouched: {}".format(dest))
            if plan["exists"]:
                backup = backup_root / ("{}-{}-{}-{}".format(
                    SKILL_NAME, plan["agent"], time.strftime("%Y%m%dT%H%M%SZ", time.gmtime()), uuid.uuid4().hex[:8]))
                copy_previous(dest, backup)
                backups[dest] = backup
        for plan in plans:
            if not plan["changed"]:
                continue
            dest = plan["dest"]
            if plan["exists"]:
                old = dest.parent / ("." + dest.name + ".previous-" + uuid.uuid4().hex)
                os.replace(str(dest), str(old))
                held[dest] = old
            os.replace(str(staged[dest]), str(dest))
            installed.append(dest)
        committed = True
    except Exception:
        rollback_errors = []
        for dest in reversed(installed):
            try:
                remove_path(dest)
            except Exception as exc:
                rollback_errors.append("remove {}: {}".format(dest, exc))
        for dest, old in reversed(list(held.items())):
            try:
                if old.exists() or old.is_symlink():
                    os.replace(str(old), str(dest))
            except Exception as exc:
                rollback_errors.append("restore {} from {}: {}".format(dest, old, exc))
        if rollback_errors:
            raise RuntimeError("install failed and rollback needs attention: " + "; ".join(rollback_errors))
        raise
    finally:
        for stage in staged.values():
            remove_path(stage)
        for lock in reversed(locks):
            remove_path(lock)
    if committed:
        for old in held.values():
            try:
                remove_path(old)
            except Exception as exc:
                print("warning: installed successfully; temporary previous copy retained at {} ({})".format(old, exc), file=sys.stderr)
        for plan in plans:
            status = "unchanged" if not plan["changed"] else ("updated" if plan["exists"] else "installed")
            backup = backups.get(plan["dest"])
            print("{}: {} {}{}".format(plan["agent"], status, plan["dest"],
                  " (previous copy: {})".format(backup) if backup else ""))


def main():
    parser = argparse.ArgumentParser(description="Install/update the West Monroe presentation skill from a trusted checkout.")
    parser.add_argument("--agent", choices=("codex", "claude", "both"), default="codex")
    parser.add_argument("--project", help="project root; installs to its .agents/skills or .claude/skills folder")
    parser.add_argument("--source", help="skill folder; defaults to skills/{} beside this script".format(SKILL_NAME))
    parser.add_argument("--dest", help="explicit skill folder for one agent; does not create a second global copy")
    parser.add_argument("--dry-run", action="store_true", help="show install/update/no-op actions without writing")
    parser.add_argument("--backup-dir", help="store replaced skill copies here; defaults outside discovery roots")
    args = parser.parse_args()
    if args.project and args.dest:
        parser.error("--project and --dest cannot be combined")
    if args.dest and args.agent == "both":
        parser.error("--dest requires one agent (codex or claude)")
    try:
        install(args)
    except Exception as exc:
        print("install-skill: {}".format(exc), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
