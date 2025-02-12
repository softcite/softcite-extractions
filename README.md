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

Where field names are repeated between tables, they have identical meaning (e.g. "paper_id").

Tables are mostly normalized, but with several technically-redundant precalculated fields (such as "published_year" from "published_date") which have been added for convenience.

### Papers

This table contains paper metadata.
Each entry represents a single paper analyzed by SoftCite.
Many papers do not have any associated mentions - see the `has_mentions` field.

- **paper_id** is a unique key for each paper, specific to this dataset.
- **softcite_id** is the UUID for each paper in the original SoftCite dataset.
- **title** is the title of the paper as parsed by SoftCite.
- **published_year** is the year the paper was published, calculated from published_date.
- **published_date** is the publication date of the paper as parsed by SoftCite.
- **publication_venue** is the venue the paper was published in. This covers
- **publisher_name** is the publisher of the paper's venue.
- **doi** is the raw DOI of the paper (non-URL form).
- **pmcid** is the PubMed Central identifier for the paper, if one exists.
- **pmid** is the PubMed identifier of the paper, if one exists.
- **genre** is the type of document the paper is, such as a journal article or a book. The full list of types is shown [below](#genres).
- **license_Type*** is 

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

### Appendix

#### Genres

For reference, these are the known values for the "genre" field:

- "book"
- "book-chapter"
- "book-part"
- "book-section"
- "book-series"
- "book-set"
- "database"
- "dataset"
- "dissertation"
- "edited-book"
- "grant"
- "journal"
- "journal-article"
- "journal-issue"
- "journal-volume"
- "monograph"
- "other"
- "peer-review"
- "posted-content"
- "proceedings"
- "proceedings-article"
- "proceedings-series"
- "reference-book"
- "reference-entry"
- "report"
- "report-component"
- "report-series"
- "standard"
- NA (not present)

#### Licenses

For reference, these are the known value for the "license" field:

TBD.
