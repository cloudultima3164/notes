package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrEmptyHeader = errors.New("empty header")

type ErrInvalidHeader struct {
	line string
}

func (e ErrInvalidHeader) Error() string {
	return fmt.Sprintf("could not parse header line: %v", e.line)
}

//TODO: make parser a bit more robust, in particular we want it to be able to gracefully handle non-note text files
func ParseNote(reader io.Reader, path string, justHeader bool) (*Note, error) {
	in := bufio.NewReader(reader)
	var curLine string
	done := false
	isHeader := true
	buf := make([]byte, 2000)
	result := &Note{
		Path: path,
	}
	for !done {
		if isHeader {
			bytes, prefix, err := in.ReadLine()
			if errors.Is(err, io.EOF) {
				done = true
			}
			curLine += string(bytes)
			// prefix means it wasn't able to stick the full thing into the buffer
			if prefix {
				continue
			}
			if strings.TrimSpace(curLine) == DIVIDER {
				isHeader = false
				curLine = ""
				if len(result.rawHeader) == 0 {
					return nil, ErrEmptyHeader
				}
				continue
			}
			result.rawHeader += fmt.Sprintf("%v\n", curLine)
			headerData := strings.Split(curLine, ":")
			if len(headerData) < 2 {
				return nil, ErrInvalidHeader{line: curLine}
			}
			field := headerData[0]
			value := strings.Join(headerData[1:], ":")
			switch strings.TrimSpace(strings.ToLower(field)) {
			case "title":
				result.Title = strings.TrimSpace(value)
			case "tags":
				if len(value) == 0 {
					break
				}
				splitVal := strings.Split(value, ",")
				tags := make([]string, 0)
				for _, val := range splitVal {
					trimmed := strings.TrimSpace(val)
					if len(trimmed) == 0 {
						continue
					}
					contains := false
					for _, existingTag := range tags {
						if strings.EqualFold(trimmed, existingTag) {
							contains = true
						}
					}
					if !contains {
						tags = append(tags, trimmed)
					}
				}
				if len(tags) > 0 {
					result.Tags = tags
				}
			}
			curLine = ""
		} else if !justHeader {
			bytesRead, err := in.Read(buf)
			if errors.Is(err, io.EOF) {
				done = true
			}
			if bytesRead > 0 {
				curLine += string(buf[:bytesRead])
			}
		}
		if justHeader && !isHeader {
			done = true
		}

	}
	if !justHeader {
		result.Content = curLine
	}
	if len(strings.TrimSpace(result.Title)) == 0 {
		result.Title = path
	}
	return result, nil
}
