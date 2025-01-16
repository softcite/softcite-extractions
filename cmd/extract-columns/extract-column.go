package main

import (
	"compress/gzip"
	"encoding/csv"
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
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cmd = cobra.Command{
	Use:     "extract-columns [papers|software|latex|jats|grobid|pub2tei] IN_DIR OUT_DIR",
	Short:   "converts parts of the dataset into the Apache Parquet format",
	Args:    cobra.ExactArgs(3),
	Version: "0.1.0",
	RunE:    runE,
}

func runE(_ *cobra.Command, args []string) error {
	extractType := args[0]
	inPath := args[1]
	outDir := args[2]

	inFile, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer func() {
		err := inFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	statusBar, err := toReader(inPath, extractType)
	if err != nil {
		return fmt.Errorf("creating reader: %w", err)
	}

	// gzip correctly handles concatenated files.
	var reader io.Reader
	reader, err = gzip.NewReader(statusBar)
	if err != nil {
		return fmt.Errorf("creating gzip reader: %w", err)
	}

	switch extractType {
	case "papers":
		return extractPapers(statusBar, reader, outDir)
	default:
		err := extractMentions(reader, extractType, outDir)
		if err != nil {
			return fmt.Errorf("extracting %s mentions: %w", extractType, err)
		}
		return nil
	}
}

var (
	paperPattern   = regexp.MustCompile(`[0-9a-f]{2}\.jsonl.gz`)
	pdfPattern     = regexp.MustCompile(`[0-9a-f]{2}\.software\.jsonl\.gz`)
	latexPattern   = regexp.MustCompile(`[0-9a-f]{2}\.latex\.tei\.software\.jsonl\.gz`)
	jatsPattern    = regexp.MustCompile(`[0-9a-f]{2}\.jats\.software\.jsonl\.gz`)
	grobidPattern  = regexp.MustCompile(`[0-9a-f]{2}\.grobid\.tei\.software\.jsonl\.gz`)
	pub2teiPattern = regexp.MustCompile(`[0-9a-f]{2}\.pub2tei\.tei\.jsonl\.gz`)
)

func toReader(inPath, extractType string) (*fileio.StatusBarMultiReader, error) {
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

		var pattern *regexp.Regexp
		switch extractType {
		case "papers":
			pattern = paperPattern
		case "pdf":
			pattern = pdfPattern
		case "latex":
			pattern = latexPattern
		case "jats":
			pattern = jatsPattern
		case "grobid":
			pattern = grobidPattern
		case "pub2tei":
			pattern = pub2teiPattern
		default:
			return nil, fmt.Errorf("must be one of [papers|software|latex|jats|grobid|pub2tei], not %s", extractType)
		}

		for _, entry := range entries {
			if !pattern.MatchString(entry.Name()) {
				continue
			}

			entryPath := filepath.Join(inPath, entry.Name())
			inPaths = append(inPaths, entryPath)
		}
	} else {
		inPaths = append(inPaths, inPath)
	}

	r, err := fileio.NewStatusBarMultiReader(inPaths)
	if err != nil {
		return nil, fmt.Errorf("creating status bar multi reader: %w", err)
	}
	return r, nil
}

type SoftwareMentions struct {
	File     string            `json:"file"`
	Mentions []SoftwareMention `json:"mentions"`
}

type SoftwareMention struct {
	SoftwareName              SoftwareName      `json:"software-name"`
	Version                   Name              `json:"version"`
	Publisher                 Name              `json:"publisher"`
	Language                  Name              `json:"language"`
	URL                       Name              `json:"url"`
	Context                   string            `json:"context"`
	MentionContextAttributes  ContextAttributes `json:"mentionContextAttributes"`
	DocumentContextAttributes ContextAttributes `json:"documentContextAttributes"`
}

type ContextAttributes struct {
	Created ScoreValue `json:"created"`
	Shared  ScoreValue `json:"shared"`
	Used    ScoreValue `json:"used"`
}

type ScoreValue struct {
	Score float64 `json:"score"`
	Value bool    `json:"value"`
}

type SoftwareName struct {
	NormalizedForm string `json:"normalizedForm"`
	RawForm        string `json:"rawForm"`
	WikidataId     string `json:"wikidataId"`
}

type Name struct {
	NormalizedForm string `json:"normalizedForm"`
	RawForm        string `json:"rawForm"`
}

func extractMentions(reader io.Reader, extractType, outDir string) error {
	softwareMentions := jsonio.NewReader(reader, func() *SoftwareMentions {
		return &SoftwareMentions{}
	})

	allocator := memory.NewGoAllocator()

	softwareMentionsRecordBuilder := array.NewRecordBuilder(allocator, tables.SoftwareMentions)
	defer softwareMentionsRecordBuilder.Release()

	softwareMentionsFields := softwareMentionsRecordBuilder.Fields()
	softwareMentionIdField := softwareMentionsFields[0].(*array.StringBuilder)
	softwareMentionPaperIdField := softwareMentionsFields[1].(*array.Uint32Builder)
	softwareMentionSourceFileTypeField := softwareMentionsFields[2].(*array.BinaryDictionaryBuilder)
	softwareMentionIndexField := softwareMentionsFields[3].(*array.Uint16Builder)
	softwareMentionNameRawField := softwareMentionsFields[4].(*array.StringBuilder)
	softwareMentionNameNormalizedField := softwareMentionsFields[5].(*array.StringBuilder)
	softwareMentionVersionRawField := softwareMentionsFields[6].(*array.StringBuilder)
	softwareMentionVersionNormalizedField := softwareMentionsFields[7].(*array.StringBuilder)
	softwareMentionPublisherRawField := softwareMentionsFields[8].(*array.StringBuilder)
	softwareMentionPublisherNormalizedField := softwareMentionsFields[9].(*array.StringBuilder)
	softwareMentionLanguageRawField := softwareMentionsFields[10].(*array.StringBuilder)
	softwareMentionLanguageNormalizedField := softwareMentionsFields[11].(*array.StringBuilder)
	softwareMentionUrlRawField := softwareMentionsFields[12].(*array.StringBuilder)
	softwareMentionUrlNormalizedField := softwareMentionsFields[13].(*array.StringBuilder)
	softwareMentionContextField := softwareMentionsFields[14].(*array.StringBuilder)

	purposeAssessmentRecordBuilder := array.NewRecordBuilder(allocator, tables.PurposeAssessment)
	defer purposeAssessmentRecordBuilder.Release()

	purposeAssessmentFields := purposeAssessmentRecordBuilder.Fields()
	purposeAssessmentIdField := purposeAssessmentFields[0].(*array.StringBuilder)
	purposeAssessmentPaperIdField := purposeAssessmentFields[1].(*array.Uint32Builder)
	purposeAssessmentSourceFileTypeField := purposeAssessmentFields[2].(*array.BinaryDictionaryBuilder)
	purposeAssessmentIndexField := purposeAssessmentFields[3].(*array.Uint16Builder)
	purposeAssessmentScopeField := purposeAssessmentFields[4].(*array.BinaryDictionaryBuilder)
	purposeAssessmentPurposeField := purposeAssessmentFields[5].(*array.BinaryDictionaryBuilder)
	purposeAssessmentCertaintyField := purposeAssessmentFields[6].(*array.Float64Builder)

	// Mention writer logic
	mentionsPath := filepath.Join(outDir, tables.MentionsName+"."+extractType+tables.ParquetExt)
	mentionsFile, err := os.Create(mentionsPath)
	if err != nil {
		return fmt.Errorf("creating mentions file: %w", err)
	}
	// Don't close mentionsFile; parquet handles closing it.
	mentionsWriter, err := pqarrow.NewFileWriter(
		tables.SoftwareMentions,
		mentionsFile,
		parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Gzip),
			parquet.WithCompressionLevel(gzip.BestCompression)),
		pqarrow.DefaultWriterProps(),
	)
	if err != nil {
		return fmt.Errorf("creating mentions writer: %w", err)
	}

	defer func() {
		err := mentionsWriter.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// Purpose writer logic
	purposePath := filepath.Join(outDir, tables.PurposeAssessmentsName+"."+extractType+tables.ParquetExt)
	purposeFile, err := os.Create(purposePath)
	if err != nil {
		return fmt.Errorf("creating purpose file: %w", err)
	}
	// Don't close outFile; parquet handles closing it.
	purposeWriter, err := pqarrow.NewFileWriter(
		tables.PurposeAssessment,
		purposeFile,
		parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Gzip),
			parquet.WithCompressionLevel(gzip.BestCompression)),
		pqarrow.DefaultWriterProps(),
	)
	if err != nil {
		return fmt.Errorf("creating purpose writer: %w", err)
	}

	defer func() {
		err := purposeWriter.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	paperIdsPath := filepath.Join(outDir, paperIdsFileName)
	paperIdsFile, err := os.Open(paperIdsPath)
	if err != nil {
		return fmt.Errorf("opening paper ids file: %w", err)
	}
	defer func() {
		err := paperIdsFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	paperIdsReader := csv.NewReader(paperIdsFile)
	paperIdsReader.FieldsPerRecord = 2

	// 1a6eb5f1-b93b-4344-bdfd-d9c0710337a2

	paperIdRecord, err := paperIdsReader.Read()
	if err != nil {
		return fmt.Errorf("reading paper ids: %w", err)
	}

	paperIdMap := make(map[string]uint32)
	for ; ; paperIdRecord, err = paperIdsReader.Read() {
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		paperId, err := strconv.ParseUint(paperIdRecord[0], 10, 32)
		if err != nil {
			return fmt.Errorf("parsing paper id: %w", err)
		}

		paperIdMap[paperIdRecord[1]] = uint32(paperId)
	}
	fmt.Println("finished reading paper ids")

	// Loop
	i := 0
	for softwareMention, err := range softwareMentions.Read() {
		i++
		//if i > 1000 {
		//	break
		//}

		if i%100000 == 0 || errors.Is(err, io.EOF) {
			mentionRecord := softwareMentionsRecordBuilder.NewRecord()
			errWrite := mentionsWriter.Write(mentionRecord)
			if errWrite != nil {
				return fmt.Errorf("writing mentions: %w", errWrite)
			}
			mentionRecord.Release()

			purposeRecord := purposeAssessmentRecordBuilder.NewRecord()
			errWrite = purposeWriter.Write(purposeRecord)
			if errWrite != nil {
				return fmt.Errorf("writing purpose: %w", errWrite)
			}
			purposeRecord.Release()
		}

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("reading software mentions: %w", err)
			}
			break
		}

		for i, mention := range softwareMention.Mentions {
			// Ids
			softciteId := softwareMention.File[:36]

			paperId, found := paperIdMap[softciteId]
			if !found {
				fmt.Println("missing paper id for", softciteId)
			}

			softwareMentionId := fmt.Sprintf("%10d.%s.%05d", paperId, extractType, i)

			// Software Mention
			softwareMentionIdField.Append(softwareMentionId)
			softwareMentionPaperIdField.Append(uint32(paperId))
			err = softwareMentionSourceFileTypeField.AppendString(extractType)
			if err != nil {
				return fmt.Errorf("appending source file type: %w", err)
			}
			softwareMentionIndexField.Append(uint16(i))

			// Mentions
			softwareMentionNameRawField.Append(mention.SoftwareName.RawForm)
			softwareMentionNameNormalizedField.Append(mention.SoftwareName.NormalizedForm)

			if mention.Version.RawForm == "" {
				softwareMentionVersionRawField.AppendNull()
			} else {
				softwareMentionVersionRawField.Append(mention.Version.RawForm)
			}
			if mention.Version.NormalizedForm == "" {
				softwareMentionVersionNormalizedField.AppendNull()
			} else {
				softwareMentionVersionNormalizedField.Append(mention.Version.NormalizedForm)
			}

			if mention.Publisher.RawForm == "" {
				softwareMentionPublisherRawField.AppendNull()
			} else {
				softwareMentionPublisherRawField.Append(mention.Publisher.RawForm)
			}
			if mention.Publisher.NormalizedForm == "" {
				softwareMentionPublisherNormalizedField.AppendNull()
			} else {
				softwareMentionPublisherNormalizedField.Append(mention.Publisher.NormalizedForm)
			}

			if mention.Language.RawForm == "" {
				softwareMentionLanguageRawField.AppendNull()
			} else {
				softwareMentionLanguageRawField.Append(mention.Language.RawForm)
			}
			if mention.Language.NormalizedForm == "" {
				softwareMentionLanguageNormalizedField.AppendNull()
			} else {
				softwareMentionLanguageNormalizedField.Append(mention.Language.NormalizedForm)
			}

			if mention.URL.RawForm == "" {
				softwareMentionUrlRawField.AppendNull()
			} else {
				softwareMentionUrlRawField.Append(mention.URL.RawForm)
			}
			if mention.URL.NormalizedForm == "" {
				softwareMentionUrlNormalizedField.AppendNull()
			} else {
				softwareMentionUrlNormalizedField.Append(mention.URL.NormalizedForm)
			}

			if mention.Context == "" {
				softwareMentionContextField.AppendNull()
			} else {
				softwareMentionContextField.Append(mention.Context)
			}

			// Purpose Assessments
			for _, scope := range []string{"document", "local"} {
				var contextAttributes ContextAttributes
				switch scope {
				case "document":
					contextAttributes = mention.DocumentContextAttributes
				case "local":
					contextAttributes = mention.MentionContextAttributes
				default:
					panic("invalid scope " + scope)
				}

				for _, purpose := range []string{"created", "shared", "used"} {
					var purposeScoreValue ScoreValue
					switch purpose {
					case "created":
						purposeScoreValue = contextAttributes.Created
					case "shared":
						purposeScoreValue = contextAttributes.Shared
					case "used":
						purposeScoreValue = contextAttributes.Used
					default:
						panic("invalid purpose " + purpose)
					}

					//// Purpose Assessment
					purposeAssessmentIdField.Append(softwareMentionId)
					purposeAssessmentPaperIdField.Append(uint32(paperId))
					err = purposeAssessmentSourceFileTypeField.AppendString(extractType)
					if err != nil {
						return fmt.Errorf("appending source file type: %w", err)
					}
					purposeAssessmentIndexField.Append(uint16(i))
					err = purposeAssessmentScopeField.AppendString(scope)
					if err != nil {
						return fmt.Errorf("appending scope: %w", err)
					}
					err = purposeAssessmentPurposeField.AppendString(purpose)
					if err != nil {
						return fmt.Errorf("appending purpose: %w", err)
					}
					purposeAssessmentCertaintyField.Append(purposeScoreValue.Score)
				}
			}
		}
	}

	return nil
}

func panicIfEmptyString(s string) {
	if s == "" {
		panic("empty string")
	}
}

type Paper struct {
	File          string `json:"file"`
	Title         string `json:"title"`
	PublishedYear int    `json:"year"`
	PublishedDate string `json:"published_date"`
	JournalName   string `json:"journal_name"`
	PublisherName string `json:"publisher_name"`
	DOI           string `json:"doi"`
	PMCID         string `json:"pmcid"`
	PMID          string `json:"pmid"`
	Genre         string `json:"genre"`
	LicenseType   string `json:"license_type"`
}

const paperIdsFileName = "paper_ids.csv"

func extractPapers(statusBar *fileio.StatusBarMultiReader, reader io.Reader, outDir string) error {
	papers := jsonio.NewReader(reader, func() *Paper {
		return &Paper{}
	})

	schema := tables.Papers

	allocator := memory.NewGoAllocator()
	paperRecordBuilder := array.NewRecordBuilder(allocator, schema)
	defer paperRecordBuilder.Release()

	paperFields := paperRecordBuilder.Fields()
	paperIdField := paperFields[0].(*array.Uint32Builder)
	softciteIdField := paperFields[1].(*array.StringBuilder)
	titleField := paperFields[2].(*array.StringBuilder)
	yearField := paperFields[3].(*array.Uint16Builder)
	publishedDateField := paperFields[4].(*array.Date32Builder)
	journalNameField := paperFields[5].(*array.StringBuilder)
	publisherNameField := paperFields[6].(*array.StringBuilder)
	doiField := paperFields[7].(*array.StringBuilder)
	pmcidField := paperFields[8].(*array.StringBuilder)
	pmidField := paperFields[9].(*array.StringBuilder)
	genreField := paperFields[10].(*array.BinaryDictionaryBuilder)
	licenseTypeField := paperFields[11].(*array.BinaryDictionaryBuilder)

	paperId := uint32(0)

	paperIdsPath := filepath.Join(outDir, paperIdsFileName)
	paperIdsFile, err := os.Create(paperIdsPath)
	if err != nil {
		return fmt.Errorf("creating paper ids file: %w", err)
	}
	defer func() {
		err := paperIdsFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	paperIdsWriter := csv.NewWriter(paperIdsFile)

	i := 0

	for paper, err := range papers.Read() {
		i++
		if i%10000 == 0 {
			statusBar.UpdateStatusBar()
		}

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}

		paperId++
		paperIdField.Append(paperId)

		softciteId := paper.File[:36]
		softciteIdField.Append(softciteId)

		err = paperIdsWriter.Write([]string{fmt.Sprint(paperId), softciteId})
		if err != nil {
			return fmt.Errorf("writing paper id: %w", err)
		}

		if paper.Title == "" {
			titleField.AppendNull()
		} else {
			titleField.Append(paper.Title)
		}

		if paper.PublishedYear == 0 {
			yearField.AppendNull()
		} else {
			yearField.Append(uint16(paper.PublishedYear))
		}

		if paper.PublishedDate == "" {
			publishedDateField.AppendNull()
		} else {
			publishedDate, err := time.Parse("2006-01-02", paper.PublishedDate)
			if err != nil {
				return err
			}
			publishedDateField.Append(arrow.Date32FromTime(publishedDate))
		}

		if paper.JournalName == "" {
			journalNameField.AppendNull()
		} else {
			journalNameField.Append(paper.JournalName)
		}

		if paper.PublisherName == "" {
			publisherNameField.AppendNull()
		} else {
			publisherNameField.Append(paper.PublisherName)
		}

		doiField.Append(paper.DOI)

		if paper.PMCID == "" {
			pmcidField.AppendNull()
		} else {
			pmcidField.Append(paper.PMCID)
		}

		if paper.PMID == "" {
			pmidField.AppendNull()
		} else {
			pmidField.Append(paper.PMID)
		}

		if paper.Genre == "" {
			genreField.AppendNull()
		} else {
			err = genreField.AppendString(paper.Genre)
			if err != nil {
				return err
			}
		}

		if paper.LicenseType == "" {
			licenseTypeField.AppendNull()
		} else {
			err = licenseTypeField.AppendString(paper.LicenseType)
			if err != nil {
				return err
			}
		}
	}

	paperIdsWriter.Flush()
	if paperIdsWriter.Error() != nil {
		return fmt.Errorf("flushing paper ids: %w", paperIdsWriter.Error())
	}

	return writeRecords(schema, paperRecordBuilder, outDir, tables.PapersName)
}

func writeRecords(schema *arrow.Schema, recordBuilder *array.RecordBuilder, outDir, outTable string) error {
	outPath := filepath.Join(outDir, outTable+tables.ParquetExt)
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	// Don't close outFile; parquet handles closing it.
	writer, err := pqarrow.NewFileWriter(
		schema,
		outFile,
		parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Gzip),
			parquet.WithCompressionLevel(gzip.BestCompression)),
		pqarrow.DefaultWriterProps(),
	)
	if err != nil {
		return err
	}

	defer func() {
		err := writer.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	record := recordBuilder.NewRecord()
	defer record.Release()

	return writer.Write(record)
}
