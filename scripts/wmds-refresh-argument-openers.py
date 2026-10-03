#!/usr/bin/env python3
"""Create source/changed-content pairs for the 12 added Argument/Openers designs."""
import argparse
import copy
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def run(cli, *args):
    subprocess.run([str(cli), *args], cwd=ROOT, check=True)


def alternate(slide):
    result = copy.deepcopy(slide)
    result["id"] = "alternate-" + slide["id"]
    slots = result["values"]["slots"]
    key = result["template"]
    for name, value in list(slots.items()):
        if isinstance(value, str):
            slots[name] = value.replace("Northfield", "Lakeview")
    if key.startswith("agenda/"):
        slots["title"] = "Controls and close working session"
        for name in slots:
            if name.endswith(".item02.t"):
                slots[name] = "Where approvals stall"
            elif name.endswith(".item03.t"):
                slots[name] = "Target controls and cadence"
            elif name.endswith(".item04.t"):
                slots[name] = "Pilot scope and owners"
            elif name.endswith(".body"):
                slots[name] = "An agreed pilot scope and a named owner for each decision."
    elif key.startswith("key-message/"):
        slots["title"] = "Close the books in [[6 days]] instead of 14"
        slots["eyebrow"] = "Illustrative finance pilot"
        for name, value in list(slots.items()):
            if value == "$4.2M":
                slots[name] = "$3.8M"
            elif value == "12 → 5 days":
                slots[name] = "14 → 6 days"
            elif isinstance(value, str) and "controller signs" in value:
                slots[name] = "The controller signs off on day six, with fewer late exceptions."
            elif isinstance(value, str) and "standardize reconciliations" in value:
                slots[name] = "We align reconciliations, automate matching and review early."
    elif key.startswith("comparison/"):
        slots.update({"node03.metric.value": "Day 120",
                      "node03.metric.label": "Your team owns the next release.",
                      "node01.text": "Finance, data and engineering work as one team, with decisions grounded in shared evidence.",
                      "node02.text": "We build alongside your owners from discovery through release, then transfer the operating knowledge.",
                      "node09.text": "Your team can run, support and improve the platform.",
                      "node11.text": "Business and technical owners decide together.",
                      "node13.text": "Security and data controls shape delivery.",
                      "node15.text": "Track lead time, adoption and service cost."})
    elif key.startswith("objective/"):
        slots.update({"node01.text": "Why evidence changes team performance",
                      "node02.items.item01.title": "Smaller slices move faster",
                      "node02.items.item01.text": "Teams use drafts and prototypes to shorten feedback cycles and make decisions earlier.",
                      "node02.items.item02.title": "Clear intent prevents rework",
                      "node02.items.item02.text": "Evidence, constraints and acceptance criteria keep fast-moving work aligned with its purpose.",
                      "node02.items.item03.title": "Shared reviews create confidence",
                      "node02.items.item03.text": "Visible evidence and named approvers help teams repeat useful results across delivery.",
                      "node03.title": "Coach the team through each decision",
                      "node03.body.item01.bullets.item01": "Brief framing, then team work",
                      "node03.body.item01.bullets.item04": "Named approval at each gate"})
    elif key.startswith("session/"):
        slots["title"] = "Build together, review evidence, decide next steps"
        slots["node02.steps.item01.label"] = "20 min"
        slots["node02.steps.item04.label"] = "20 min"
        slots["node02.steps.item02.state"] = "current"
        slots["node02.steps.item03.state"] = "next"
        slots["node02.steps.item01.title"] = "Agree the decision"
        slots["node02.steps.item04.title"] = "Name the next actions"
        if "node04.text" in slots and slots["node04.text"]:
            slots["node04.text"] = "Bring current evidence, make active decisions and share constraints early."
        if "node03.title" in slots:
            slots["node03.title"] = "Coach each team through its decisions"
        if "node07.title" in slots:
            slots["node07.title"] = "Choose a useful opportunity"
            slots["node07.body.item01.p"] = "A bounded problem to solve"
            slots["node12.body.item01.p"] = "Ownership and clear controls"
    elif key.startswith("pillars/"):
        slots.update({"node01.title": "Value led",
                      "node01.body.item01.bullets.item01": "Rank work by its business impact.",
                      "node04.text": "Invest where the expected return is clear and measurable.",
                      "node05.title": "Fit for purpose",
                      "node08.text": "Keep stable capabilities and change what limits progress.",
                      "node09.title": "Evidence driven",
                      "node09.body.item01.bullets.item01": "Use agents to map code and dependencies.",
                      "node12.text": "Owners review evidence before the next implementation decision.",
                      "node13.title": "Build capability",
                      "node16.text": "Grow delivery capacity and transfer the knowledge to run the result."})
    elif key.startswith("transformation/"):
        slots.update({"node03.text": "Reviews wait until period end, and unresolved exceptions extend the close to 14 days.",
                      "node05.text": "Review every day", "node06.text": "Check entries before the close starts.",
                      "node09.text": "Shared matching rules", "node10.text": "Match balances daily and route exceptions.",
                      "node13.text": "One close workspace", "node14.text": "See tasks, owners and blockers together.",
                      "node16.text": "Finance and operations use different definitions, so reviews reopen settled decisions.",
                      "node18.text": "Agreed data model", "node19.text": "Use the same measures in both teams.",
                      "node22.text": "Accountable owners", "node23.text": "Name an owner for each key metric.",
                      "node26.text": "Reusable definitions", "node27.text": "Agree terms before the next release."})
    for name, keys in result["values"]["keys"].items():
        result["values"]["keys"][name] = [f"variant-{i+1:02d}" for i in range(len(keys))]
    nav = result["values"].get("nav")
    if nav:
        for i, item in enumerate(nav["items"]):
            item["key"] = f"section-{i+1:02d}"
        nav["active"] = nav["items"][-1]["key"]
    return result


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--cli", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    cli = args.cli.resolve()
    out = args.out.resolve()
    out.mkdir(parents=True, exist_ok=False)
    inventory = json.loads((ROOT / "planning/wm-design-contracts/v2/source-update-2026-10-02.json").read_text())
    keys = [t["key"] for t in inventory["templates"] if t["change"] == "added" and t["family"] in ("argument", "openers")]
    common = ["--bundle", "library/wm-design-system/v2", "--engine", "wmds-go-foundation.v2"]
    run(cli, "library-reference", *common, "--year", "2026", "--template-keys", ",".join(keys), "--out", str(out / "source"))
    input_doc = json.loads((out / "source/template-content.json").read_text())
    input_doc["slides"] = [alternate(s) for s in input_doc["slides"]]
    (out / "bound-content.json").write_text(json.dumps(input_doc, indent=2) + "\n")
    run(cli, "template", *common, "--spec", str(out / "bound-content.json"), "--out", str(out / "alternate"))
    source = json.loads((out / "source/compiled-document.json").read_text())
    changed = json.loads((out / "alternate/compiled-document.json").read_text())
    combined = copy.deepcopy(source)
    combined["slides"] = []
    receipts = []
    for first, second, bound in zip(source["slides"], changed["slides"], input_doc["slides"]):
        key = bound["template"]
        first["id"] = "source-" + key.replace("/", "-")
        second["id"] = "alternate-" + key.replace("/", "-")
        for slide in (first, second):
            slide["template_binding"]["slide_id"] = slide["id"]
        combined["slides"].extend([first, second])
        original_values = json.loads((out / "source/template-content.json").read_text())["slides"][len(receipts)]["values"]
        changed_slots = [k for k, v in bound["values"]["slots"].items() if original_values["slots"][k] != v]
        receipts.append({"template": key, "generation": "source_and_alternate_generated",
                         "changed_slots": changed_slots, "native_review": "pending"})
    (out / "combined.foundation.json").write_text(json.dumps(combined, indent=2) + "\n")
    (out / "generation-receipts.json").write_text(json.dumps(receipts, indent=2) + "\n")


if __name__ == "__main__":
    main()
