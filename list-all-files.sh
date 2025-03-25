#!/bin/bash

# Use git to list all files excluding those in .gitignore
git ls-files -c -o --exclude-standard
