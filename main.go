package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"
)

var build = "dev"

const DIVIDER = "------"
const JOURNAL_DATE_FORMAT = "2006-01-02"

type Note struct {
	Path    string
	Title   string
	Tags    []string
	Content string
	// header may include metadata that's not necessarily tracked in this struct
	rawHeader string
}

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
				continue
			}
			result.rawHeader += fmt.Sprintf("%v\n", curLine)
			headerData := strings.Split(curLine, ":")
			if len(headerData) < 2 {
				fmt.Println(result.rawHeader)
				return nil, fmt.Errorf("could not parse header line: %v", curLine)
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
						if fuzzy.MatchNormalizedFold(trimmed, existingTag) {
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
	return result, nil
}

func addTimestamp(file *os.File, path string, ts time.Time) error {
	note, err := ParseNote(file, path, false)
	if err != nil {
		return err
	}

	note.Content = fmt.Sprintf("%v:\n\n\n%v", ts.Format(JOURNAL_DATE_FORMAT), note.Content)
	out := []byte(fmt.Sprintf("%v%v\n\n%v", note.rawHeader, DIVIDER, note.Content))
	file.Truncate(0)
	wrote := 0
	for written := 0; written <= len(out)-1; written += wrote {
		wrote, err = file.Write(out[written:])

		if err != nil {
			return fmt.Errorf("problem writing to file: %w", err)
		}
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func NewNoteFile(filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0770); err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file (at path: %v): %w", filePath, err)
	}
	initialName := path.Base(filePath)

	fmt.Fprintf(f, `title: %v
tags:
%v
`, strings.TrimSuffix(initialName, ".txt"), DIVIDER,
	)
	defer f.Close()
	return nil
}

func checkExistance(userInput string, wantExistance bool) (string, error) {
	if len(userInput) == 0 {
		return "", fmt.Errorf("empty filename")
	}
	if !strings.HasSuffix(userInput, ".txt") {
		userInput += ".txt"
	}
	curDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get working directory: %w", err)
	}
	outPath := path.Join(curDir, userInput)
	if exists(outPath) != wantExistance {
		var message string
		if wantExistance {
			message = "file `%v` does not exist"
		} else {
			message = "file `%v` already exists"
		}
		return "", fmt.Errorf(message, outPath)
	}
	return outPath, nil
}

// TODO: hmmm maybe a validate files command?
//func CheckNote(c *cli.Context) error {
//	name := c.Args().First()
//	outPath, err := checkExistance(name, true)
//	if err != nil {
//		return err
//	}
//
//	file, err := os.Open(outPath)
//	if err != nil {
//		return fmt.Errorf("could not open file: %v, %w", outPath, err)
//	}
//	defer file.Close()
//
//	note, err := ParseNote(file, outPath, false)
//	if err != nil {
//		return fmt.Errorf("could not parse file: %w", err)
//	}
//	fmt.Printf("loaded file: \n「%v」\n%+q\n%v\n", note.Title, note.Tags, note.Content)
//
//	return nil
//}

func CatNote(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file: %v, %w", filePath, err)
	}
	defer file.Close()

	note, err := ParseNote(file, filePath, false)
	if err != nil {
		return fmt.Errorf("could not parse file: %w", err)
	}
	fmt.Printf("%v", note.Content)

	return nil
}

func files(dir string) []string {
	results := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".txt") {
			results = append(results, path)
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return results
}

func checkTags(in chan string, out chan Note, wg *sync.WaitGroup, targetTag string) {
	for filename := range in {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Printf("could not open file: %v", err)
			continue
		}
		note, err := ParseNote(f, filename, true)
		if err != nil {
			fmt.Printf("could not parse file: %v", err)
			continue
		}
		f.Close()

		for _, val := range note.Tags {
			if strings.Contains(strings.ToLower(val), strings.ToLower(targetTag)) {
				out <- *note
				continue
			}
		}

	}
	wg.Done()
}

func CheckTags(input []string) error {
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}
	fileList := files(curDir)
	wg := &sync.WaitGroup{}
	searchTag := strings.Join(input, " ")

	wg.Add(10)
	in := make(chan string, 10)
	out := make(chan Note, len(fileList))
	for i := 0; i < 10; i++ {
		go checkTags(in, out, wg, searchTag)
	}
	for _, fileName := range fileList {
		in <- fileName
	}
	close(in)

	wg.Wait()
	close(out)
	for result := range out {
		fmt.Printf("%v : %v\n", result.Title, result.Path)
	}

	return nil
}

var catCmd = &cobra.Command{
	Use:     "cat",
	Example: "notes cat <filepath>",
	Short:   "output the contents of a file",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("got an unexpected number of args (%v), expected %v", len(args), 1)
		}
		// TODO: if no filename is provided go into an interactive mode
		preparedFileName, err := checkExistance(args[0], true)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		if err := CatNote(args[0]); err != nil {
			fmt.Printf("Problem trying to cat: %v", err)
		}
	},
}

var checkTagsCmd = &cobra.Command{
	Use:   "tagged",
	Short: "lists out files that match the tag",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if err := CheckTags(args[0:]); err != nil {
			fmt.Printf("Problem trying to check for tags: %v", err)
		}
	},
}

var newEntryCmd = &cobra.Command{
	Use:     "entry",
	Aliases: []string{"e"},
	Short:   "adds a YYYY-MM-DD date at the top of the provided note",
	Example: "notes entry <directory/file>",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("got an unexpected number of args (%v), expected %v", len(args), 1)
		}
		// TODO: if no filename is provided go into an interactive mode
		preparedFileName, err := checkExistance(args[0], true)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		file, err := os.OpenFile(args[0], os.O_RDWR|os.O_APPEND, os.ModePerm)
		if err != nil {
			fmt.Printf("Could not open file: %v, %v", args[0], err)
		}
		defer file.Close()

		err = addTimestamp(file, args[0], time.Now())
		if err != nil {
			fmt.Printf("Could not add timestamp to file: %v", err)
		}
	},
}

var newNoteCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "creates a new note at the given path/name",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("got an unexpected number of args (%v), expected %v", len(args), 1)
		}
		// TODO: if no filename is provided go into an interactive mode
		preparedFileName, err := checkExistance(args[0], false)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		if err := NewNoteFile(args[0]); err != nil {
			fmt.Printf("Problem trying to cat: %v", err)
		}
	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("notes %v\n", build)
	},
}

var rootCmd = &cobra.Command{
	Use:   "notes",
	Short: "Notes is a cli toolbox for plain text notes",
	Long: `A cli toolbox for creating and managing plain text notes. 
	all files are .txt so you do not need to specify .txt in the cli`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(checkTagsCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(newNoteCmd)
	rootCmd.AddCommand(newEntryCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
