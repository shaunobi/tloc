package tokenizer

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"path"
	"strconv"
	"sync"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

const (
	o200kAssetName     = "o200k_base.tiktoken"
	o200kExpectedRanks = 199998
)

// o200kRanks is the canonical OpenAI o200k_base mergeable-ranks file. Keeping
// it in the binary ensures tokenizer construction never reads a cache or makes
// a network request at runtime.
//
//go:embed data/o200k_base.tiktoken
var o200kRanks []byte

type embeddedBPELoader struct{}

func (embeddedBPELoader) LoadTiktokenBpe(source string) (map[string]int, error) {
	if path.Base(source) != o200kAssetName {
		return nil, fmt.Errorf("embedded BPE loader has no asset for %q", source)
	}

	ranks := make(map[string]int, o200kExpectedRanks)
	scanner := bufio.NewScanner(bytes.NewReader(o200kRanks))
	// Rank lines are short, but use a generous ceiling so a corrupt file yields
	// a useful parser error instead of Scanner's small default limit.
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		separator := bytes.IndexByte(line, ' ')
		if separator <= 0 || separator == len(line)-1 {
			return nil, fmt.Errorf("parse %s line %d: expected base64 token and rank", o200kAssetName, lineNumber)
		}

		encodedToken := line[:separator]
		rankText := bytes.TrimSpace(line[separator+1:])
		decodedToken := make([]byte, base64.StdEncoding.DecodedLen(len(encodedToken)))
		n, err := base64.StdEncoding.Decode(decodedToken, encodedToken)
		if err != nil {
			return nil, fmt.Errorf("parse %s line %d token: %w", o200kAssetName, lineNumber, err)
		}
		rank, err := strconv.Atoi(string(rankText))
		if err != nil {
			return nil, fmt.Errorf("parse %s line %d rank: %w", o200kAssetName, lineNumber, err)
		}
		ranks[string(decodedToken[:n])] = rank
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", o200kAssetName, err)
	}
	if len(ranks) != o200kExpectedRanks {
		return nil, fmt.Errorf("parse %s: got %d unique ranks, want %d", o200kAssetName, len(ranks), o200kExpectedRanks)
	}
	return ranks, nil
}

type o200kCounter struct {
	encoding *tiktoken.Tiktoken
}

var (
	o200kOnce     sync.Once
	o200kEncoding *tiktoken.Tiktoken
	o200kInitErr  error
)

func newO200KCounter() (*o200kCounter, error) {
	o200kOnce.Do(func() {
		// tiktoken-go's loader is process-global and its encoding cache keeps the
		// first loaded ranks. Configure and initialize it once before workers run.
		tiktoken.SetBpeLoader(embeddedBPELoader{})
		o200kEncoding, o200kInitErr = tiktoken.GetEncoding(tiktoken.MODEL_O200K_BASE)
	})
	if o200kInitErr != nil {
		return nil, o200kInitErr
	}
	return &o200kCounter{encoding: o200kEncoding}, nil
}

func (c *o200kCounter) Count(content []byte) (int64, error) {
	// EncodeOrdinary treats strings such as <|endoftext|> as source text rather
	// than control tokens and cannot panic on disallowed special-token text.
	return int64(len(c.encoding.EncodeOrdinary(string(content)))), nil
}
