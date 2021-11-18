#!/bin/bash

programname=$0

function usage {
    echo -e "\nUsage: $programname <command>\n"
    echo "'command' argument can be one of: 'build', 'up', 'down'"
    exit 1
}

# Check the specified command
cmd=()
if [ $# -eq 0 ]; then
    usage
elif [ "$1" == "build" ]; then
    cmd=$1
elif [ "$1" == "up" ]; then
    cmd="$1 -d"
elif [ "$1" == "down" ]; then
    cmd="$1 -v"
else
    usage
fi

# Get the parent directory of where this script is
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd "$DIR"

# Retrieve the architecture type
res=false
architecture=()
while IFS= read -r line; do
    if [[ "$line" =~ "ALGO" ]] && [[ ! "$line" =~ "#" ]]; then
        res=true
        if [[ "$line" =~ "=centralized" ]]; then
            architecture="centralized"
        else
            architecture="decentralized"
        fi
    fi
done < deployments/.env

# Check the configuration correctness
if [ "$res" = false ]; then
    echo -e "\n'deployments/.env' syntax error\n"
    exit 2
fi

# Check the profile option
profile="--profile sequencer"
opt=()
if [ "$architecture" == "centralized" ]; then
    opt=$profile
else
    opt=""
fi

# Execute the Docker Compose command
docker-compose -f deployments/docker-compose.yml $opt $cmd
