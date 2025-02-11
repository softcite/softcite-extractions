package tables

import "github.com/apache/arrow/go/v18/arrow"

const (
	comment = "comment"
)

// MetadataBuilder is a convenience type to aid readability of code that
// specifies metadata for Arrow types.
type MetadataBuilder struct {
	keys   []string
	values []string
}

func NewMetadataBuilder() *MetadataBuilder {
	return &MetadataBuilder{}
}

func (b *MetadataBuilder) Add(key, value string) *MetadataBuilder {
	b.keys = append(b.keys, key)
	b.values = append(b.values, value)
	return b
}

// Build constructs and returns the arrow.Metadata.
func (b *MetadataBuilder) Build() arrow.Metadata {
	return arrow.NewMetadata(b.keys, b.values)
}

// BuildReference constructs and returns the arrow.Metadata result as a
// reference.
func (b *MetadataBuilder) BuildReference() *arrow.Metadata {
	result := b.Build()
	return &result
}
