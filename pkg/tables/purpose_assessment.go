package tables

import "github.com/apache/arrow/go/v18/arrow"

var PurposeAssessmentsName = "purpose_assessments"

var PurposeAssessment = arrow.NewSchema([]arrow.Field{
	{Name: softwareMentionId,
		Type: arrow.BinaryTypes.String,
		Metadata: NewMetadataBuilder().Add(
			comment, softwareMentionIdComment,
		).Build()},
	{Name: paperId,
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
	{Name: "scope",
		Type: &arrow.DictionaryType{
			IndexType: arrow.PrimitiveTypes.Uint8,
			ValueType: arrow.BinaryTypes.String,
			Ordered:   false,
		},
		Metadata: NewMetadataBuilder().Add(
			comment,
			"Whether the assessment is about the local or document-level scope",
		).Build()},
	{Name: "purpose",
		Type: &arrow.DictionaryType{
			IndexType: arrow.PrimitiveTypes.Uint8,
			ValueType: arrow.BinaryTypes.String,
			Ordered:   false,
		},
		Metadata: NewMetadataBuilder().Add(
			comment,
			"Whether the assessment is about the software being used, created, or shared in this paper",
		).Build()},
	{Name: "certainty_score",
		Type: arrow.PrimitiveTypes.Float64,
		Metadata: NewMetadataBuilder().Add(
			comment,
			"The confidence SoftCite model has that this is the purpose of this mention, from 0.0 to 1.0",
		).Build()},
}, nil)
