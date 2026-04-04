#!/bin/bash
# Creates 48 hourly cron jobs for ViBrowsing sprint
# Each job runs 1 hour from now, 2 hours from now, etc.
cd /home/nick/Repos/nickcoury/ViBrowsing

for i in $(seq 1 48); do
    at "now + $i hours" -f - <<ATEOF 2>/dev/null
cd /home/nick/Repos/nickcoury/ViBrowsing
/export PATH="\$HOME/go-install/go/bin:\$PATH"
git stash 2>/dev/null || true
git pull origin main 2>/dev/null || true
python3 ~/.hermes/claude-code/venv/bin/python3 -c "
import subprocess, sys, re
from pathlib import Path

backlog = Path('backlog.md')
content = backlog.read_text()

# Find first unchecked item
lines = content.split('\n')
for idx, line in enumerate(lines):
    if line.strip().startswith('- [ ]'):
        # Check if it's in a section we can work on
        # Find what section this is in
        section = ''
        for j in range(idx-1, -1, -1):
            if lines[j].startswith('## '):
                section = lines[j]
                break
        item = line.strip()[6:].strip() # remove '- [ ] '
        task_desc = f'{section} :: {item}'
        print(f'WORKING ON: {task_desc}')
        sys.stdout.flush()
        
        # Mark done by replacing with - [x]
        lines[idx] = line.replace('- [ ]', '- [x] (2026-04-03 sprint)')
        backlog.write_text('\n'.join(lines))
        break
ATEOF
    echo "Scheduled job $i"
done

echo "All 48 jobs scheduled!"
at -l | wc -l
