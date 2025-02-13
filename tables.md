# Tables

This document lists every table in 

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
- **genre** is the type of document the paper is, such as a journal article or a book. The full list of genres is shown [below](#genres).
- **license_type*** is the license of the document parsed by SoftCite. The full list of licenses is shown [below](#licenses).
- **has_mentions** is whether SoftCite identified any software mentions for the paper.

### Mentions

This table contains an entry for every identified mention of a piece of software in the analyzed papers.

- **software_mention_id** is a unique key for each software mention. It is a composite of _paper_id_, _source_file_type_, and _mention_index_.
- **paper_id** is the equivalent to _paper_id_ in the Papers table.
- **source_file_type** is the format of the document parsed by SoftCite. For now this is always "pdf", but in the future may include other formats.
- **mention_index** is a unique key for each mention within a paper.
- **software_raw** is the raw string of the mentioned software.
- **software_normalized** is a normalized form of _software_raw_.
- **version_raw** is the version of the mentioned software, if present in the mention.
- **version_normalized** is a normalized form of _version_raw_.
- **publisher_raw** is the raw string of the publisher of the mentioned software, if present in the mention.
- **publisher_normalized** is a normalized form of _publisher_raw_.
- **language_raw** is the raw string of the mentioned software's programming language, if present in the  mention.
- **language_normalized** is a normalized form of _language_raw_.
- **url_raw** is the raw string of the URL for the mentioned software, if present in the mention.
- **url_normalized** is a normalized form of _url_raw_.
- **context_full_text** is the surrounding context of the software mention in the paper, as parsed by SoftCite. This is often a sentence, but can be a fragment.

### PurposeAssessments

Each mention has Purpose Assessments which try to determine whether the mention has a given purpose.
Each Mention in the Mentions table has exactly six of these assessments, one for each possible combination of scope and purpose (see below).

- **software_mention_id** is identical to _software_mention_id in the Mentions table.
- **paper_id** is identical to _paper_id_ in the Papers table.
- **source_file_type** is identical to _source_file_type_ in the Mentions table.
- **mention_index** is identical to _mention_index_ in the Mentions table.
- **scope** is either "document" or "local". A "local" scope indicates the analysis was done specifically on the local context of the mention when determining its purpose. A "document" scope indicates that the analysis covered the entire document.
- **purpose** is either "created", "used", and "shared", representing the reason the software was mentioned in this context. These purposes are not necessarily distinct: a mention could both indicate that some software was created by the papers' authors and is available on GitHub, for instance, making it both "created" and "shared".

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

- "CC BY"
- "CC BY-NC"
- "CC BY-NC-ND"
- "CC BY-NC-SA"
- "CC BY-ND"
- "CC BY-SA"
- "cc-by"
- "cc-by-nc-nd"
- "CC0"
- "NO-CC CODE"
- NA (not present)
