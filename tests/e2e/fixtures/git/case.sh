#!/usr/bin/env bash
set -eu
git init -q
echo hello > a
git add a
git commit -q -m "a"
echo world >> a
git add a
git commit -q -m "b"
git log --oneline
