package tables

import "github.com/apache/arrow/go/v18/arrow"

const (
	PapersName           = "papers"
	PaperIdFieldName     = "paper_id"
	HasMentionsFieldName = "has_mentions"
	paperIdComment       = "A unique identifier for the paper in this dataset"
)

var Papers = arrow.NewSchema([]arrow.Field{
	{Name: PaperIdFieldName,
		Type: arrow.PrimitiveTypes.Uint32,
		Metadata: NewMetadataBuilder().Add(
			comment, paperIdComment,
		).Build(),
	},
	{Name: "softcite_id",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The UUID of the paper in SoftCite",
		).Build(),
	},
	{Name: "title",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The parsed title of the paper",
		).Build(),
		Nullable: true,
	},
	{Name: "published_year",
		Type: arrow.PrimitiveTypes.Uint16,
		Metadata: NewMetadataBuilder().Add(
			comment, "The parsed publication year of the paper",
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
	{Name: "publication_venue",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The parsed venue the paper was published in",
		).Build(),
	},
	{Name: "publisher_name",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The parsed publisher of the paper's venue",
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
			comment, "The PubMed identifier of the paper",
		).Build(),
		Nullable: true,
	},
	{Name: "genre",
		Type: &arrow.DictionaryType{
			IndexType: arrow.PrimitiveTypes.Uint8,
			ValueType: arrow.BinaryTypes.String,
			Ordered:   false,
		},
		Metadata: NewMetadataBuilder().Add(
			comment, "The type of document the paper is, such as a journal article or a book",
		).Build(),
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
	{Name: HasMentionsFieldName,
		Type: arrow.FixedWidthTypes.Boolean,
		Metadata: NewMetadataBuilder().Add(
			comment, "Whether any mentions exist for this paper",
		).Build(),
	},
}, NewMetadataBuilder().Add(
	comment, "Papers from the SoftCite dataset",
).BuildReference())
