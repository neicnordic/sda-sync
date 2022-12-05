#!/bin/bash

get=""
if [ "$(command -v curl)" ]; then
    get="curl"
elif [ "$(command -v wget)" ]; then
    get="wget"
else
    echo "Neither curl or wget found, exiting"
    exit 1
fi

pubkey=""
#parse input
while (( "$#" )); do
    case "$1" in
        -p|--pubkey)
            shift
            if [ ! -f "$1" ]
            then
                echo "ERROR: '$1' is not a file"
                exit 1
            fi
            pubkey="$1"
            ;;
    esac
    shift
done

cd /tmp || exit 1

ARCH=$(uname -sm | sed 's/ /_/' | tr '[:upper:]' '[:lower:]')

# Check for needed commands
C4GH=$(command -v crypt4gh)
if [ ! "$C4GH" ] || crypt4gh --version | grep -q version ; then
    echo "crypt4gh not found, downloading v1.5.3"
    if [ $get == "curl" ]; then
        curl -sL "https://github.com/neicnordic/crypt4gh/releases/download/v1.5.3/crypt4gh_$ARCH.tar.gz" | tar zxf - -C /tmp
    else
        wget -qO- "https://github.com/neicnordic/crypt4gh/releases/download/v1.5.3/crypt4gh_$ARCH.tar.gz" | tar zxf  - -C /tmp
    fi
fi


if [ -f "/keys/repo.pub.pem" ]; then
    pubkey="/keys/repo.pub.pem"
fi

# create repository crypt4gh keys
if [ -z "$pubkey" ];then
    echo "no public key supplied, creating repository key"
    /tmp/crypt4gh generate -n repo -p repoPass
    if [ -f /.dockerenv ]; then
        cp /tmp/repo* /keys
        pubkey="/keys/repo.pub.pem"
    else
        pubkey="/tmp/repo.pub.pem"
    fi
fi


finnishkey=""

if [ -f "/keys/finnish-repo.pub.pem" ]; then
    finnishkey="/keys/finnish-repo.pub.pem"
fi

# create repository crypt4gh keys
if [ -z "$finnishkey" ];then
    echo "no public key supplied, creating repository key"
    /tmp/crypt4gh generate -n finnish-repo -p repoPass
    if [ -f /.dockerenv ]; then
        cp /tmp/finnish-repo* /keys
        finnishkey="/keys/finnish-repo.pub.pem"
    else
        finnishkey="/tmp/finnish-repo.pub.pem"
    fi
fi

