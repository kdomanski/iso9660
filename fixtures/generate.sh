#!/usr/bin/env bash

set -eu
set -o pipefail

touch -d '2018-07-25 22:01:02' test.iso_source
mkisofs -V my-vol-id -publisher gopher -volset test-volset-id -preparer "$(id -un)" -o test.iso test.iso_source
