# SDA-sync

The sda-sync is an integration created to solve the sync/backup issue in BigPicture. In this project, the files (and datasets) that are uploaded in one node, should be synced/backed up in another and the ingestion process should be run on both sides. However, the accession and dataset IDs should be the same in both nodes. Therefore, the integration is taking advantage of the sda-pipeline sync-api in order to send the required data and create the messages needed on the receiving side of the syncing.

## How to run the integration

First create an image from the sftp server by navigating to the `dev_utils/sftp-server` and running
```sh
cd dev_utils/sftp-server
docker build -t sftp-server .
```

The services for the two nodes can be started in parallel:
Start the services in Sweden
```sh
docker compose --profile sda-sweden up
```

Start the services in Finland
```sh
docker compose --profile sda-finland up
```
Note: In case the mock-cega service fails for any of the nodes, try running the command(s) above again.

## Encrypt and upload files

There is a script under the `dev_utils` folder that cleans up the files folder and exports the keys for encryption. Running the script should download the crypt4gh keys that would allow for encrypting the `file.test` included in the `dev_utils/tools` folder
```sh
./prepare.sh
```
Note: This script should be run every time the services are re-run

The next step is to encrypt the file, using either the sda-cli downloadable [here](https://github.com/NBISweden/sda-cli/releases) and using:
```sh
./sda-cli encrypt -key keys/repo.pub.pem tools/file.test
```
or using the crypt4gh downloadable [here](https://github.com/neicnordic/crypt4gh/releases) and using:
```sh
./crypt4gh encrypt -p keys/repo.pub.pem -f tools/file.test
```

You can try to upload this file using the `s3cmd`. This should also trigger a message from the s3Proxy to the RabbitMQ:
```sh
s3cmd -c proxyS3 put tools/file.test.c4gh s3://dummy/
```

Sometimes that fails, so you might need to first upload the file manually in the `s3://inbox/dummy` location by accessing the s3 at `localhost:9001`. Then you can run the command above, which will again trigger the messages.

## Making sure everything worked

Once the file is uploaded, it should be backed up in the Finnish SFTP. Also, a message should be sent from the Swedish sync API to the Finnish sync API, therefore, triggering the ingestion on the other side. Finally, the file should be backuped in the s3 of the Finnish side, under `s3://backup/dummy/file.test.c4gh`.

