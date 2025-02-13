# software-mentions-dataset-analysis
Analyses of software mentions and dependencies

## What this dataset is

The software-mentions dataset is a collection of ML-identified mentions of software
detected in about 24,000,000 academic papers.

### The data model

A __paper__ can contain many __mentions__, each of which was found in a full-text snippet of __context__, and extracts the (raw and normalized) __software name__ , __version number__, __creator__, __url__, as well as associated __citation__ to the reference list of the paper.

Each __mention__ has multiple __purpose assessments__ about the relationship between the software and the paper: Was the software __used__ in the research?, Was it __created__ in the course of the research?, Was the software __shared__ alongside this paper? These probabilistic assessments (0..1 range) are made in two ways: using only the information from the specific mention and using all the mentions within a single paper together (mention-level vs document-level); thus each mention has six __purpose assessments__.

ER diagram goes here.

## Getting Started

### Getting the Parquet files

If you want to extract the .parquet tables yourself, or work with the original dataset, see [Extracting Tables](EXTRACTING_TABLES.md).
Otherwise, you can download the tables in a friendlier format from (INSERT LOCATION).
