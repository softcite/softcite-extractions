package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/apache/arrow/go/v18/parquet"
	"github.com/apache/arrow/go/v18/parquet/compress"
	"github.com/apache/arrow/go/v18/parquet/pqarrow"
	"github.com/spf13/cobra"
	"github.com/willbeason/bondsmith/fileio"
	"github.com/willbeason/bondsmith/jsonio"
	"github.com/willbeason/software-mentions/pkg/tables"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cmd = cobra.Command{
	Use:     "extract-columns [papers|software] IN_DIR OUT_DIR",
	Short:   "converts parts of the dataset into the Apache Parquet format",
	Args:    cobra.ExactArgs(3),
	Version: "0.1.0",
	RunE:    runE,
}

func runE(_ *cobra.Command, args []string) error {
	inPath := args[1]
	outDir := args[2]

	inFile, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer func() {
		err = inFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var reader io.Reader
	reader, err = toReader(inPath, args[0])
	if err != nil {
		return err
	}

	// gzip correctly handles concatenated files.
	reader, err = gzip.NewReader(reader)
	if err != nil {
		return err
	}

	switch args[0] {
	case "papers":
		return extractPapers(reader, outDir, err)
	case "software":
		return extractSoftware(reader, outDir, err)
	default:
		return fmt.Errorf("must be either papers or software, not %s", args[0])
	}
}

func toReader(inPath, extractType string) (*fileio.MultiReader, error) {
	stat, err := os.Stat(inPath)
	if err != nil {
		return nil, err
	}

	var inPaths []string
	if stat.IsDir() {
		entries, err := os.ReadDir(inPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if extractType == "software" {
				if !strings.Contains(entry.Name(), "software") {
					continue
				}
				if entry.Name()[2:] != ".software.jsonl.gz" {
					continue
				}
			} else if entry.Name()[2:] != ".jsonl.gz" {
				continue
			}

			entryPath := filepath.Join(inPath, entry.Name())
			fmt.Println(entry.Name())

			inPaths = append(inPaths, entryPath)
		}
	} else {
		inPaths = append(inPaths, inPath)
	}

	return fileio.NewMultiFileReader(inPaths), nil
}

type SoftwareMentions struct {
	File     string            `json:"file"`
	Mentions []SoftwareMention `json:"mentions"`
}

type SoftwareMention struct {
	SoftwareName SoftwareName `json:"software-name"`
	SoftwareType string       `json:"software-type"`

	DocumentContextAttributes ContextAttributes `json:"documentContextAttributes"`
	MentionContextAttributes  ContextAttributes `json:"mentionContextAttributes"`
}

type ContextAttributes struct {
	Created ScoreValue `json:"created"`
	Shared  ScoreValue `json:"shared"`
	Used    ScoreValue `json:"used"`
}

type ScoreValue struct {
	Score float32 `json:"score"`
	Value bool    `json:"value"`
}

type SoftwareName struct {
	NormalizedForm string `json:"normalizedForm"`
	WikidataId     string `json:"wikidataId"`
}

func extractSoftware(reader io.Reader, outDir string, err error) error {
	softwareMentions := jsonio.NewReader(reader, func() *SoftwareMentions {
		return &SoftwareMentions{}
	})

	allocator := memory.NewGoAllocator()

	softwareRecordBuilder := array.NewRecordBuilder(allocator, tables.SoftwareSchema)
	defer softwareRecordBuilder.Release()

	mentionsRecordBuilder := array.NewRecordBuilder(allocator, tables.MentionsSchema)
	defer mentionsRecordBuilder.Release()

	seenSoftware := make(map[string]bool)

	for softwareMention, err := range softwareMentions.Read() {
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}

		if len(softwareMention.Mentions) > math.MaxUint16 {
			panic(len(softwareMention.Mentions))
		}

		for index, mention := range softwareMention.Mentions {
			normalizedForm := mention.SoftwareName.NormalizedForm

			mentionsRecordBuilder.Field(0).(*array.StringBuilder).
				Append(softwareMention.File[:36])
			mentionsRecordBuilder.Field(1).(*array.Uint16Builder).
				Append(uint16(index))
			mentionsRecordBuilder.Field(2).(*array.StringBuilder).
				Append(normalizedForm)
			mentionsRecordBuilder.Field(3).(*array.BooleanBuilder).
				Append(mention.DocumentContextAttributes.Created.Value)
			mentionsRecordBuilder.Field(4).(*array.BooleanBuilder).
				Append(mention.DocumentContextAttributes.Shared.Value)
			mentionsRecordBuilder.Field(5).(*array.BooleanBuilder).
				Append(mention.DocumentContextAttributes.Used.Value)
			mentionsRecordBuilder.Field(6).(*array.BooleanBuilder).
				Append(mention.MentionContextAttributes.Created.Value)
			mentionsRecordBuilder.Field(7).(*array.BooleanBuilder).
				Append(mention.MentionContextAttributes.Shared.Value)
			mentionsRecordBuilder.Field(8).(*array.BooleanBuilder).
				Append(mention.MentionContextAttributes.Used.Value)

			// We assume software with the same normalizedForm are identical.
			if seenSoftware[normalizedForm] {
				continue
			}
			seenSoftware[normalizedForm] = true

			softwareRecordBuilder.Field(0).(*array.StringBuilder).
				Append(normalizedForm)
			softwareRecordBuilder.Field(1).(*array.StringBuilder).
				Append(mention.SoftwareName.WikidataId)
			//softwareRecordBuilder.Field(2).(*array.StringBuilder).
			//	Append(mention.SoftwareType)
		}
	}

	err = writeRecords(tables.SoftwareSchema, softwareRecordBuilder, outDir, tables.Software)
	if err != nil {
		return err
	}

	err = writeRecords(tables.MentionsSchema, mentionsRecordBuilder, outDir, tables.Mentions)
	if err != nil {
		return err
	}

	return nil
}

type Paper struct {
	File string `json:"file"`
	Doi  string `json:"doi"`
	Year uint16 `json:"year"`
}

func extractPapers(reader io.Reader, outDir string, err error) error {
	papers := jsonio.NewReader(reader, func() *Paper {
		return &Paper{}
	})

	allocator := memory.NewGoAllocator()
	paperRecordBuilder := array.NewRecordBuilder(allocator, tables.PapersSchema)
	defer paperRecordBuilder.Release()

	uuidBuilder := paperRecordBuilder.Field(0).(*array.StringBuilder)
	doiBuilder := paperRecordBuilder.Field(1).(*array.StringBuilder)
	yearBuilder := paperRecordBuilder.Field(2).(*array.Uint16Builder)

	for paper, err := range papers.Read() {
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}

		uuidBuilder.Append(paper.File[:36])
		doiBuilder.Append(paper.Doi)
		yearBuilder.Append(paper.Year)
	}

	return writeRecords(tables.PapersSchema, paperRecordBuilder, outDir, tables.Papers)
}

func writeRecords(schema *arrow.Schema, recordBuilder *array.RecordBuilder, outDir, outTable string) error {
	record := recordBuilder.NewRecord()
	defer record.Release()

	outPath := filepath.Join(outDir, outTable+tables.ParquetExt)
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	// Don't close outFile; parquet handles closing it.
	writer, err := pqarrow.NewFileWriter(
		schema,
		outFile,
		parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Gzip)),
		pqarrow.DefaultWriterProps(),
	)
	if err != nil {
		return err
	}
	defer func() {
		err = writer.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	err = writer.Write(record)
	if err != nil {
		return err
	}

	return nil
}
