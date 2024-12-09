# Extracting Tables

These instructions are for those who wish to extract their own tables from the
software-mentions dataset, or otherwise want to handle the dataset in its
original form. If you just want the already-available Parquet tables, see the
[Readme](README.md). 

If you already have the .jsonl files, you can skip to part 2.

## Part 1: Converting to .jsonl

Before starting, have an empty hard drive with at least 2TB of storage and
at least 120 million inodes available.

Please read this entire paragraph before doing anything, or your system may
become temporarily unusable. The raw software-mentions dataset is ~150 GiB
before being uncompressed, and ~800 GiB uncompressed. The dataset is a
hierarchical collection of about ~100 million folders and JSON files. This means
if your disk does not have at least this many inodes available, the extraction
process will fail when the filesystem runs out of inodes, even if there is still
disk space available. Running out of inodes can make many common operations
unstable. To check the free inodes available on a Linux system, use
`df -i` and check the `IFree` column. Ensure this number is comfortably over 100
million - preferably at least 120 million.

Even if your disk has sufficient inodes, we do not recommend running this
process on the same disk as your operating system. The fast response time of
many filesystem operations is due to your disk caching recently
accessed parts of the filesystem. After you have extracted software-mentions,
you can expect many operations, such as listing the contents of a
directory, to take over a second.

## Part 2: Extracting Parquet Tables
