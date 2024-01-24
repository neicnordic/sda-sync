#!/bin/sh
set -e

apt-get -o DPkg::Lock::Timeout=60 update > /dev/null
apt-get -o DPkg::Lock::Timeout=60 install -y openssh-client openssl >/dev/null

# pip install --upgrade pip > /dev/null
# pip install aiohttp Authlib joserfc requests > /dev/null

# create EC256 key for signing the JWT tokens
mkdir -p /shared/keys/pub
if [ ! -f "/shared/keys/jwt.key" ]; then
    echo "creating jwt key"
    openssl ecparam -genkey -name prime256v1 -noout -out /shared/keys/jwt.key
    openssl ec -in /shared/keys/jwt.key -outform PEM -pubout >/shared/keys/pub/jwt.pub
    chmod 644 /shared/keys/pub/jwt.pub /shared/keys/jwt.key
fi

echo "creating token"
token="$(python /sign_jwt.py)"

cat >/shared/s3cfg <<EOD
[default]
access_key=test_dummy.org
secret_key=test_dummy.org
access_token=$token
check_ssl_certificate = False
check_ssl_hostname = False
encoding = UTF-8
encrypt = False
guess_mime_type = True
host_base = localhost:8000
host_bucket = localhost:8000
human_readable_sizes = true
multipart_chunk_size_mb = 50
use_https = False
socket_timeout = 30
EOD
