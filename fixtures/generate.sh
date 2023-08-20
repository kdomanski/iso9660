#!/usr/bin/env bash

set -eu
set -o pipefail

touch -d '2018-07-25 22:01:02' test.iso_source
chmod 0640 test.iso_source/dir1/lorem_ipsum.txt
mkisofs -V my-vol-id -publisher gopher -volset test-volset-id -preparer "$(id -un)" -o test.iso test.iso_source

ln -s /usr/share/some-random-directory/even-deeper-path/symlink-target test.iso_source/this-is-a-symlink
mkisofs -R -V my-vol-id -publisher gopher -volset test-volset-id -preparer "$(id -un)" -o test_rockridge.iso test.iso_source
rm test.iso_source/this-is-a-symlink
