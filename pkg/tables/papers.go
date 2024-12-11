package tables

import "github.com/apache/arrow/go/v18/arrow"

const PapersName = "papers"

var Papers = arrow.NewSchema([]arrow.Field{
	{Name: "id", Type: arrow.BinaryTypes.String},
	{Name: "title", Type: arrow.BinaryTypes.String},
	{Name: "published_date", Type: arrow.PrimitiveTypes.Date32},
	{Name: "journal_name", Type: arrow.BinaryTypes.String},
	{Name: "publisher", Type: arrow.BinaryTypes.String},
	{Name: "doi", Type: arrow.BinaryTypes.String},
	{Name: "pmcid", Type: arrow.BinaryTypes.String},
	{Name: "pmid", Type: arrow.BinaryTypes.String},
	{Name: "genre", Type: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint8,
		ValueType: arrow.BinaryTypes.String,
		Ordered:   false,
	}},
	{Name: "license", Type: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint8,
		ValueType: arrow.BinaryTypes.String,
		Ordered:   false,
	}},
}, nil)
