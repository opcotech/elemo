#!/bin/bash

set -euo pipefail

if [ -n "$(aws s3api head-bucket --bucket elemo 2>&1 || true)" ]; then
  echo "Init localstack s3"
  awslocal s3 mb s3://elemo
fi
