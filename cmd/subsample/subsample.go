package main

import (
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/apache/arrow/go/v18/parquet"
	"github.com/apache/arrow/go/v18/parquet/compress"
	"github.com/apache/arrow/go/v18/parquet/file"
	"github.com/apache/arrow/go/v18/parquet/pqarrow"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/willbeason/software-mentions/pkg/tables"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const (
	batchSize = 1 << 20
)

const (
	FlagPartitions = "partitions"
	FlagSeed       = "seed"
)

func init() {
	cmd.Flags().Float64Slice(FlagPartitions, []float64{0.01, 0.05}, "dataset partitions")
	cmd.Flags().Int64(FlagSeed, 0, "random seed")
}

func main() {
	flag.Parse()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cmd = cobra.Command{
	Use:     "subsample IN_DIR OUT_DIR",
	Short:   "subsamples the SoftCite dataset",
	Args:    cobra.ExactArgs(2),
	Version: "0.1.0",
	RunE:    runE,
}

func runE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	inPath := args[0]
	outDir := args[1]

	err := os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	partitions, err := cmd.Flags().GetFloat64Slice(FlagPartitions)
	if err != nil {
		return fmt.Errorf("getting partitions: %w", err)
	}

	thresholds := make([]float64, len(partitions))
	sum := 0.0
	for i, partition := range partitions {
		sum += partition
		thresholds[i] = sum
	}

	seed, err := getSeed(cmd)
	if err != nil {
		return fmt.Errorf("getting seed: %w", err)
	}
	inPapers := filepath.Join(inPath, tables.PapersName+tables.ParquetExt)
	paperPartitions, err := getPartitions(ctx, seed, inPapers, thresholds)
	if err != nil {
		return fmt.Errorf("getting paper partitions: %w", err)
	}

	for _, partition := range paperPartitions {
		fmt.Println(len(partition))
	}

	outPapers := filepath.Join(outDir, tables.PapersName+tables.ParquetExt)
	err = partitionParquet(ctx, inPapers, outPapers, paperPartitions)
	if err != nil {
		return fmt.Errorf("partitioning papers: %w", err)
	}

	inMentions := filepath.Join(inPath, tables.MentionsName+".pdf"+tables.ParquetExt)
	outMentions := filepath.Join(outDir, tables.MentionsName+".pdf"+tables.ParquetExt)
	err = partitionParquet(ctx, inMentions, outMentions, paperPartitions)
	if err != nil {
		return fmt.Errorf("partitioning mentions: %w", err)
	}

	inAssessments := filepath.Join(inPath, tables.PurposeAssessmentsName+".pdf"+tables.ParquetExt)
	outAssessments := filepath.Join(outDir, tables.PurposeAssessmentsName+".pdf"+tables.ParquetExt)
	err = partitionParquet(ctx, inAssessments, outAssessments, paperPartitions)
	if err != nil {
		return fmt.Errorf("partitioning assessments: %w", err)
	}

	return nil
}

func partitionParquet(ctx context.Context, inPath, outPath string, partitions []map[uint32]struct{}) error {
	allocator := memory.NewGoAllocator()
	inFileReader, err := file.OpenParquetFile(inPath, true)
	if err != nil {
		return fmt.Errorf("opening parquet file %q: %w", inPath, err)
	}

	inReader, err := pqarrow.NewFileReader(inFileReader,
		pqarrow.ArrowReadProperties{Parallel: true, BatchSize: batchSize},
		allocator,
	)
	if err != nil {
		return fmt.Errorf("creating pqarrow FileReader: %w", err)
	}

	recordReader, err := inReader.GetRecordReader(ctx, nil, nil)
	if err != nil {
		return fmt.Errorf("getting record reader: %w", err)
	}

	paperSchema, err := inReader.Schema()
	if err != nil {
		return fmt.Errorf("getting schema: %w", err)
	}

	outPaperFiles := make([]string, len(partitions))
	writers := make([]*pqarrow.FileWriter, len(partitions))
	recordBuilders := make([]*array.RecordBuilder, len(partitions))
	for i := range partitions {
		ext := filepath.Ext(outPath)
		outPathI := fmt.Sprintf("%s_%d%s", outPath[:len(outPath)-len(ext)], i, ext)
		outPapersFile, err := os.Create(outPathI)
		outPaperFiles[i] = outPathI

		if err != nil {
			return fmt.Errorf("creating paper ids file %q: %w", outPath, err)
		}
		writer, err := pqarrow.NewFileWriter(
			paperSchema,
			outPapersFile,
			parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Gzip), parquet.WithCompressionLevel(gzip.BestCompression)),
			pqarrow.DefaultWriterProps(),
		)
		if err != nil {
			return fmt.Errorf("creating writer: %w", err)
		}
		writers[i] = writer

		defer func() {
			err := writer.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()

		recordBuilder := array.NewRecordBuilder(allocator, paperSchema)
		defer recordBuilder.Release()
		recordBuilders[i] = recordBuilder
	}

	var paperIdFieldIndex int
	for i, field := range paperSchema.Fields() {
		if field.Name == tables.PaperIdFieldName {
			paperIdFieldIndex = i
			break
		}
	}

	var record arrow.Record
	for record, err = recordReader.Read(); err == nil; record, err = recordReader.Read() {
		idColumn, ok := record.Column(paperIdFieldIndex).(*array.Uint32)
		if !ok {
			return fmt.Errorf("expected paper id column to be of type *array.Uint32, got %T", record.Column(paperIdFieldIndex))
		}

		for i, id := range idColumn.Uint32Values() {
			partitionNum := -1
			for i, partition := range partitions {
				if _, found := partition[id]; found {
					partitionNum = i
					break
				}
			}

			if partitionNum == -1 {
				continue
			}

			recordBuilder := recordBuilders[partitionNum]

			for j, field := range recordBuilder.Schema().Fields() {
				switch field.Type {
				case arrow.FixedWidthTypes.Boolean:
					recordBuilder.Field(j).(*array.BooleanBuilder).Append(record.Column(j).(*array.Boolean).Value(i))
				case arrow.PrimitiveTypes.Uint16:
					recordBuilder.Field(j).(*array.Uint16Builder).Append(record.Column(j).(*array.Uint16).Value(i))
				case arrow.PrimitiveTypes.Uint32:
					recordBuilder.Field(j).(*array.Uint32Builder).Append(record.Column(j).(*array.Uint32).Value(i))
				case arrow.PrimitiveTypes.Float64:
					recordBuilder.Field(j).(*array.Float64Builder).Append(record.Column(j).(*array.Float64).Value(i))
				case arrow.PrimitiveTypes.Date32:
					recordBuilder.Field(j).(*array.Date32Builder).Append(record.Column(j).(*array.Date32).Value(i))
				case arrow.BinaryTypes.String:
					recordBuilder.Field(j).(*array.StringBuilder).Append(record.Column(j).(*array.String).Value(i))
				default:
					return fmt.Errorf("unsupported field type %s", field.Type)
				}
			}
		}
	}

	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading records: %w", err)
	}

	for i := range partitions {
		recordBuilder := recordBuilders[i]
		writer := writers[i]

		recordToWrite := recordBuilder.NewRecord()
		defer recordToWrite.Release()
		err = writer.Write(recordToWrite)
		if err != nil {
			return fmt.Errorf("writing record: %w", err)
		}
	}

	return nil
}

func getPartitions(ctx context.Context, seed int64, inPapers string, thresholds []float64) ([]map[uint32]struct{}, error) {
	allocator := memory.NewGoAllocator()
	inPapersFileReader, err := file.OpenParquetFile(inPapers, true)
	if err != nil {
		return nil, fmt.Errorf("opening parquet file %q: %w", inPapers, err)
	}

	inPapersReader, err := pqarrow.NewFileReader(inPapersFileReader,
		pqarrow.ArrowReadProperties{Parallel: true, BatchSize: batchSize},
		allocator,
	)
	if err != nil {
		return nil, fmt.Errorf("creating pqarrow FileReader: %w", err)
	}

	schema, err := inPapersReader.Schema()
	if err != nil {
		return nil, fmt.Errorf("getting schema: %w", err)
	}
	var paperIdFieldIndex int
	var hasMentionsFieldIndex int
	for i, field := range schema.Fields() {
		switch field.Name {
		case tables.PaperIdFieldName:
			paperIdFieldIndex = i
		case tables.HasMentionsFieldName:
			hasMentionsFieldIndex = i
		}
	}

	recordReader, err := inPapersReader.GetRecordReader(ctx, []int{paperIdFieldIndex, hasMentionsFieldIndex}, nil)
	if err != nil {
		return nil, fmt.Errorf("getting record reader: %w", err)
	}
	rng := rand.New(rand.NewSource(seed))
	paperPartitions := make([]map[uint32]struct{}, len(thresholds))

	var record arrow.Record
	for record, err = recordReader.Read(); err == nil; record, err = recordReader.Read() {
		columns := record.Columns()
		idColumn, ok := columns[0].(*array.Uint32)
		if !ok {
			return nil, fmt.Errorf("expected paper id column to be of type *array.Uint32, got %T", columns[0])
		}
		hasMentionsColumn, ok := columns[1].(*array.Boolean)
		if !ok {
			return nil, fmt.Errorf("expected has mentions column to be of type *array.Boolean, got %T", columns[1])
		}

		for i, id := range idColumn.Uint32Values() {
			if hasMentionsColumn.Value(i) {
				randValue := rng.Float64()
				for j, threshold := range thresholds {
					if randValue < threshold {
						if paperPartitions[j] == nil {
							paperPartitions[j] = make(map[uint32]struct{})
						}
						paperPartitions[j][id] = struct{}{}
						break
					}
				}
			}
		}
	}
	if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("reading records: %w", err)
	}
	return paperPartitions, nil
}

func getSeed(cmd *cobra.Command) (int64, error) {
	// Check if the user set the seed manually.
	seedSet := false
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Name == FlagSeed {
			seedSet = true
		}
	})

	if seedSet {
		// User-provided seed.
		seed, err := cmd.Flags().GetInt64(FlagSeed)
		if err != nil {
			return 0, err
		}
		return seed, nil
	} else {
		// Use time as seed.
		return time.Now().UnixNano(), nil
	}
}
