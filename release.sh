#!/usr/bin/env bash

# -----------------------------------------
# TORM Release Script
# Usage:
#   chmod +x release.sh
#   ./release.sh v0.1.0
# -----------------------------------------

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
  echo "❌ Please provide a version tag. Example: ./release.sh v0.1.0"
  exit 1
fi

# 1. Pastikan working directory bersih
if [ -n "$(git status --porcelain)" ]; then
  echo "❌ Uncommitted changes found. Please commit or stash first."
  git status
  exit 1
fi

# 2. Periksa go.mod module path
MODULE=$(grep "^module " go.mod | awk '{print $2}')
echo "📦 Module: $MODULE"

# 3. Jalankan go mod tidy
echo "🧹 Running go mod tidy..."
go mod tidy

# 4. Commit perubahan (kalau ada perubahan di go.mod/go.sum)
if [ -n "$(git status --porcelain)" ]; then
  git add go.mod go.sum
  git commit -m "chore: tidy go.mod before release $VERSION"
fi

# 5. Buat dan push tag
echo "🏷️  Tagging version $VERSION..."
git tag $VERSION
git push origin $VERSION

echo "✅ Release $VERSION pushed successfully!"

echo "🔗 Next steps:"
echo "👉 Check your release at https://github.com/adipras/torm/releases"
echo "👉 Visit https://pkg.go.dev/$MODULE and ensure your new version is indexed."
echo "👉 Test: go get $MODULE@$VERSION"

echo "🎉 Done!"
