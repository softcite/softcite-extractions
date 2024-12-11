package tables

import "github.com/apache/arrow/go/v18/arrow"

var PurposeAssessmentsName = "purpose_assessment"

var PurposeAssessments = arrow.NewSchema([]arrow.Field{
	{Name: "paper_id", Type: arrow.BinaryTypes.String},
	{Name: "parse_type", Type: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint8,
		ValueType: arrow.BinaryTypes.String,
		Ordered:   false,
	}},
	{Name: "mention_index", Type: arrow.PrimitiveTypes.Uint16},
	{Name: "context", Type: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint8,
		ValueType: arrow.BinaryTypes.String,
		Ordered:   false,
	}},
	{Name: "purpose", Type: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint8,
		ValueType: arrow.BinaryTypes.String,
		Ordered:   false,
	}},
	{Name: "certainty", Type: arrow.PrimitiveTypes.Float64},
	{Name: "is_purpose", Type: arrow.FixedWidthTypes.Boolean},
}, nil)
