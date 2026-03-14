#!/usr/bin/env python3
"""Gemini codebase analysis — ATLAS uses this before generating tasks.

Sends the project structure and key files to Gemini 2.5 Pro for
architecture analysis and recommendations.

Usage: python gemini-analyze.py [--focus <area>]
Requires: GEMINI_API_KEY environment variable
"""

import os
import sys
import subprocess
import argparse

def get_project_root():
    result = subprocess.run(
        ["git", "rev-parse", "--show-toplevel"],
        capture_output=True, text=True, check=True
    )
    return result.stdout.strip()

def get_file_tree(root, max_depth=3):
    result = subprocess.run(
        ["find", root, "-maxdepth", str(max_depth),
         "-not", "-path", "*/node_modules/*",
         "-not", "-path", "*/.git/*",
         "-not", "-path", "*/_colony/vendor/*",
         "-not", "-path", "*/.worktrees/*",
         "-not", "-path", "*/__pycache__/*"],
        capture_output=True, text=True
    )
    return result.stdout.strip()

def get_recent_changes():
    result = subprocess.run(
        ["git", "log", "--oneline", "--stat", "-10"],
        capture_output=True, text=True
    )
    return result.stdout.strip()

def read_key_files(root):
    key_files = ["CLAUDE.md", "_colony/GOALS.md", "_colony/ROADMAP.md"]
    contents = {}
    for f in key_files:
        path = os.path.join(root, f)
        if os.path.exists(path):
            with open(path) as fh:
                contents[f] = fh.read()
    return contents

def analyze(focus=None):
    api_key = os.environ.get("GEMINI_API_KEY")
    if not api_key:
        print("ERROR: GEMINI_API_KEY not set", file=sys.stderr)
        sys.exit(1)

    try:
        import google.generativeai as genai
    except ImportError:
        print("ERROR: google-generativeai not installed. Run: pip install google-generativeai", file=sys.stderr)
        sys.exit(1)

    genai.configure(api_key=api_key)
    model = genai.GenerativeModel("gemini-2.5-pro")

    root = get_project_root()
    tree = get_file_tree(root)
    changes = get_recent_changes()
    key_files = read_key_files(root)

    prompt = f"""You are analyzing a software project for an AI development colony.

## Project Structure
```
{tree}
```

## Recent Changes
```
{changes}
```

## Key Files
"""
    for name, content in key_files.items():
        prompt += f"\n### {name}\n```\n{content}\n```\n"

    if focus:
        prompt += f"\n## Focus Area\nAnalyze specifically: {focus}\n"

    prompt += """
## Your Task
Provide:
1. Architecture overview (what patterns are used, what the project does)
2. Current state assessment (what's built, what's missing)
3. Dependency map (what depends on what)
4. Recommended next tasks (ordered by priority and dependency)
5. Potential risks or technical debt
6. Files that should NOT be modified concurrently by two developers

Be specific. Reference actual file paths and function names.
"""

    response = model.generate_content(prompt)
    print(response.text)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Gemini codebase analysis for ATLAS")
    parser.add_argument("--focus", help="Specific area to analyze")
    args = parser.parse_args()
    analyze(focus=args.focus)
