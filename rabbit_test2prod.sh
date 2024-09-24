#!/bin/bash

collections=("authors" "agreements" "grants" "patents" "articles" "sponsor_research")

for collection in "${collections[@]}"; do
echo "BatchMigration==> $collection"
./milvus-migration start -t="$collection" -c=/home/lcmjlo/repo/milvus-migration/configs/migration.yaml
done
