#!/usr/bin/env python3
"""Gemini code review — AUDIT uses this for cross-validation.

Sends a branch diff to Gemini 2.5 Pro for independent review.

Usage: python gemini-review.py <branch-name>
Requires: GEMINI_API_KEY environment variable
"""

import os
import sys
import subprocess

def get_diff(branch):
    result = subprocess.run(
        ["git", "diff", f"main...{branch}"],
        capture_output=True, text=True, check=True
    )
    return result.stdout.strip()

def get_task_file(branch):
    """Try to find the task file for this branch."""
    # Extract task number from branch name (e.g., task/001 → 001)
    parts = branch.split("/")
    if len(parts) == 2:
        task_num = parts[1]
        for dir_name in ["review", "active", "queue", "done"]:
            path = f"_colony/{dir_name}/TASK-{task_num}.md"
            if os.path.exists(path):
                with open(path) as f:
                    return f.read()
    return None

def review(branch):
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

    diff = get_diff(branch)
    if not diff:
        print("No diff found — branch may be up to date with main")
        sys.exit(0)

    task_content = get_task_file(branch)

    prompt = f"""You are a senior code reviewer. Review this diff thoroughly.

## Branch: {branch}
"""

    if task_content:
        prompt += f"""
## Task Specification
```markdown
{task_content}
```
"""

    prompt += f"""
## Diff
```diff
{diff}
```

## Review Checklist
For each item, state PASS or FAIL with explanation:

1. **Tests:** Do tests exist? Do they cover the acceptance criteria?
2. **Logic:** Any logic errors, off-by-one bugs, race conditions?
3. **Security:** Any injection, XSS, SSRF, auth bypass, path traversal, hardcoded secrets?
4. **Architecture:** Does the code follow established patterns?
5. **Scope:** Is there any scope creep beyond the task?
6. **Error handling:** Are errors handled properly?
7. **Dependencies:** Any broken imports or missing dependencies?
8. **Readability:** Clear names, reasonable function length, no dead code?

## Verdict
End with one of:
- **APPROVE** — Code is ready to merge
- **REJECT** — Must fix before merge (list specific issues)
"""

    response = model.generate_content(prompt)
    print(response.text)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <branch-name>", file=sys.stderr)
        sys.exit(1)
    review(sys.argv[1])
