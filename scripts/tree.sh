#!/usr/bin/bash

# tree view all files and folders
find . -path "./.git" -prune -o -print |
awk -F/ '
{
    indent = ""
    for (i = 2; i < NF; i++) indent = indent "|   "
    print indent "├── " $NF
}'