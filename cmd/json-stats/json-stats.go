package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"github.com/willbeason/bondsmith"
	"github.com/willbeason/bondsmith/jsonio"
	"github.com/willbeason/software-mentions/pkg/jsonl"
	"golang.org/x/term"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const IncEvery = 1 << 10

func main() {
	cmd.Flags().String("out", "", "output file path (default: stdout)")

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cmd = cobra.Command{
	Use:     "json-stats FILE|DIR",
	Short:   "Collect statistics about keys and values in .jsonl files",
	Args:    cobra.ExactArgs(1),
	Version: "0.1.0",
	RunE:    runE,
}

var ErrJsonStats = errors.New("getting JSON statistics")

func runE(cmd *cobra.Command, args []string) error {
	inPath := args[0]

	f, err := os.Stat(inPath)
	if err != nil {
		return fmt.Errorf("%w: stat %q: %w", ErrJsonStats, inPath, err)
	}

	keyValueSets := make(map[string]jsonl.Field)

	pattern := `[0-9a-f]{2}\.software\.jsonl\.gz`

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return fmt.Errorf("%w: getting terminal size: %w", ErrJsonStats, err)
	}
	p := mpb.New(mpb.WithWidth(width))

	if f.IsDir() {
		matcher, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("%w: compiling pattern %q: %w", ErrJsonStats, pattern, err)
		}
		err = processDirectory(p, inPath, 0, matcher, keyValueSets)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(inPath, ".jsonl") || strings.HasSuffix(inPath, ".jsonl.gz") {
		err = processJsonFile(p, inPath, keyValueSets)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("%w: file %q is neither a directory nor a .jsonl file", ErrJsonStats, inPath)
	}

	paths := make([]string, len(keyValueSets))
	i := 0
	for path := range keyValueSets {
		paths[i] = path
		i++
	}
	sort.Strings(paths)

	outPath, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}

	outFile := os.Stdout
	if outPath != "" {
		outFile, err = os.Create(outPath)
		if err != nil {
			return err
		}
	}

	for _, path := range paths {
		field, ok := keyValueSets[path]
		if !ok {
			continue
		}
		_, err = fmt.Fprintf(outFile, "%s;%s\n", path, field)
		if err != nil {
			return err
		}
	}

	return nil
}

func filterNames(names []os.DirEntry, matcher *regexp.Regexp) []os.DirEntry {
	var result []os.DirEntry

	for _, name := range names {
		if matcher.MatchString(name.Name()) {
			result = append(result, name)
		}
	}

	return result
}

func processDirectory(p *mpb.Progress, inPath string, depth int, matcher *regexp.Regexp, keyValueSets map[string]jsonl.Field) error {
	names, err := os.ReadDir(inPath)
	if err != nil {
		return fmt.Errorf("%w: stat %q: %w", ErrJsonStats, inPath, err)
	}

	pathName := filepath.Base(inPath)

	names = filterNames(names, matcher)

	var bar *mpb.Bar
	var now time.Time
	if depth < 2 {
		bar = p.AddBar(int64(len(names)),
			mpb.AppendDecorators(decor.AverageETA(decor.ET_STYLE_GO)),
			mpb.PrependDecorators(decor.Name(pathName)),
			mpb.PrependDecorators(decor.CountersNoUnit("%d/%d", decor.WCSyncSpace)),
			mpb.BarRemoveOnComplete())
		now = time.Now()
	}

	sort.Slice(names, func(i, j int) bool {
		return names[i].Name() < names[j].Name()
	})

	for _, name := range names {
		entryPath := filepath.Join(inPath, name.Name())
		if name.IsDir() {
			err = processDirectory(p, entryPath, depth+1, matcher, keyValueSets)
			if err != nil {
				return err
			}
		} else if matcher != nil {
			err = processJsonFile(p, entryPath, keyValueSets)
			if err != nil {
				return err
			}
		} else {
			err = processJsonFile(p, entryPath, keyValueSets)
			if err != nil {
				return err
			}
		}

		if bar != nil {
			bar.IncrBy(1, time.Since(now))
		}
	}

	return nil
}

func processJsonFile(p *mpb.Progress, inPath string, keyValueSets map[string]jsonl.Field) error {
	file, err := os.Open(inPath)
	if err != nil {
		return fmt.Errorf("%w: opening %q: %w", ErrJsonStats, inPath, err)
	}

	stat, err := os.Stat(inPath)
	if err != nil {
		return fmt.Errorf("%w: getting stat for %q: %w", ErrJsonStats, inPath, err)
	}

	countReader := bondsmith.NewCountReader(file)
	var reader io.Reader = countReader
	if strings.HasSuffix(inPath, ".gz") {
		reader, err = gzip.NewReader(countReader)
		if err != nil {
			return fmt.Errorf("%w: starting gzip reader stream for %q: %w", ErrJsonStats, inPath, err)
		}
	}

	entries := jsonio.NewReader(reader, func() *map[string]any {
		v := make(map[string]any)
		return &v
	})

	bar := p.AddBar(stat.Size(),
		mpb.AppendDecorators(decor.AverageETA(decor.ET_STYLE_GO)),
		mpb.PrependDecorators(decor.Name(filepath.Base(inPath))),
		mpb.BarRemoveOnComplete(),
	)

	i := 0
	lastSeen := 0
	start := time.Now()
	for entry, err := range entries.Read() {
		createdMap = make(map[string]float64)
		sharedMap = make(map[string]float64)
		usedMap = make(map[string]float64)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		currentDocument = (*entry)["file"].(string)

		if currentDocument == "883eb47b-b85e-4f3f-a7dc-878bdd9b0471.software.json" {
			s, _ := json.MarshalIndent(entry, "", "  ")
			f, err := os.Create("out.txt")
			if err != nil {
				return err
			}
			_, err = f.WriteString(string(s))
			if err != nil {
				return err
			}
			_ = f.Close()
			return errors.New("stop")
		}
		err = addKVs("", *entry, keyValueSets)
		if err != nil {
			uuid := (*entry)["file"].(string)
			return fmt.Errorf("%w: processing %q: %w", ErrJsonStats, uuid, err)
		}

		i++
		if i%IncEvery == 0 {
			curProgress := int(countReader.Count())
			bar.IncrBy(curProgress-lastSeen, time.Since(start))

			lastSeen = curProgress
		}
	}
	bar.IncrBy(int(countReader.Count())-lastSeen, time.Since(start))

	return nil
}

var currentDocument string
var createdMap map[string]float64
var sharedMap map[string]float64
var usedMap map[string]float64

func addKVs(path string, obj interface{}, kvs map[string]jsonl.Field) error {
	if path == ".mentions[]" {
		softwareName := obj.(map[string]interface{})["software-name"].(map[string]interface{})["normalizedForm"].(string)
		createdScore := obj.(map[string]interface{})["documentContextAttributes"].(map[string]interface{})["created"].(map[string]interface{})["score"].(float64)

		gotScore, ok := createdMap[softwareName]
		if !ok {
			createdMap[softwareName] = createdScore
		} else {
			if gotScore != createdScore {
				fmt.Println(fmt.Errorf("created score mismatch for %q %q: %f != %f", currentDocument, softwareName, gotScore, createdScore))
				fmt.Println()
			}
		}
	}
	if path == ".mentions[]" {
		softwareName := obj.(map[string]interface{})["software-name"].(map[string]interface{})["normalizedForm"].(string)
		createdScore := obj.(map[string]interface{})["documentContextAttributes"].(map[string]interface{})["shared"].(map[string]interface{})["score"].(float64)

		gotScore, ok := sharedMap[softwareName]
		if !ok {
			sharedMap[softwareName] = createdScore
		} else {
			if gotScore != createdScore {
				fmt.Println(fmt.Errorf("shared score mismatch for %q %q: %f != %f", currentDocument, softwareName, gotScore, createdScore))
				fmt.Println()
			}
		}
	}
	if path == ".mentions[]" {
		softwareName := obj.(map[string]interface{})["software-name"].(map[string]interface{})["normalizedForm"].(string)
		createdScore := obj.(map[string]interface{})["documentContextAttributes"].(map[string]interface{})["used"].(map[string]interface{})["score"].(float64)

		gotScore, ok := usedMap[softwareName]
		if !ok {
			usedMap[softwareName] = createdScore
		} else {
			if gotScore != createdScore {
				fmt.Println(fmt.Errorf("used score mismatch for %q %q: %f != %f", currentDocument, softwareName, gotScore, createdScore))
				fmt.Println()
			}
		}
	}

	switch o := obj.(type) {
	case []interface{}:
		for _, v := range o {
			err := addKVs(fmt.Sprintf("%s[]", path), v, kvs)
			if err != nil {
				return err
			}
		}
	case map[string]interface{}:
		for k, v := range o {
			err := addKVs(fmt.Sprintf("%s.%s", path, k), v, kvs)
			if err != nil {
				return err
			}
		}
	default:
		pathCounts := kvs[path]
		if pathCounts == nil {
			pathCounts = &jsonl.EmptyField{}
		}

		pathCounts, err := pathCounts.Add(obj)
		if err != nil {
			return err
		}
		kvs[path] = pathCounts
	}

	return nil
}
