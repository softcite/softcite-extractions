# software-mentions-dataset-analysis
Analyses of software mentions and dependencies

## What this dataset is

The software-mentions dataset is a collection of ML-identified mentions of software
detected in about 24,000,000 academic papers.

## Getting Started

### Getting the Parquet files

If you want to extract the .parquet tables yourself, or work with the original dataset, see [Extracting Tables](EXTRACTING_TABLES.md).
Otherwise, you can download the tables in a friendlier format from (INSERT LOCATION).

## Table Definitions

The Parquet files are three tables of the SoftCite data.
They do not contain all fields in the SoftCite dataset, but are a (hopefully useful) subset specifically related to mentions.

Much of the information below can be gleaned from the metadata field `comment`, which is present in every table and for every field.
Where this documentation conflicts with what is in `comment`, trust what is in `comment`.

### Papers

This table contains paper metadata.
Each entry represents a single paper analyzed by SoftCite.
Many papers do not have any associated mentions - see the `has_mentions` field.

### Mentions

This table contains an entry for every identified mention of a piece of software in the analyzed papers.

### PurposeAssessments

Each mention has Purpose Assessments which try to determine whether the mention has a given purpose.
Each Mention in the Mentions table has exactly six of these assessments.

There are three possible mention purposes: "created", "used", and "shared".
These purposes are not necessarily distinct: a mention could both indicate that some software was created by the papers' authors and is available on GitHub, for instance, making it both "created" and "shared".

There are two possible mention scopes: "local" and "document".
A "local" scope indicates the analysis was done specifically on the local context of the mention when determining its purpose.
A "document" scope indicates that the analysis covered the entire document.
