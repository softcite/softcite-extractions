# Extracting Tables

The software-mentions dataset exists in three different formats.

The first, the raw extractions, is the raw output of running the `software-mentions` application on the paper dataset.
This is a large hierarchical file structure with several directories and files for each of the ~20 million parsed papers.
The second, the JSONL files, is a collated form of the raw output.
These files are sequences of JSON objects, each representing paper metadata or extracted software mentions in the paper.
Part 1 of these instructions details converting to this format.
The last, the Parquet tables, is a more user-friendly form of the data in a columnar data format.
Part 2 of these instructions details how to convert the JSONL files to the Parquet tables, or to create new Parquet table definitions.

These instructions are for those who wish to extract their own tables from the software-mentions dataset, or otherwise want to handle the dataset in its  original form.
If you just want the already-available Parquet tables, see the [Readme](README.md).
Instructions assume commands are run from a Linux distribution.

If you already have the .jsonl files, you can skip to part 2.

## Part 1: Converting to .jsonl

Before starting, have an empty secondary hard drive with at least 2TB of storage and at least 120 million inodes available.
Most 1TB disks do not have sufficient inodes to successfully extract the software-mentions dataset.
This process may take over a week in total.
Performing other work on the target disk will be slow until you complete these steps.

Please read this entire paragraph before doing anything, or your system may become temporarily unusable. The raw software-mentions dataset is ~150 GB compressed and ~800 GB uncompressed.
The dataset is a hierarchical collection of about 100 million folders and JSON files.
If your disk does not have at least this many inodes available, the extraction process will fail when the filesystem runs out of inodes, even if there is still disk space available.
Running out of inodes can make many common operations slow or unstable.
To check the free inodes available on your disk, use `df -i` and check the `IFree` column.
Ensure this number is comfortably over 100  million - preferably at least 120 million for the disk you will extract the dataset to.

Even if your disk has sufficient inodes, we do not recommend running this process on the same disk as your operating system.
The fast response time of many filesystem operations is due to your disk caching recently accessed parts of the filesystem, and this task will disrupt many of the caching logic's assumptions.
After you have extracted software-mentions, you can expect many operations, such as listing the contents of a directory, to take over a second on the target disk.

The below steps have not been rigorously tested.
Please contact beason@utexas.edu if you encounter problems or wish to assist in improving this documentation.
As-is the process takes about a week and so making these robust is not a priority.

1. Use `unzip` to decompress the archive to the target disk.

This operation may take 48-72 hours.
If you chose to use an alternative such as 7zip, you may experience performance issues and the process may hang indefinitely.
The result of the operation will be a nested directory structure containing multiple files for each analyzed paper.
The directory is based on the software-mentions UUID of the analyzed paper, using the first 8 characters to generate directories.

The file `AABBCCDD-EEFF-GGHH-IIJJ-KKLLMMNNOOPP.json`, representing the paper's metadata, will be located in the directory:

    AA/BB/CC/DD/AABBCCDD-EEFF-GGHH-IIJJ-KKLLMMNNOOPP/...

Each directory will contain at least two files, and rarely six or more.
Different filename patterns indicate different parses of the paper, and potentially contain different sets of detected software mentions from that parse.

2. Use the `merge` command in `cmd/merge` to merge the small JSON files into JSONL files.

```shell
go run cmd/merge/merge.go IN OUT
```

The `IN` path should be the top level directory containing the 256 subdirectories representing the first two characters of the UUID.
The `OUT` path should ideally be an empty directory, as about 1,500 files will be written to it.
This process overwrites existing conflicting files without warning.
This process may take 48-72 hours, and will run faster if `IN` and `OUT` point to different disks.

The `merge` command may be used on a subdirectory to only merge JSON files in a portion of the dataset.
Before merging the entire dataset, we recommend merging a small portion of the files to ensure the process is running smoothly, such as with:

```shell
go run cmd/merge/merge.go IN/AA OUT
```

This will merge approximately 1/256th of the data.
If you are not interested in the remaining files that were not covered by this merge process, you may now delete the directory
extracted by `unzip` and proceed to Part 2.

3. \[Optional\] Use `rm-processed` to recursively remove the files and directories that were merged into JSONL via the above.

```shell
go run cmd/rm-processed/rm-processed.go IN
```

There will be a small number of files and directories remaining (~2,000).
These files are not handled by the logic above, and their data is not present in the resulting JSONL files.
Per the author of the dataset, these files are errors and may be safely ignored and deleted.

## Part 2: Extracting Parquet Tables

To begin this step, you need the dataset as .jsonl files.
There should be about 1,500 of these in a flat directory.
Files are groups of either metadata or detected software mentions, grouped by the first two characters of the software-mentions UUID.
Each of the 256 two-letter prefixes has up to six files.

- Every prefix includes `.papers.jsonl.gz`, which is the paper metadata.
- Every prefix includes `.software.jsonl.gz`, which is the extracted software mentions from the default paper parse.
- The four other prefixes indicate alternative parses (JATS, Pub2Tei, LaTeX, and GROBID) for each paper.
  Many papers do not have alternative parses, and some UUID prefix groups do not include any of a particular parse.

### The JSONL Format

The .jsonl format is a sequence of JSON objects delimited by newline.
Many JSON decoders, such as `encoding/json.Encoder` in Go, handle this automatically and can treat the contents of these files as a stream of JSON objects.

### Table Definitions

Table definitions are defined in Go code.
These are currently a work in progress.

### Extracting Tables

To extract tables, run `extract-columns`, passing both the IN_DIR containing the JSONL files and the out directory to write tables to.

To fully extract tables, you need to run the three commands in order. That the first and third command are identical is not a mistake. The full process will take approximately one hour and will use about 10GiB of memory while running.

```shell
IN_DIR=path/to/input
OUT_DIR=path/to/output
go run cmd/extract-columns/extract-columns.go papers "${IN_DIR}" "${OUT_DIR}"
go run cmd/extract-columns/extract-columns.go pdf "${IN_DIR}" "${OUT_DIR}"
go run cmd/extract-columns/extract-columns.go papers "${IN_DIR}" "${OUT_DIR}"
```

The reason for this is a circular dependency between the datasets, which can only be resolved by iterating over at least one of the datasets twice:
1. Papers has a "has_mentions" field, which requires knowledge from the Mentions table of whether any mentions exist for a paper.
2. Mentions has a "paper_id" field, which is computed as part of extracting the Papers table.

This process produces two incidental files, `paper_ids.csv` and `has_mentions.csv`.
These files are produced deterministically by `extract-columns`, and so it is unnecessary to maintain them.
Respectively, they contain a map from SoftCite UUID to paper_id and a list of SoftCite UUIDs which have at least one software mention.
