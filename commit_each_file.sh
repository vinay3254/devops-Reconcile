#!/usr/bin/env bash
set -e

cd /mnt/c/Users/Admin/Desktop/devops

rm -f .git/index.lock

echo "=== Processing untracked and modified files ==="

readarray -t UNTRACKED < <(git ls-files --others --exclude-standard)

for FILE in "${UNTRACKED[@]}"; do
    if [ -f "$FILE" ]; then
        echo "Committing file: $FILE"
        git add "$FILE"
        git commit -m "feat: add $FILE" || true
    fi
done

readarray -t MODIFIED < <(git ls-files --modified)

for FILE in "${MODIFIED[@]}"; do
    if [ -f "$FILE" ]; then
        echo "Committing modified file: $FILE"
        git add "$FILE"
        git commit -m "refactor: update $FILE" || true
    fi
done

echo "=== Processing Complete ==="
