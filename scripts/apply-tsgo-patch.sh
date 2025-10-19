#!/bin/bash

# apply tsgo patch
cd typescript-go

# Check if patch is already applied by trying to apply in reverse
if git apply --reverse --check ../__patches__/typescript-go.patch 2>/dev/null; then
  echo "Patch already applied, skipping..."
  exit 0
fi

git apply ../__patches__/typescript-go.patch