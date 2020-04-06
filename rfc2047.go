package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"mime"
	"os"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var (
	charsetMappings = map[string]encoding.Encoding{
		"gb2312": simplifiedchinese.GBK,
	}

	dec = mime.WordDecoder{
		CharsetReader: func(charset string, input io.Reader) (io.Reader, error) {
			for _, index := range []*ianaindex.Index{ianaindex.MIME, ianaindex.IANA, ianaindex.MIB} {
				enc, err := index.Encoding(charset)
				if err == nil && enc != nil {
					return enc.NewDecoder().Reader(input), nil
				}
			}

			if enc, ok := charsetMappings[charset]; ok {
				return enc.NewDecoder().Reader(input), nil
			}
			return nil, fmt.Errorf("unknown charset: %s", charset)
		},
	}
)

func main() {

	fs := flag.NewFlagSet("qp", flag.ExitOnError)
	dec := fs.Bool("D", false, "set this flag if you'd like to decode insteaed of encode")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(fs.Args()) == 0 {
		stdin := bufio.NewScanner(os.Stdin)

		for stdin.Scan() {
			input := strings.TrimSpace(stdin.Text())
			if *dec {
				fmt.Println(decode(input))
				continue
			}
			fmt.Println(encode(input))
		}
		if err := stdin.Err(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	input := strings.TrimSpace(os.Args[len(os.Args)-1])
	if *dec {
		fmt.Println(decode(input))
		return
	}
	encode(input)
	fmt.Println(encode(input))
}

func encode(input string) string {
	if strings.ContainsAny(input, "\"#$%&'(),.:;<>@[]^`{|}~") {
		return mime.BEncoding.Encode("utf-8", input)
	}
	return mime.QEncoding.Encode("utf-8", input)
}

func decode(input string) string {
	dec := mime.WordDecoder{}
	out, err := dec.Decode(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return out
}
