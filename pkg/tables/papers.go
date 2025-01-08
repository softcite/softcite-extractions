package tables

import "github.com/apache/arrow/go/v18/arrow"

const PapersName = "papers"

const (
	paperId = "paper_id"
)

const paperIdComment = "The UUID of the paper in SoftCite"

var Papers = arrow.NewSchema([]arrow.Field{
	{Name: paperId,
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, paperIdComment,
		).Build()},
	{Name: "title",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The parsed title of the paper",
		).Build(),
		Nullable: true,
	},
	{Name: "published_date",
		Type: arrow.PrimitiveTypes.Date32,
		Metadata: NewMetadataBuilder().Add(
			comment, "The parsed publication date of the paper",
		).Build(),
		Nullable: true,
	},
	{Name: "journal_name",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The parsed journal the paper was published in",
		).Build(),
	},
	{Name: "publisher_name",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The parsed publisher of the paper's journal",
		).Build(),
		Nullable: true,
	},
	{Name: "doi",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The doi of the paper",
		).Build(),
	},
	{Name: "pmcid",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The PubMed Central identifier of the paper",
		).Build(),
		Nullable: true,
	},
	{Name: "pmid",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The PubMed Identifier of the paper",
		).Build(),
		Nullable: true,
	},
	{Name: "genre",
		Type: &arrow.DictionaryType{
			IndexType: arrow.PrimitiveTypes.Uint8,
			ValueType: arrow.BinaryTypes.String,
			Ordered:   false,
		},
		Nullable: true,
	},
	{Name: "license_type",
		Type: &arrow.DictionaryType{
			IndexType: arrow.PrimitiveTypes.Uint8,
			ValueType: arrow.BinaryTypes.String,
			Ordered:   false,
		},
		Metadata: NewMetadataBuilder().Add(
			comment, "A normalized string of the license under which the paper was published",
		).Build(),
		Nullable: true,
	},
}, NewMetadataBuilder().Add(
	comment, "Papers from the SoftCite dataset",
).BuildReference())
