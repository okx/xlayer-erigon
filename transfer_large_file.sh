#!/bin/bash
set -e

SOURCE_FILE="./test/mainnet/seq/chaindata/mdbx.dat"
POD_NAME="dev-xlayer-689687fb9d-jf276"
DEST_PATH="/mainnet/mainnet/seq/chaindata"
CHUNK_SIZE="1G"

# echo "Splitting file..."
# split -b $CHUNK_SIZE -d "$SOURCE_FILE" "${SOURCE_FILE}.part_"

echo "Transferring parts..."
for part in "${SOURCE_FILE}.part_"[0-9][0-9]; do
    echo "Transferring $part..."
    filename=$(basename "$part")
    ./krsync.sh -vz --progress --stats --checksum "$part" "$POD_NAME:$DEST_PATH/$filename"
done

echo "Calculating source file checksum..."
SOURCE_CHECKSUM=$(sha256sum "$SOURCE_FILE" | awk '{print $1}')

echo "Reassembling file in pod..."
kubectl exec "$POD_NAME" -- bash -c "cd $DEST_PATH && cat ${SOURCE_FILE}.part_* > $SOURCE_FILE"

echo "Verifying checksum..."
DEST_CHECKSUM=$(kubectl exec "$POD_NAME" -- bash -c "cd $DEST_PATH && sha256sum $(basename $SOURCE_FILE)" | awk '{print $1}')

if [ "$SOURCE_CHECKSUM" = "$DEST_CHECKSUM" ]; then
    echo "Checksum verification successful!"
else
    echo "Checksum verification failed!"
    echo "Source checksum: $SOURCE_CHECKSUM"
    echo "Destination checksum: $DEST_CHECKSUM"
    exit 1
fi

# echo "Cleaning up local parts..."
# rm "${SOURCE_FILE}.part_"*

echo "Done!" 