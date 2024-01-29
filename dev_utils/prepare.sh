#!/bin/bash

rm tools/ch*
rm tools/file.test.c4gh
rm -r keys

docker cp ingest:/keys .
docker cp oidc:/shared/s3cfg .
