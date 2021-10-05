package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/textproto"
	"os"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func main() {
	fs := flag.NewFlagSet("rfc2047", flag.ExitOnError)
	dec := fs.Bool("D", false, "set this flag if you'd like to decode insteaed of encode")
	header := fs.Bool("H", false, "set this flag if you're decoding an entire header -H implies -D")
	charset := fs.String("c", "utf-8", "set the charset you'd like to encode to")
	forceQ := fs.Bool("Q", false, "force using Q encoding")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// -H implies -D
	if *header {
		*dec = true
	}

	if len(fs.Args()) == 0 && !*header {
		stdin := bufio.NewScanner(os.Stdin)

		for stdin.Scan() {
			input := strings.TrimSpace(stdin.Text())
			if *dec {
				fmt.Println(decode(input))

				continue
			}
			fmt.Println(encode(input, *charset, *forceQ))
		}
		if err := stdin.Err(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		return
	} else if len(fs.Args()) == 0 && *header {
		rd := textproto.NewReader(bufio.NewReader(os.Stdin))
		hds, err := rd.ReadMIMEHeader()
		if err != nil && err != io.EOF {
			fmt.Println(err)
			os.Exit(1)
		}
		for h, hvals := range hds {
			for _, hval := range hvals {
				fmt.Println(h, ":", decodeHeader(hval))
			}
		}

		return
	}

	input := strings.TrimSpace(os.Args[len(os.Args)-1])
	if *dec {
		fmt.Println(decode(input))

		return
	}
	fmt.Println(encode(input, *charset, *forceQ))
}

func encode(input, charset string, forceQ bool) string {
	if strings.ContainsAny(input, "\"#$%&'(),.:;<>@[]^`{|}~") && !forceQ {
		return mime.BEncoding.Encode(charset, input)
	}

	return mime.QEncoding.Encode(charset, input)
}

func decodeHeader(input string) string {
	charsetMappings := map[string]encoding.Encoding{
		"gb2312": simplifiedchinese.GBK,
	}

	dec := mime.WordDecoder{
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
	dh, err := dec.DecodeHeader(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return dh
}

func decode(input string) string {
	charsetMappings := map[string]encoding.Encoding{
		"gb2312": simplifiedchinese.GBK,
	}

	dec := mime.WordDecoder{
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
	out, err := dec.Decode(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return out
}
