package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

// These get overwritten at build time with goreleaser
var (
	build  = "dev"
	commit = "local"
	date   = "20XX-01-01"
)

func addTimestamp(file *os.File, path string, ts time.Time) error {
	note, err := ParseNote(file, path, false)
	if err != nil {
		return err
	}

	note.Content = fmt.Sprintf("%v:\n\n\n%v", ts.Format(JOURNAL_DATE_FORMAT), note.Content)
	return updateNoteFile(note, file)
	// out := []byte(fmt.Sprintf("%v%v\n\n%v", note.rawHeader, DIVIDER, note.Content))
	// file.Truncate(0)
	// wrote := 0
	// for written := 0; written <= len(out)-1; written += wrote {
	// 	wrote, err = file.Write(out[written:])

	// 	if err != nil {
	// 		return fmt.Errorf("problem writing to file: %w", err)
	// 	}
	// }
	// return nil
}

func insertValueAtPos[T any](slice *[]T, val T, pos int) {
	var new T
	// Add new element to slice
	*slice = append(*slice, new)
	dst := (*slice)[pos+1:]
	src := (*slice)[pos:]
	// shift elements to right of insert pos
	copy(dst, src)
	// insert value at pos
	(*slice)[pos] = val
}

func addTask(file *os.File, path string, details string) error {
	note, err := ParseNote(file, path, false)
	if err != nil {
		return err
	}

	var taskWriteLine int
	firstNewLine := strings.Index(note.Content, "\n")
	// We only care if newline is after a date or not
	// If the line is shorter than 10 characters, there is probably no date
	if firstNewLine > 10 {
		maybeDate := note.Content[:10]
		_, err := time.Parse("2006/01/02", maybeDate)
		if err != nil {
			taskWriteLine = 0
		} else {
			taskWriteLine = 1
		}
	} else {
		taskWriteLine = 0
	}
	// We're either adding the task on the first line or after the first newline
	newTask := fmt.Sprintf("-[]: %v\n", details)
	contents := strings.Split(note.Content, "\n")
	insertValueAtPos(&contents, newTask, taskWriteLine)
	note.Content = strings.Join(contents, "\n")
	return updateNoteFile(note, file)
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

// Will this work for adding interactive later?
// Per the comment in var newNoteCmd
// func prepareFileName(cmd *cobra.Command, args []string, wantExistance bool) error {
// if len(args) == 0 {
// 	return nil
// }
// preparedFileName, err := checkExistance(args[0], wantExistance)
// if err != nil {
// 	return err
// }
// args[0] = preparedFileName
// cmd.SetArgs(args)

// return nil
// }

func chooseFileInteractive(title string, headerOnly bool) (selectorModel, error) {
	mod, err := NewFileSelector(title, headerOnly)
	if err != nil {
		return selectorModel{}, fmt.Errorf("could not select a file: %v", err)
	}
	m, err := tea.NewProgram(mod).StartReturningModel()
	if err != nil {
		return selectorModel{}, fmt.Errorf("problem trying to get selection: %v", err)
	}
	model, ok := m.(selectorModel)
	if !ok {
		return selectorModel{}, fmt.Errorf("could not read selection")
	}
	return model, nil
}

// func openNote(args []string, interactive_title string, headerOnly bool) (*os.File, string, error) {
// 	var selectedFile string
// 	if len(args) == 0 {
// 		mod, err := chooseFileInteractive(interactive_title, headerOnly)
// 		if err != nil {
// 			return nil, "", err
// 		}
// 		selectedFile = mod.choice.Path
// 	} else {
// 		selectedFile = args[0]
// 	}
// 	file, err := os.OpenFile(selectedFile, os.O_RDWR|os.O_APPEND, os.ModePerm)
// 	if err != nil {
// 		return nil, "", fmt.Errorf("could not open file: %v, %v", args[0], err)
// 	}
// 	return file, selectedFile, nil
// }

func updateNoteFile(note *Note, file *os.File) error {
	out := []byte(fmt.Sprintf("%v%v\n\n%v", note.rawHeader, DIVIDER, note.Content))
	file.Truncate(0)
	wrote := 0
	var err error
	for written := 0; written <= len(out)-1; written += wrote {
		wrote, err = file.Write(out[written:])

		if err != nil {
			return fmt.Errorf("problem writing to file: %w", err)
		}
	}
	return nil
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
		go parseWorker(in, out, wg, true, false)
	}
	for _, fileName := range fileList {
		in <- fileName
	}
	close(in)

	wg.Wait()
	close(out)
	for result := range out {
		results := fuzzy.Find(searchTag, result.Tags)
		if results.Len() > 0 {
			fmt.Printf("%v : %v\n", result.Title, result.Path)
		}
	}

	return nil
}

func parseWorker(in chan string, out chan Note, wg *sync.WaitGroup, justHeader, outputErrors bool) {
	for filename := range in {
		f, err := os.Open(filename)
		if err != nil {
			if outputErrors {
				fmt.Printf("could not open file: %v", err)
			}
			continue
		}
		note, err := ParseNote(f, filename, justHeader)
		if err != nil {
			if outputErrors {
				fmt.Printf("could not parse file: %v", err)
			}
			f.Close()
			continue
		}
		f.Close()
		if note != nil {
			out <- *note
		}
	}
	wg.Done()
}

func collectFiles(justHeader, outputFileErrors bool) ([]Note, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get working directory: %w", err)
	}
	fileList := files(curDir)
	wg := &sync.WaitGroup{}

	wg.Add(10)
	in := make(chan string, 10)
	out := make(chan Note, len(fileList))
	for i := 0; i < 10; i++ {
		go parseWorker(in, out, wg, justHeader, outputFileErrors)
	}
	for _, fileName := range fileList {
		in <- fileName
	}
	close(in)

	wg.Wait()
	close(out)
	results := make([]Note, 0)
	for result := range out {
		results = append(results, result)
	}

	return results, nil
}

var taskCmd = &cobra.Command{
	Use:     "task",
	Example: "notes task [filepath] [task]",
	Short:   "output the contents of a note",
	Long:    "output the contents of a note. if no note is specified, it goes into an interactive mode to select a note.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		preparedFileName, err := checkExistance(args[0], true)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			// selectMod, err := NewFileSelector("Select File to Edit Tasks for", false)
			// if err != nil {
			// 	fmt.Printf("Could not select a file: %v", err)
			// 	return
			// }
			// m, err := tea.NewProgram(selectMod).StartReturningModel()
			// if err != nil {
			// 	fmt.Printf("Problem trying to get selection: %v", err)
			// 	return
			// }
			// model, ok := m.(selectorModel)
			// if !ok {
			// 	fmt.Println("Could not read selection")
			// 	return
			// }
			model, err := chooseFileInteractive("Select File to Edit Tasks for", false)
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			taskViewerMod, err := NewTaskViewer(model.choice)
			if err != nil {
				fmt.Printf("Problem updating task: %v\n", err)
				return
			}
			if err := tea.NewProgram(taskViewerMod).Start(); err != nil {
				fmt.Printf("Problem updating task: %v\n", err)
			}

			return

		}
		if err := CatNote(args[0]); err != nil {
			fmt.Printf("Problem trying to cat: %v", err)
		}
	},
}

var newTaskCmd = &cobra.Command{
	Use:     "new",
	Example: "notes task new [filepath] [task]",
	Short:   "append a new task to a note",
	Long:    "append a new task to a note. if no filepath or task details are provided,  it goes into an interactive mode to select a note and add details.",
	Args: func(cmd *cobra.Command, args []string) error {
		// return prepareFileName(cmd, args, true)
		if len(args) == 0 {
			return nil
		}
		preparedFileName, err := checkExistance(args[0], true)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		// file, path, err := openNote(args, "Select File to Add Task", false)
		var path string
		var details string
		if len(args) == 0 {
			pathModel, err := chooseFileInteractive("Select File to Add Task", false)
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			detailsModel := NewTaskAdd()
			if err := detailsModel.StartGetTaskDetails(); err != nil {
				fmt.Printf("Problem getting task details: %v\n", err)
				return
			}
			path = pathModel.choice.Path
			details = detailsModel.result
			if details == "" {
				fmt.Printf("Expected task details, but got nothing.\n")
				return
			}

		} else {
			if len(args) == 1 {
				fmt.Printf("Expected 2 arguments but got 1")
				return
			}
			path = args[0]
			details = args[1]
		}
		file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, os.ModePerm)
		if err != nil {
			fmt.Printf("Could not open file: %v, %v", path, err)
			return
		}
		defer file.Close()

		err = addTask(file, path, details)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
	},
}

var catCmd = &cobra.Command{
	Use:     "cat",
	Example: "notes cat [filepath]",
	Short:   "output the contents of a note",
	Long:    "output the contents of a note. if no note is specified, it goes into an interactive mode to select a note.",
	Args: func(cmd *cobra.Command, args []string) error {
		// return prepareFileName(cmd, args, true)
		if len(args) == 0 {
			return nil
		}
		preparedFileName, err := checkExistance(args[0], true)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			// mod, err := NewFileSelector("Select File to Print the content from", false)
			// if err != nil {
			// 	fmt.Printf("Could not select a file: %v", err)
			// 	return
			// }
			// m, err := tea.NewProgram(mod).StartReturningModel()
			// model, ok := m.(selectorModel)
			// if !ok {
			// 	fmt.Println("Could not read selection")
			// 	return
			// }
			mod, err := chooseFileInteractive("Select File to Output to Terminal", false)
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			fmt.Printf("%s", mod.choice.Content)
			return

		}
		if err := CatNote(args[0]); err != nil {
			fmt.Printf("Problem trying to cat: %v\n", err)
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
	Short:   "adds today's YYYY-MM-DD date at the top of the provided note",
	Long:    "adds today's YYYY-MM-DD date at the top of the provided note. if no note is specified, it goes into an interactive mode to select one.",
	Example: "notes entry [directory/file]",
	Args: func(cmd *cobra.Command, args []string) error {
		// return prepareFileName(cmd, args, true)
		if len(args) == 0 {
			return nil
		}
		preparedFileName, err := checkExistance(args[0], true)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		var selectedFile string
		if len(args) == 0 {
			mod, err := chooseFileInteractive("Select File to Add a Date Entry to", true)
			if err != nil {
				fmt.Printf("%v", err)
				return
			}
			// mod, err := NewFileSelector("Select File to Add a Date Entry to", true)
			// if err != nil {
			// 	fmt.Printf("Could not select a file: %v", err)
			// 	return
			// }
			// m, err := tea.NewProgram(mod).StartReturningModel()
			// if err != nil {
			// 	fmt.Printf("Problem trying to get selection: %v", err)
			// 	return
			// }
			// mod, ok := m.(selectorModel)
			// if !ok {
			// 	fmt.Println("Could not read selection")
			// 	return
			// }
			selectedFile = mod.choice.Path
		} else {
			selectedFile = args[0]
		}
		file, err := os.OpenFile(selectedFile, os.O_RDWR|os.O_APPEND, os.ModePerm)
		if err != nil {
			fmt.Printf("Could not open file: %v, %v", args[0], err)
			return
		}
		defer file.Close()

		err = addTimestamp(file, selectedFile, time.Now())
		if err != nil {
			fmt.Printf("Could not add timestamp to file: %v", err)
			return
		}
		fmt.Printf("Added %v entry line to %v", time.Now().Format(JOURNAL_DATE_FORMAT), selectedFile)

	},
}

var newNoteCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "creates a new note at the given path/name",
	Args: func(cmd *cobra.Command, args []string) error {
		// return prepareFileName(cmd, args, false)
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
		fmt.Printf("notes %v %v %v\n", build, commit, date)
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
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(newTaskCmd)
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
