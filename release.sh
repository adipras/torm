#!/bin/bash

# ========================================
# release.sh — Rilis versi baru TORM (idempotent)
# ========================================

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
  echo "❌ Please provide version. Example: ./release.sh v0.1.0"
  exit 1
fi

echo "📦 Module: github.com/adipras/torm"
echo "🧹 Running go mod tidy..."
go mod tidy

# Check if tag exists on GitHub remote
echo "🔍 Checking if tag $VERSION exists on GitHub..."
if git ls-remote --tags origin | grep "refs/tags/$VERSION" > /dev/null; then
  echo "✅ Tag $VERSION already exists on GitHub. Skipping tagging."
else
  echo "🏷️  Tagging version $VERSION ..."
  git add .
  git commit -m "Release $VERSION" || echo "ℹ️  No changes to commit."
  git tag $VERSION
fi

echo "🚀 Pushing to GitHub..."
git push origin main --tags

# Check GitLab host reachable
GITLAB_HOST="gitlab-cloud.uii.ac.id"
echo "🔍 Checking GitLab access..."
if ping -c 1 $GITLAB_HOST &> /dev/null
then
  echo "✅ GitLab reachable. Pushing to GitLab..."
  git push https://$GITLAB_HOST/adipras/torm.git main --tags
else
  echo "⚠️  GitLab unreachable (VPN?). Skipping GitLab push."
fi

echo "🎉 Release $VERSION done!"
