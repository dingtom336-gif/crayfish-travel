#!/bin/bash
# Annual lunar calendar verification script.
# Cross-validates 6tail/lunar-go output against chinese-calendar-golang
# for winter peak dates (La Yue 23 ~ Zheng Yue 15) across 2026-2035.
#
# Run: bash scripts/lunar-verify.sh
# Requires: backend Go tests to be built first.

set -euo pipefail

echo "Running lunar calendar cross-validation..."
cd "$(dirname "$0")/../backend"
go test ./internal/nlp/ -run TestLunarCrossValidation -v -count=1
echo "Lunar verification complete."
