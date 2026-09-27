#!/usr/bin/env python3
"""Run repeatability and rejection checks for the proposal stress assembly."""

import argparse
import hashlib
import json
import subprocess
import tempfile
import time
from pathlib import Path


REPO = Path(__file__).resolve().parents[1]
CONFIG = REPO / "library/proposal/stress/assembly.json"
INDEX = REPO / "samples/proposal-authoring/capacity-v6.sqlite"


def run(cmd):
    start = time.perf_counter()
    p = subprocess.run(cmd, cwd=REPO, text=True, capture_output=True)
    return p, time.perf_counter() - start


def digest_specs(root):
    files = sorted(root.glob("slides/*/spec.json"))
    return {str(f.relative_to(root)): hashlib.sha256(f.read_bytes()).hexdigest() for f in files}


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--cli", default="/tmp/pptxlib-wave4-integration", help="pptxlib executable")
    ap.add_argument("--index", type=Path, default=INDEX, help="library SQLite index")
    ap.add_argument("--out", type=Path, default=Path("/tmp/pptxlib-proposal-stress-qualification"), help="new output directory")
    args = ap.parse_args()
    cli = str(Path(args.cli).resolve())
    out = args.out.resolve()
    if out.exists():
        raise SystemExit(f"refusing to overwrite existing output directory: {out}")
    out.mkdir(parents=True)
    index = args.index.resolve()
    report = {"schema": "pptxgengo.proposal-stress-qualification.v1", "cli": cli, "config": str(CONFIG), "index": str(index), "status": "running", "native_visual_qualified": False, "checks": []}

    def save_report():
        report["diagnostics_json"] = str(out / "diagnostics.json")
        (out / "diagnostics.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
    save_report()
    try:
        report.update({"cli_sha256": hashlib.sha256(Path(cli).read_bytes()).hexdigest(), "config_sha256": hashlib.sha256(CONFIG.read_bytes()).hexdigest(), "index_sha256": hashlib.sha256(index.read_bytes()).hexdigest()})
        save_report()

        help_result, _ = run([cli, "assemble", "--help"])
        report["cli_help_exit"] = help_result.returncode
        report["cli_help"] = (help_result.stdout + help_result.stderr).strip()

        repeat = []
        for n in (1, 2):
            target = out / f"assemble-{n}"
            p, seconds = run([cli, "assemble", "--root", str(REPO), "--index", str(index), "--config", str(CONFIG), "--out", str(target), "--allow-unqualified"])
            report.setdefault("assembly_commands", []).append({"run": n, "exit_code": p.returncode, "seconds": seconds, "stdout": p.stdout, "stderr": p.stderr})
            save_report()
            if p.returncode:
                raise RuntimeError(f"assemble {n} failed ({p.returncode}): {p.stderr}")
            hashes = digest_specs(target)
            repeat.append({"seconds": seconds, "spec_count": len(hashes), "spec_hashes": hashes,
                           "combined_spec_sha256": hashlib.sha256((target / "spec.json").read_bytes()).hexdigest()})
            report["repeat_assemblies"] = repeat
            save_report()
        if any(r["spec_count"] != 13 for r in repeat):
            raise RuntimeError(f"expected 13 assembled specs, found {len(repeat[0]['spec_hashes'])}")
        if (repeat[0]["spec_hashes"] != repeat[1]["spec_hashes"] or
                repeat[0]["combined_spec_sha256"] != repeat[1]["combined_spec_sha256"]):
            raise RuntimeError("repeat assembly produced different spec hashes")
        report["repeat_assemblies"] = repeat
        report["repeatable"] = True

        with tempfile.TemporaryDirectory(prefix="proposal-stress-negative-") as scratch:
            scratch = Path(scratch)

            def check(name, slide_id, expected, style=None, array_mutator=None):
                config = json.loads(CONFIG.read_text())
                config["narrative_path"] = str(REPO / "library/proposal/narrative.json")
                for item in config["slides"]:
                    item["values_path"] = str((CONFIG.parent / item["values_path"]).resolve())
                slide = next(s for s in config["slides"] if s["slide_id"] == slide_id)
                original = json.loads((CONFIG.parent / slide["values_path"]).read_text())
                if array_mutator:
                    original["slots"], target_slot = array_mutator(original["slots"])
                else:
                    target_slot = None
                if style is not None:
                    original["style_variant"] = style
                valpath = scratch / f"{name}.json"
                valpath.write_text(json.dumps(original, ensure_ascii=False, indent=2) + "\n")
                slide["values_path"] = str(valpath)
                cfgpath = scratch / f"{name}-assembly.json"
                cfgpath.write_text(json.dumps(config, ensure_ascii=False, indent=2) + "\n")
                target = out / f"negative-{name}"
                p, seconds = run([cli, "assemble", "--root", str(REPO), "--index", str(index), "--config", str(cfgpath), "--out", str(target), "--allow-unqualified"])
                message = (p.stdout + p.stderr).strip()
                success = p.returncode != 0 and expected in message and not target.exists()
                report["checks"].append({"name": name, "slide_id": slide_id, "expected_error": expected, "exit_code": p.returncode, "seconds": seconds, "diagnostic": message, "no_output_published": not target.exists(), "passed": success, "slot": target_slot})
                save_report()
                if not success:
                    raise RuntimeError(f"negative check {name} failed: {message}; output_exists={target.exists()}")

            def seven_steps(slots):
                key = next(k for k in slots if k.endswith("/request/steps"))
                slots[key] = slots[key] + ["Record a seventh step"]
                return slots, key

            check("seven-steps", "parallel-delivery-paths", "count outside [4,6]", array_mutator=seven_steps)

            inspect, _ = run([cli, "inspect", "--root", str(REPO), "--index", str(index), "--id", "wm/workflow-matrix"])
            if inspect.returncode:
                raise RuntimeError(inspect.stderr)
            contract = json.loads(inspect.stdout)["contract"]
            bounded = next(s for s in contract["composition"]["slots"] if s.get("value_type") == "string" and s.get("max_chars", 0) > 0)
            char_limit_value = "x" * (bounded["max_chars"] + 1)
            def over_cap(slots):
                slots[bounded["name"]] = char_limit_value
                return slots, bounded["name"]
            check("one-char-over-cap", "situation-and-friction", "exceeds bounded copy", array_mutator=over_cap)

            def unknown_slot(slots):
                slots["deliberately-unknown-slot"] = "unexpected"
                return slots, "deliberately-unknown-slot"
            check("unknown-slot", "situation-and-friction", "unknown slot deliberately-unknown-slot", array_mutator=unknown_slot)
            check("unknown-style", "situation-and-friction", "unknown style variant deliberately-unknown-style", style="deliberately-unknown-style")

        report["status"] = "passed"
        save_report()
        print(json.dumps({"out": str(out), "repeatable": report["repeatable"], "assemble_seconds": [r["seconds"] for r in repeat], "spec_count": repeat[0]["spec_count"], "negative_checks_passed": len(report["checks"]), "diagnostics": str(out / "diagnostics.json")}, indent=2))
    except Exception as exc:
        report["status"] = "failed"
        report["error"] = str(exc)
        raise
    finally:
        save_report()



if __name__ == "__main__":
    main()
