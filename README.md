# SDA-sync

The sda-sync is an integration created to solve the sync issue in BigPicture. In this project, the files (and datasets) that are uploaded in one node, should be synced/backed up in another and the ingestion process should be run on both sides. However, the accession and dataset IDs should be the same in both nodes. Therefore, the integration is taking advantage of the sda-pipeline `sync` and `sync-api` services in order to send the required data and create the messages needed on the receiving side of the syncing.

## How to run the integration

The services for the two nodes can be started in parallel:
Start the services in Sweden
```sh
docker compose --profile sda-sweden up
```

Start the services in Finland
```sh
docker compose --profile sda-finland up
```

## Encrypt and upload files

There is a script under the `dev_utils` folder that cleans up the files folder and exports the keys for encryption. Running the script should download the crypt4gh keys that would allow for encrypting the `file.test` included in the `dev_utils/tools` folder
```sh
./prepare.sh
```
Note: This script should be run every time the services are re-run

Login to LS-AAI and get the token that will be used for uploading data. For example, you can login [here](https://login.gdi.nbis), get the JWToken for the user and paste it in the `tools/proxyS3` file at the `<USER-TOKEN>` and replace the `<USER-ELIXIR-ID>` in the same file. If you are using the gdi url above, the name of the user should be on the top of the page. Copy this and replace the `@` character with `_`, then paste it in the `tools/proxyS3`.

The next step is to encrypt and upload the file in the `s3Inbox` of the Swedish node, using the sda-cli downloadable [here](https://github.com/NBISweden/sda-cli/releases) (and available under the `tools` folder), running:
```sh
./sda-cli upload --config proxyS3 -encrypt-with-key keys/repo.pub.pem tools/file.test
```

## Ingest the file
Now that the file is uploaded in the S3 backend (that can be checked by logging into the minio via the browser at `localhost:9000` and making sure the file is in the inbox bucket), the ingestion process need to be initiated. 

That can be achieved using the `sda-admin` tool located at `dev_utils/tools`. The script has detailed documentation, however, here are the main commands needed to ingest the specific file. First ingest the file running the following command, after replacing the `<USER-ELIXIR-ID>` with the value used in the previous step:
```sh
./sda-admin --mq-queue-prefix sda --user <USER-ELIXIR-ID> ingest file.test.c4gh 

```

To check that the file has been ingested, run
```sh
./sda-admin --mq-queue-prefix sda --user <USER-ELIXIR-ID> accession
```
You should be able to see the file in the list, similar to:
```sh
file.test.c4gh
```
To give an accession id to this file, run the following command, replacing the `<USER-ELIXIR-ID>` and the `<ACCESSION-ID>`:
```sh
./sda-admin --mq-queue-prefix sda --user <USER-ELIXIR-ID> accession <ACCESSION-ID> file.test.c4gh
```

Finally, to create a dataset including this file, run the following command, replacing the `<USER-ELIXIR-ID>` and the `<DATASET-ID>`.
**NOTE:** The `<DATASET-ID>` should have the `centerPrefix` defined under `config.yaml` and it should be minimum 11 characters long:
```sh
./sda-admin --mq-queue-prefix sda --user `<USER-ELIXIR-ID>` dataset `<DATASET-ID>` file.test.c4gh
```
for example, if the `centerPrefix` value is `EGA`, the `<DATASET-ID>` should be of the format `EGA-<SOME-ID>`.


## Making sure everything worked

Once the dataset is created, the swedish `sync-api` service should send the required messages to the finnish endpoint, which should start the ingestion on that side.

In order to make sure that everything worked, and apart for the docker logs, you can check the bucket of the finnish (receiver) side and make sure that the file exists in the archive. Also, check that the file and the dataset exist in the database of the finnish side.
