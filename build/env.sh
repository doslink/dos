#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
dosdir="$workspace/src/github.com/doslink"
if [ ! -L "$dosdir/dos" ]; then
    mkdir -p "$dosdir"
    cd "$dosdir"
    ln -s ../../../../../. dos
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$dosdir/dos"
PWD="$dosdir/dos"

# Launch the arguments with the configured environment.
exec "$@"
