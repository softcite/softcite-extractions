package tables

import "github.com/apache/arrow/go/v18/arrow"

const MentionsName = "mentions"

const (
	softwareMentionId = "software_mention_id"
	sourceFileType    = "source_file_type"
	mentionIndex      = "mention_index"
)

const (
	softwareMentionIdComment = "A concatenation of paper_id, source_file_type, and mention_index"
	sourceFileTypeComment    = "The extension of the source file parsed by SoftCite. " +
		"There may be more than one source file per paper."
	mentionIndexComment = "The index of the mention parsed from the source file"
)

var SoftwareMentions = arrow.NewSchema([]arrow.Field{
	{Name: softwareMentionId,
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, softwareMentionIdComment,
		).Build()},
	{Name: PaperIdFieldName,
		Type: arrow.PrimitiveTypes.Uint32,
		Metadata: NewMetadataBuilder().Add(
			comment, paperIdComment,
		).Build()},
	{Name: sourceFileType,
		Type: &arrow.DictionaryType{
			IndexType: arrow.PrimitiveTypes.Uint8,
			ValueType: arrow.BinaryTypes.String,
			Ordered:   false,
		},
		Metadata: NewMetadataBuilder().Add(
			comment,
			sourceFileTypeComment,
		).Build()},
	{Name: mentionIndex,
		Type: arrow.PrimitiveTypes.Uint16,
		Metadata: NewMetadataBuilder().Add(
			comment, mentionIndexComment,
		).Build()},
	{Name: "software_raw",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The raw string of the software mentioned",
		).Build()},
	{Name: "software_normalized",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "A normalized string of the software mentioned",
		).Build()},
	{Name: "version_raw",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The raw string of the mentioned software's version",
		).Build(),
		Nullable: true},
	{Name: "version_normalized",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "A normalized string of the mentioned software's version",
		).Build(),
		Nullable: true},
	{Name: "publisher_raw",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The raw string of the mentioned software's publisher",
		).Build(),
		Nullable: true},
	{Name: "publisher_normalized",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "A normalized string of the mentioned software's publisher",
		).Build(),
		Nullable: true},
	{Name: "language_raw",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The raw string of the mentioned software's programming language",
		).Build(),
		Nullable: true},
	{Name: "language_normalized",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "A normalized string of the mentioned software's programming language",
		).Build(),
		Nullable: true},
	{Name: "url_raw",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The raw string of the a URL for the mentioned software",
		).Build(),
		Nullable: true},
	{Name: "url_normalized",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "A normalized string of the a URL for the mentioned software",
		).Build(),
		Nullable: true},
	{Name: "context_full_text",
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, "The software mention as it appears in the full text of the paper",
		).Build(),
		Nullable: true},
}, NewMetadataBuilder().BuildReference())
