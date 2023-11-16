# SDA-sync

The sda-sync is an integration created to solve the *data* syncing requirement in BigPicture. In this project, every data submission that takes place in a country's node must be *mirrored* in the other country's node. This means that both the submitted data files and their corresponding database identifiers (e.g. accession and dataset ID's) in one country's node should be replicated in the other country's node. This is achieved by syncing the data between the two nodes.


Specifically, the files (and datasets) that are uploaded in one node, are synced/backed up to the other node by making sure that the ingestion process that is run on both sides will result in accession and dataset IDs that are the same in both nodes. To achieve this, the sync integration utilizes the sensitive-data-archive `sync` and `sync-api` services.

In brief, the syncing will be triggered whenever a local dataset submission, i.e. a dataset that is created on the local node, is detected. The detection is based on the dataset prefix which is unique to each country's node. In such a case, the `sync` service copies the dataset files to the inbox of the other node after these have been re-encrypted with the recipient nodes's crypt4gh public key. It also posts JSON messages to the receiving node's `sync-api` service from which `sync-api` creates and sends the necessary RabbitMQ messages to orchestrate the ingestion cycle on the receiving side of the syncing.

The `sync-api` service will be triggered whenever it receives a RabbitMQ message from the `sync` service of the other node. The `sync-api` service will then create the RabbitMQ messages needed to trigger the ingestion cycle on the receiving node.

The example integration in the `docker-compose.yml` assummes that both nodes are running inboxes that are S3 based but POSIX and sftp inbox types are also supported and may be configured.

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

**Note:** the version of the `sensitive-date-archive` services for the whole stack can be changed by editing the `.env` file that is located in the root of the repo.

## Encrypt and upload files

There is a script under the `dev_utils` folder that cleans up the files folder and exports the keys for encryption. Running the script should download the crypt4gh keys that would allow for encrypting the `file.test` included in the `dev_utils/tools` folder
```sh
./prepare.sh
```
Note: This script should be run every time the services are re-run

Login to LS-AAI and get the token that will be used for uploading data. For example, you can login [here](https://login.gdi.nbis), get the JWToken for the user and paste it in the `tools/proxyS3` file at the `<USER-TOKEN>` and replace the `<USER-ELIXIR-ID>` in the same file. If you are using the gdi url above, the name of the user should be on the top of the page. Copy this and replace the `@` character with `_`, then paste it in the `tools/proxyS3`.

The next step is to encrypt and upload the file in the `s3Inbox` of the Swedish node, using the sda-cli downloadable [here](https://github.com/NBISweden/sda-cli/releases) (and available under the `tools` folder), running:
```sh
./sda-cli upload --config ../proxyS3 -encrypt-with-key ../keys/repo.pub.pem file.test
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
