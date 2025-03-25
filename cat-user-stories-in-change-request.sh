#!/bin/bash

change_request_file=$1

if [ -z "$change_request_file" ]; then
    echo "Usage: $0 <change_request_file>"
    exit 1
fi

# the change request file is a markdown file that contains user stories
# the user stories are listed in the file with the following format:
#
#  - title: Clear Search Filter
#    file: docs/user-stories/create-change-request-tui/10-clear-search-filter.md
#    content-hash: a24715fecc46d95bc34c5adcc473b9d7
#  - title: Auto-Focus First Result After Search
#    file: docs/user-stories/create-change-request-tui/11-auto-focus-first-result-after-search.md
#    content-hash: b60d0147a3881dbe84df5a1933447e22
#
# Extract user stories from the change request file
grep -E 'file:' "$change_request_file" | cut -d':' -f2 | while read -r user_story_file; do
    echo "[//]: # ($user_story_file)"
    cat "$user_story_file"
done
