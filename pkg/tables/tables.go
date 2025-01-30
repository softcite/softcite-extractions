package tables

import "github.com/apache/arrow/go/v18/arrow"

const (
	Software = "software"

	ParquetExt = ".parquet"
)

var (
	GrobidRunSchema = arrow.NewSchema([]arrow.Field{
		{Name: "uuid", Type: arrow.BinaryTypes.String},
		{Name: "application", Type: &arrow.DictionaryType{
			IndexType: arrow.PrimitiveTypes.Uint8,
			ValueType: arrow.BinaryTypes.String,
			Ordered:   false,
		}},
		{Name: "date", Type: arrow.BinaryTypes.String},
		{Name: "file", Type: arrow.BinaryTypes.String},
		{Name: "softcite_file_name", Type: arrow.BinaryTypes.String},
		{Name: "id", Type: arrow.BinaryTypes.String},
		{Name: "md5", Type: arrow.BinaryTypes.String},
		{Name: "metadata.id", Type: arrow.BinaryTypes.String},
		{Name: "original_file_path", Type: arrow.BinaryTypes.String},
		{Name: "runtime", Type: arrow.PrimitiveTypes.Uint32},
		{Name: "version", Type: &arrow.DictionaryType{
			IndexType: arrow.PrimitiveTypes.Uint8,
			ValueType: arrow.BinaryTypes.String,
			Ordered:   false,
		}},
	}, nil)

	PapersSchema = arrow.NewSchema([]arrow.Field{
		{Name: "uuid", Type: arrow.BinaryTypes.String},
		{Name: "doi", Type: arrow.BinaryTypes.String},
		{Name: "year", Type: arrow.PrimitiveTypes.Uint16},
	}, nil)

	SoftwareSchema = arrow.NewSchema([]arrow.Field{
		{Name: "normalizedForm", Type: arrow.BinaryTypes.String},
		{Name: "wikidataId", Type: arrow.BinaryTypes.String},
		//{Name: "softwareType", Type: &arrow.DictionaryType{
		//	IndexType: arrow.PrimitiveTypes.Uint8,
		//	ValueType: arrow.BinaryTypes.String,
		//	Ordered:   false,
		//},
		//},
	}, nil)

	MentionsSchema = arrow.NewSchema([]arrow.Field{
		{Name: "paperId", Type: arrow.BinaryTypes.String},
		{Name: "mentionIndex", Type: arrow.PrimitiveTypes.Uint16},
		{Name: "normalizedForm", Type: arrow.BinaryTypes.String},
		{Name: "documentContextAttributes.created.value", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "documentContextAttributes.shared.value", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "documentContextAttributes.used.value", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "mentionContextAttributes.created.value", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "mentionContextAttributes.shared.value", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "mentionContextAttributes.used.value", Type: arrow.FixedWidthTypes.Boolean},
	}, nil)
)
