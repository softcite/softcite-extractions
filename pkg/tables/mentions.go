package tables

import "github.com/apache/arrow/go/v18/arrow"

const MentionsName = "mentions"

var Mentions = arrow.NewSchema([]arrow.Field{
	{Name: "paper_id", Type: arrow.BinaryTypes.String},
	{Name: "parse_type", Type: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint8,
		ValueType: arrow.BinaryTypes.String,
		Ordered:   false,
	}},
	{Name: "mention_index", Type: arrow.PrimitiveTypes.Uint16},
	{Name: "software_raw", Type: arrow.BinaryTypes.String},
	{Name: "software_normalized", Type: arrow.BinaryTypes.String},
	{Name: "version_raw", Type: arrow.BinaryTypes.String},
	{Name: "version_normalized", Type: arrow.BinaryTypes.String},
	{Name: "publisher_raw", Type: arrow.BinaryTypes.String},
	{Name: "publisher_normalized", Type: arrow.BinaryTypes.String},
	{Name: "language_raw", Type: arrow.BinaryTypes.String},
	{Name: "language_normalized", Type: arrow.BinaryTypes.String},
	{Name: "url_raw", Type: arrow.BinaryTypes.String},
	{Name: "url_normalized", Type: arrow.BinaryTypes.String},
	{Name: "context", Type: arrow.BinaryTypes.String},
}, nil)
