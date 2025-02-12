# software-mentions-dataset-analysis
Analyses of software mentions and dependencies

## What this dataset is

The software-mentions dataset is a collection of ML-identified mentions of software
detected in about 24,000,000 academic papers.

### The data model

A _paper_ can contain many _mentions_, each of which is found in a full text snippet of _context_, and extracts the _software name_ (raw and normalized), the _version number_, a _url_, a _creator_.

Each _mention_ has multiple _purpose assessments_ about the relationship between the software and the paper: Was the software _used_ in the research?, Was it _created_ in the course of the research?, Was the software _shared_ alongside this paper? These probabilistic assessments are made in two ways: using only the information from the specific mention and using all the mentions within a single paper together (a document-level purpose assessment); thus each mention has six _purpose assessments_.

ER diagram goes here.

## Getting Started

### Getting the Parquet files

If you want to extract the .parquet tables yourself, or work with the original dataset, see [Extracting Tables](EXTRACTING_TABLES.md).
Otherwise, you can download the tables in a friendlier format from (INSERT LOCATION).
