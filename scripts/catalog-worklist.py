#!/usr/bin/env python3
"""Derive a classification worklist from slide occurrences and layout decisions.

This tool only reads its inputs and creates a new JSON output. It refuses to
replace an existing output so each generated worklist remains reviewable.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import tempfile


MEMBER_RE = re.compile(r"^(.+):(\d+)$")
SHARED_STATUS = "reviewed_arrangement_classify_once_preserve_variants"
SHARED_ACTION = "define_role_quality_and_slots"
DISTINCT_STATUS = "reviewed_distinct_item"
DISTINCT_ACTION = "retain_separate_unless_new_evidence"
SINGLETON_STATUS = "not_yet_visually_classified"
SINGLETON_ACTION = "broader_dedup_then_classification"


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def read_occurrences(path):
    raw = path.read_bytes()
    sources = {}
    slides = {}
    with path.open("r", encoding="utf-8") as f:
        for line_number, line in enumerate(f, 1):
            if not line.strip():
                continue
            try:
                row = json.loads(line)
            except json.JSONDecodeError as exc:
                raise ValueError(f"Invalid JSONL at {path}:{line_number}: {exc}") from exc
            source_id = row.get("source_id")
            source_hash = row.get("source_sha256")
            if not source_id or not isinstance(source_hash, str) or not re.fullmatch(r'[0-9a-f]{64}', source_hash):
                raise ValueError(f"Missing source identity at {path}:{line_number}")
            previous_hash = sources.setdefault(source_id, source_hash)
            if previous_hash != source_hash:
                raise ValueError(f"Conflicting source hashes for {source_id} in occurrences")
            if row.get("occurrence_type") != "slide":
                continue
            try:
                slide_number = int(row["slide_number"])
            except (KeyError, TypeError, ValueError) as exc:
                raise ValueError(f"Invalid slide number at {path}:{line_number}") from exc
            if slide_number < 1:
                raise ValueError(f"Slide number must be positive at {path}:{line_number}")
            member = f"{source_id}:{slide_number:03d}"
            if member in slides:
                raise ValueError(f"Duplicate slide occurrence: {member}")
            slides[member] = {
                "source_id": source_id,
                "source_sha256": source_hash,
                "slide_number": slide_number,
            }
    if not slides:
        raise ValueError("Occurrences input contains no slide records")
    return raw, sources, slides


def member_source(member):
    match = MEMBER_RE.fullmatch(member) if isinstance(member, str) else None
    if not match:
        raise ValueError(f"Invalid decision member ID: {member!r}; expected source_id:slide_number")
    try:
        slide_number = int(match.group(2))
    except ValueError as exc:
        raise ValueError(f"Invalid decision member ID: {member!r}") from exc
    return f"{match.group(1)}:{slide_number:03d}", match.group(1)


def derive_worklist(occurrences_path, decisions_path):
    occurrence_bytes, source_hashes, slides = read_occurrences(occurrences_path)
    decisions_bytes = decisions_path.read_bytes()
    try:
        decisions = json.loads(decisions_bytes)
    except json.JSONDecodeError as exc:
        raise ValueError(f"Invalid decisions JSON: {exc}") from exc

    expected_hashes = decisions.get("source_hashes", {})
    if not isinstance(expected_hashes, dict):
        raise ValueError("Decisions source_hashes must be an object")
    if expected_hashes != source_hashes:
        raise ValueError("Decisions and occurrences must have identical full source hashes")
    for source_id, source_hash in expected_hashes.items():
        observed = source_hashes.get(source_id)
        if observed is None:
            raise ValueError(f"Decision source has no occurrences: {source_id}")
        if observed != source_hash:
            raise ValueError(f"Stale source hash for {source_id}: decisions do not match occurrences")

    groups = decisions.get("groups")
    if not isinstance(groups, list):
        raise ValueError("Decisions must contain a groups array")
    group_ids = set()
    assigned_members = set()
    work_units = []
    reviewed_members = set()
    shared_count = 0
    distinct_count = 0

    for index, group in enumerate(groups, 1):
        group_id = group.get("id")
        if not isinstance(group_id, str) or not group_id:
            raise ValueError(f"Decision group {index} is missing an ID")
        if group_id in group_ids:
            raise ValueError(f"Duplicate decision group ID: {group_id}")
        group_ids.add(group_id)

        members = group.get("members")
        if not isinstance(members, list) or not members:
            raise ValueError(f"Decision group {group_id} must have a nonempty members list")
        normalized_members = []
        local_members = set()
        for raw_member in members:
            member, source_id = member_source(raw_member)
            if member in local_members:
                raise ValueError(f"Duplicate member {member} in decision group {group_id}")
            local_members.add(member)
            if member in assigned_members:
                raise ValueError(f"Decision member assigned to multiple groups: {member}")
            occurrence = slides.get(member)
            if occurrence is None:
                raise ValueError(f"Decision member is missing from slide occurrences: {member}")
            expected_hash = expected_hashes.get(source_id)
            if expected_hash is None:
                raise ValueError(f"Decision member {member} has no full source hash in decisions")
            if occurrence["source_sha256"] != expected_hash:
                raise ValueError(f"Source hash mismatch for decision member {member}")
            normalized_members.append(member)

        representative, representative_source = member_source(group.get("representative"))
        if representative not in local_members:
            raise ValueError(f"Representative {representative} is not a member of {group_id}")
        if representative not in slides:
            raise ValueError(f"Representative is missing from slide occurrences: {representative}")
        representative_hash = expected_hashes.get(representative_source)
        if representative_hash != slides[representative]["source_sha256"]:
            raise ValueError(f"Source hash mismatch for representative {representative}")

        disposition = group.get("disposition")
        if disposition not in {'shared_layout_family', 'distinct'}:
            raise ValueError(f"Decision group {group_id} has invalid disposition")
        if disposition == "shared_layout_family":
            status, next_action = SHARED_STATUS, SHARED_ACTION
            shared_count += 1
        else:
            status, next_action = DISTINCT_STATUS, DISTINCT_ACTION
            distinct_count += 1

        assigned_members.update(local_members)
        reviewed_members.update(local_members)
        work_units.append({
            "id": group_id,
            "representative": representative,
            "members": normalized_members,
            "status": status,
            "next_action": next_action,
        })

    # Keep occurrence/source order stable as families are removed from the queue.
    unreviewed = [member for member in slides if member not in assigned_members]
    for member in unreviewed:
        if f"unclassified:{member}" in group_ids:
            raise ValueError(f"Decision ID collides with singleton work unit: {member}")
        work_units.append({
            "id": f"unclassified:{member}",
            "representative": member,
            "members": [member],
            "status": SINGLETON_STATUS,
            "next_action": SINGLETON_ACTION,
        })

    pair_decisions = decisions.get("pair_decisions", [])
    if not isinstance(pair_decisions, list):
        raise ValueError("Decisions pair_decisions must be an array when present")
    summary = {
        "source_slides": len(slides),
        "reviewed_previews": len(reviewed_members),
        "shared_visual_families": shared_count,
        "reviewed_distinct_items": distinct_count,
        "classification_work_units": len(work_units),
        "repeated_classification_units_avoided": len(reviewed_members) - len(groups),
        "unreviewed_singletons": len(unreviewed),
        "queue_pairs_reviewed": len(pair_decisions),
        "pairs_kept_separate": sum(p.get("decision") == "keep_separate" for p in pair_decisions),
        "final_unique_layout_count": None,
    }
    decision_summary = decisions.get("summary", {})
    if decision_summary is None:
        decision_summary = {}
    if not isinstance(decision_summary, dict):
        raise ValueError("Decisions summary must be an object when present")
    for key, observed in summary.items():
        if key in decision_summary and decision_summary[key] != observed:
            raise ValueError(
                f"Stale decisions summary field {key}: expected {observed!r}, found {decision_summary[key]!r}"
            )

    return {
        "schema_version": 1,
        "source_hashes": dict(sorted(source_hashes.items())),
        "input_hashes": {
            "occurrences_sha256": sha256(occurrence_bytes),
            "decisions_sha256": sha256(decisions_bytes),
        },
        "summary": summary,
        "work_units": work_units,
    }


def write_exclusive(path, content):
    path.parent.mkdir(parents=True, exist_ok=True)
    if path.exists():
        raise FileExistsError(f"Output already exists; refusing to overwrite: {path}")
    fd, temp_name = tempfile.mkstemp(prefix=f".{path.name}.", dir=path.parent)
    temp = Path(temp_name)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            f.write(content)
            f.flush()
            os.fsync(f.fileno())
        os.link(temp, path)
    finally:
        temp.unlink(missing_ok=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--occurrences", type=Path, required=True, help="Slide occurrence JSONL")
    parser.add_argument("--decisions", type=Path, required=True, help="Layout decisions JSON")
    parser.add_argument("--out", type=Path, required=True, help="New worklist JSON path (must not exist)")
    args = parser.parse_args()
    try:
        result = derive_worklist(args.occurrences, args.decisions)
        write_exclusive(args.out, json.dumps(result, indent=2, ensure_ascii=False) + "\n")
    except (OSError, ValueError) as exc:
        parser.error(str(exc))
    print(json.dumps(result["summary"], sort_keys=True))


if __name__ == "__main__":
    main()
