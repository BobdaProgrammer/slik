//A terminal based text editor

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
	"golang.design/x/clipboard"
)

var filename string = ""

type ColorMapping struct {
	Comments       KeywordColor `json:"comments"`
	Strings        KeywordColor `json:"strings"`
	Keywords       KeywordColor `json:"keywords"`
	Statements     KeywordColor `json:"statements"`
	Types          KeywordColor `json:"types"`
	Operators      KeywordColor `json:"operators"`
	Brackets       KeywordColor `json:"brackets"`
	Declarations   KeywordColor `json:"declarations"`
	FnDeclarations KeywordColor `json:"FnDeclarations"`
}

type KeywordColor struct {
	Color ColorDetails `json:"color"`
}

type ColorDetails struct {
	Color string `json:"color"`
}

var wordList = map[string]string{
	"if":        "statement",
	"else":      "statement",
	"switch":    "statement",
	"case":      "statement",
	"default":   "statement",
	"for":       "statement",
	"while":     "statement",
	"do":        "statement",
	"break":     "statement",
	"continue":  "statement",
	"return":    "statement",
	"var":       "declaration",
	"let":       "declaration",
	"const":     "declaration",
	"function":  "FnDeclaration",
	"func":      "FnDeclaration",
	"class":     "declaration",
	"interface": "declaration",
	"type":      "keyword",
	"import":    "keyword",
	"package":   "keyword",
	"struct":    "keyword",
	"int":       "type",
	"string":    "type",
	"bool":      "type",
	"float64":   "type",
	"float32":   "type",
	"array":     "type",
	"map":       "type",
	"(":         "bracket",
	")":         "bracket",
	"{":         "bracket",
	"}":         "bracket",
	"[":         "type",
	"]":         "type",
	"+":         "operator",
	"-":         "operator",
	"/":         "operator",
	"*":         "operator",
	"=":         "operator",
	"!":         "operator",
	":":         "operator",
}

var colors = map[string]termbox.Attribute{
	"comments":      termbox.ColorGreen,
	"strings":       termbox.ColorCyan,
	"statement":     termbox.ColorMagenta,
	"bracket":       termbox.ColorMagenta,
	"FnDeclaration": termbox.ColorBlue,
	"operator":      termbox.ColorBlue,
	"declaration":   termbox.ColorCyan,
	"keyword":       termbox.ColorYellow,
	"type":          termbox.ColorYellow,
}

var ColorToAttrib = map[string]termbox.Attribute{
	"Magenta":       termbox.ColorMagenta,
	"Yellow":        termbox.ColorYellow,
	"Blue":          termbox.ColorBlue,
	"Cyan":          termbox.ColorCyan,
	"Green":         termbox.ColorGreen,
	"Red":           termbox.ColorRed,
	"White":         termbox.ColorWhite,
	"Black":         termbox.ColorBlack,
	"Default":       termbox.ColorDefault,
	"BrightRed":     termbox.ColorRed | termbox.AttrBold,
	"BrightGreen":   termbox.ColorGreen | termbox.AttrBold,
	"BrightYellow":  termbox.ColorYellow | termbox.AttrBold,
	"BrightBlue":    termbox.ColorBlue | termbox.AttrBold,
	"BrightMagenta": termbox.ColorMagenta | termbox.AttrBold,
	"BrightCyan":    termbox.ColorCyan | termbox.AttrBold,
	"BrightWhite":   termbox.ColorWhite | termbox.AttrBold,
}

func loadConfig() {
	jsonData, _ := ioutil.ReadFile("config.json")

	// Parse the JSON data into the ColorMapping struct
	var colorMapping ColorMapping
	err := json.Unmarshal(jsonData, &colorMapping)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	// Create a map with the types and their colors
	colors = map[string]termbox.Attribute{
		"comments":      ColorToAttrib[colorMapping.Comments.Color.Color],
		"strings":       ColorToAttrib[colorMapping.Strings.Color.Color],
		"keyword":       ColorToAttrib[colorMapping.Keywords.Color.Color],
		"statement":     ColorToAttrib[colorMapping.Statements.Color.Color],
		"type":          ColorToAttrib[colorMapping.Types.Color.Color],
		"operator":      ColorToAttrib[colorMapping.Operators.Color.Color],
		"bracket":       ColorToAttrib[colorMapping.Brackets.Color.Color],
		"declaration":   ColorToAttrib[colorMapping.Declarations.Color.Color],
		"FnDeclaration": ColorToAttrib[colorMapping.FnDeclarations.Color.Color],
	}
}

// defining the structure of the text editor
type Action struct {
	CursorX    int
	CursorXEND int
	CursorY    int
	CursorYEND int
	Text       string
	remove     bool
}

type Editor struct {
	buffer     []string
	UndoBuffer []Action
	RedoBuffer []Action
	cursorX    int
	cursorY    int
	offsetX    int
	offsetY    int
	width      int
	height     int
}

// creating the editor
func NewEditor() *Editor {
	var width, height int = termbox.Size()
	width -= 7
	height -= 1
	return &Editor{
		buffer:     []string{""},
		UndoBuffer: []Action{},
		RedoBuffer: []Action{},
		cursorX:    0,
		cursorY:    0,
		offsetX:    0,
		offsetY:    0,
		width:      width,
		height:     height,
	}
}

// Saving the text to the file
func (e *Editor) SaveFile() {
	if filename != "" {
		//if the file already exists just truncate
		file, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		//join all the lines into one string and writing to the file
		text := strings.Join(e.buffer, "\n")
		_, err = file.WriteString(text)
		if err != nil {
			panic(err)
		}

	} else {
		//if no file is specified then create untitled.txt
		file, err := os.Create("untitled.txt")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		text := strings.Join(e.buffer, "\n")
		_, err = file.WriteString(text)
		if err != nil {
			panic(err)
		}
	}
}

// adding text from file to screen
func (e *Editor) writeEditor(data string) {
	data = strings.Replace(data, "\t", "    ", -1)
	//turn data from string into an array and display to screen
	items := strings.Split(data, "\n")
	e.buffer = items
	termbox.Flush()
}

// reading file from command line arguments
func (e *Editor) ReadFile(filename string) {
	// Try to read the file
	data, err := ioutil.ReadFile(filename)
	if err != nil {

		// If file does not exist, create a new one
		if os.IsNotExist(err) {
			// Create the file
			file, err := os.Create(filename)
			if err != nil {
				return
			}
			defer file.Close()
			// Write some initial data to the file if needed
			initialData := []byte("")
			_, err = file.Write(initialData)
			if err != nil {
				return
			}
		}
		return
	}
	// File was read successfully, proceed with processing
	e.writeEditor(string(data))
}

func includes(line string, target string) (bool, int) {
	if strings.Contains(line, target) {
		place := strings.Index(line, target)
		var inQuote bool = false
		for n, c := range line {
			if c == '"' && inQuote == false {
				inQuote = true
			} else if c == '"' {
				inQuote = false
			}

			if c == '/' && line[n+1] == '/' && !inQuote {
				return true, place
			}
		}
		return false, 0
	}
	return false, 0
}

func includesStr(line string, index int) (bool, int, int) {
	startIndex := 0
	if line[index] != '"' && line[index] != '`' && line[index] != '\'' {
		// Find the start index of the substring starting at the given index
		substring := line[:index]
		// Reverse the substring
		lookBack := reverseString(substring)
		startIndex = strings.Index(lookBack, `"`)
		if startIndex == -1 {
			startIndex = strings.Index(lookBack, `'`)
			if startIndex == -1 {
				startIndex = strings.Index(lookBack, "`")
				if startIndex == -1 {
					return false, 0, 0
				}
			}
		}
		startIndex = (len(lookBack) - startIndex) - 1

		// Find the end index of the substring
		endIndex := strings.Index(line[index+1:], `"`)
		if endIndex == -1 {
			endIndex = strings.Index(line[index+1:], `'`)
			if endIndex == -1 {
				endIndex = strings.Index(line[index+1:], "`")
				if endIndex == -1 {
					return false, 0, 0
				}
			}
		}
		endIndex += index + 1 // Adjust for the slice

		quoteCount := 0
		for i, char := range line {
			if char == '"' || char == '`' || char == '\'' {
				quoteCount++
			}
			if i == endIndex && quoteCount%2 == 0 {
				return true, startIndex, endIndex
			}
		}
		line = line[:startIndex] + line[startIndex+1:]
		line = line[:endIndex] + line[endIndex+1:]
		return includesStr(line, index)
	} else {
		return true, index, index
	}
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func SyntaxHighlight(word string, index int, line string, bracket, point bool, wordType string) termbox.Attribute {
	isString, where2, where3 := includesStr(line, index)
	if isString == true && ((where2 <= index && where3 >= index-1) || where2 == where3) {
		return colors["strings"]
	}
	iscomment, where := includes(line, "//")
	if iscomment == true && where <= index {
		return colors["comments"]
	}

	switch wordType {
	case "bracket":
		return colors["bracket"]
	case "statement":
		return colors["statement"]
	case "FnDeclaration":
		return colors["FnDeclaration"]
	case "operator":
		return colors["operator"]
	case "declaration":
		// Highlight the word if it matches any of the cases
		return colors["declaration"]
	case "keyword":
		return colors["keyword"]
	case "type":
		return colors["type"]
	default:
		if bracket {
			return termbox.ColorYellow
		} else if point {
			return termbox.ColorCyan
		}
		// Return the default color if none of the cases match
		return termbox.ColorDefault
	}
}
func inUnn(unn []string, char string) bool {
	for _, c := range unn {
		if c == char {
			return true
		}
	}
	return false
}
func getWord(line string, index int) (string, bool, bool, string) {
	var UnnAcceptable = []string{"(", ")", ".", " ", "+", "-", "/", "=", "!", "{", "}", "[", "]"}
	var isUnn = inUnn(UnnAcceptable, string(line[index]))
	switch isUnn {
	case true:
		return string(line[index]), false, false, wordList[string(line[index])]
	default:

		var endWord string = ""
		var bracket bool = false
		var point bool = false
		for _, c := range line[index:] {
			if !inUnn(UnnAcceptable, string(c)) {
				endWord += string(c)
			} else {
				if c == '(' {
					bracket = true
				} else if c == '.' {
					point = true
				}
				break
			}
		}
		var forWord string = ""
		var word string = ""
		var RightWord bool = false
		for n, bc := range line[:index] {
			if inUnn(UnnAcceptable, string(bc)) {
				if RightWord {
					forWord = word
					break
				}
				word = ""
			} else {
				word += string(bc)
				if n == index {
					RightWord = true
				}
			}
		}
		forWord = word
		fullWord := forWord + endWord
		WordType := wordList[fullWord]
		return fullWord, bracket, point, WordType
	}
}

func (editor *Editor) StatBar(index int) rune {
	lineNumber := editor.cursorY + editor.offsetY + 1        // Adding  1 because line numbers start from  1
	columnNumber := editor.cursorX + editor.offsetX + 1      // Adding  1 because column numbers start from  1
	formattedLineNumber := fmt.Sprintf("%d", lineNumber)     // Format line number with leading zeros
	formattedColumnNumber := fmt.Sprintf("%d", columnNumber) // Format column number with leading zeros
	bar := "ln: " + formattedLineNumber + " | col: " + formattedColumnNumber + " | " + filename
	for len(bar) != editor.width {
		bar += " "
	}
	return rune(bar[index])
}

// rendering the text on the screen
func (e *Editor) Render() {
	e.width, e.height = termbox.Size()
	//e.width -= 7
	e.height -= 1
	// Clear the screen and set the padding for the lines
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	maxLineLength := e.width - 2
	lineCountDigits := len(strconv.Itoa(len(e.buffer)))
	lineCountWidth := lineCountDigits + 1 // +1 for the space between line count and '>'

	for i, line := range e.buffer {
		if i < e.height {
			var side rune = ' '
			if i < len(e.buffer)-e.offsetY {
				line = e.buffer[i+e.offsetY]
			}
			// Pad the line with spaces to reach the maximum line length
			paddedLine := fmt.Sprintf("%-*s", maxLineLength, line)
			if e.cursorY == i {
				side = '>'
			}
			// Display the line count based on the Y offset
			lineNumber := i + 1 + e.offsetY
			lineCountStr := strconv.Itoa(lineNumber)
			termbox.SetCell(0, i, rune(lineCountStr[0]), termbox.ColorWhite, termbox.ColorDefault)
			for j, r := range lineCountStr[1:] {
				termbox.SetCell(j+1, i, r, termbox.ColorWhite, termbox.ColorDefault)
			}
			termbox.SetCell(lineCountWidth, i, side, termbox.ColorYellow, termbox.ColorDefault)
			termbox.SetCell(lineCountWidth+1, i, ' ', termbox.ColorDefault, termbox.ColorDefault)
			// Change characters based on line length
			for j, _ := range paddedLine {
				if j < len(paddedLine)-e.offsetX {
					word, bracket, point, WordType := getWord(paddedLine, j+(e.offsetX))
					wordColor := SyntaxHighlight(word, j+e.offsetX, paddedLine, bracket, point, WordType)
					termbox.SetCell(j+lineCountWidth+2, i, rune(paddedLine[j+e.offsetX]), wordColor, termbox.ColorDefault)
				}
			}
		}
		if i == e.height {
			break
		}
	}
	for j := 0; j < e.width; j++ {
		termbox.SetCell(j, e.height, rune(e.StatBar(j)), termbox.ColorBlack, termbox.ColorWhite)
	}
	termbox.Flush()
	termbox.SetCursor(e.cursorX+lineCountWidth+2, e.cursorY)
}

// add character to line
func (e *Editor) AppendCharacter(char rune) {
	// Get the current line and cursor position
	lineIndex := e.cursorY + e.offsetY
	cursorPositionX := e.cursorX + e.offsetX
	cursorPositionY := e.cursorY + e.offsetY

	e.buffer[lineIndex] = e.buffer[lineIndex][:cursorPositionX] + string(char) + e.buffer[lineIndex][cursorPositionX:]

	// Update the cursor position
	e.cursorX++
	if e.cursorX > e.width-7 {
		e.offsetX++
		e.cursorX = e.width - 7
	}

	// Record the action in the UndoBuffer
	e.UndoBuffer = append(e.UndoBuffer, Action{
		CursorX:    cursorPositionX,
		CursorXEND: e.cursorX + e.offsetX,
		CursorY:    cursorPositionY,
		CursorYEND: e.cursorY + e.offsetY,
		Text:       string(char),
		remove:     false,
	})
}
func (e *Editor) Redo() {
	// Check if there are actions to redo
	if len(e.RedoBuffer) == 0 {
		return
	}

	// Pop the last action from the RedoBuffer
	action := e.RedoBuffer[len(e.RedoBuffer)-1]
	e.RedoBuffer = e.RedoBuffer[:len(e.RedoBuffer)-1]

	// Perform the redo action
	if action.remove {
		// If the action was a remove action, remove the text from the buffer
		if action.Text != "\n" {
			e.buffer[action.CursorY] = e.buffer[action.CursorY][:action.CursorX-1] + e.buffer[action.CursorY][action.CursorX:]
		} else {
			e.buffer[action.CursorY] = e.buffer[action.CursorY][:action.CursorX-1] + e.buffer[action.CursorY][action.CursorX-1:]
		}

		// If the action was removing a newline, merge the current line with the next line
		if action.Text == "\n" && action.CursorY < len(e.buffer)-1 {
			e.buffer[action.CursorY] += e.buffer[action.CursorY+1]
			e.buffer = append(e.buffer[:action.CursorY+1], e.buffer[action.CursorY+2:]...)
		}
	} else {
		// If the action was an add action, add the text back to the buffer
		e.buffer[action.CursorY] = e.buffer[action.CursorY][:action.CursorX] + action.Text + e.buffer[action.CursorY][action.CursorX:]
		if action.Text == "\n" {
			// Split the line at the cursor position
			before := e.buffer[action.CursorY][:action.CursorX]
			after := e.buffer[action.CursorY][action.CursorX:]
			// Insert a new line with the text after the cursor
			e.buffer = append(e.buffer[:action.CursorY+1], append([]string{after}, e.buffer[action.CursorY+1:]...)...)
			// Replace the current line with the text before the cursor
			e.buffer[action.CursorY] = before
		}
	}

	// Add the redone action to the UndoBuffer
	e.UndoBuffer = append(e.UndoBuffer, action)
}
func (editor *Editor) Enter() {
	var nextText string = ""
	CursorPosX, CursorPosY := editor.cursorX+editor.offsetX, editor.cursorY+editor.offsetY
	if editor.cursorX+editor.offsetX < len(editor.buffer[editor.cursorY+editor.offsetY]) {
		nextText = editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX:]
		editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX+editor.offsetX]
	}
	editor.buffer = append(editor.buffer[:editor.cursorY+1+editor.offsetY], append([]string{""}, editor.buffer[editor.cursorY+1+editor.offsetY:]...)...)
	editor.cursorY++
	editor.buffer[editor.cursorY+editor.offsetY] = nextText
	editor.cursorX = 0
	editor.offsetX = 0
	if editor.cursorY > editor.height-1 {
		editor.offsetY += 1
		editor.cursorY = editor.height - 1
	}
	editor.UndoBuffer = append(editor.UndoBuffer, Action{
		CursorX:    CursorPosX,
		CursorXEND: editor.cursorX + editor.offsetX,
		CursorY:    CursorPosY,
		CursorYEND: editor.cursorY + editor.offsetY,
		Text:       "\n",
		remove:     false,
	})
}

func main() {
	//Initiate IDE
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	editor := NewEditor()
	err = clipboard.Init()
	if err != nil {
		panic(err)
	}
	//Check if a file is specified
	if len(os.Args) > 1 {
		editor.ReadFile(os.Args[1])
		filename = os.Args[1]
	}
	loadConfig()
	editor.Render()
	defer func() {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		termbox.Close()
	}()
	for {
		//go through possible user inputs
		var currentLine int = editor.cursorY
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return
			case termbox.KeyCtrlV:
				text := string(clipboard.Read(clipboard.FmtText))
				CursorPosX, CursorPosY := editor.cursorX+editor.offsetX, editor.cursorY+editor.offsetY
				if text != "" {
					hasNewLine := false
					count := 0
					for _, c := range text {
						if c == '\n' {
							hasNewLine = true
							count++
						}
					}
					if !hasNewLine {
						editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX+editor.offsetX] + text + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX:]
						if editor.cursorX == editor.width-7 {
							editor.offsetX += len(text)
						} else {
							editor.cursorX += len(text)
							if editor.cursorX > editor.width-7 {
								editor.offsetX = editor.cursorX - editor.width - 7
								editor.cursorX = editor.width - 7
							}
						}
					} else {
						NewStr := strings.Split(text, "\n")
						before := editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX+editor.offsetX]
						after := editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX:]
						line := editor.offsetY + editor.cursorY
						for i := 0; i <= count; i++ {
							if i != 0 {
								editor.buffer = append(editor.buffer[:editor.cursorY+editor.offsetY+1], append([]string{NewStr[i]}, editor.buffer[editor.cursorY+1+editor.offsetY:]...)...)
								editor.cursorY++
							} else {
								editor.buffer = append(editor.buffer[:editor.cursorY+editor.offsetY], append([]string{NewStr[i]}, editor.buffer[editor.cursorY+1+editor.offsetY:]...)...)
							}
						}
						editor.buffer[line] = before + editor.buffer[line]
						editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY] + after
						editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY]) - len(after)
						if editor.cursorX > editor.width-7 {
							editor.offsetX = editor.cursorX - editor.width - 7
							editor.cursorX = editor.width - 7
						}
					}
					editor.UndoBuffer = append(editor.UndoBuffer, Action{
						CursorX:    CursorPosX,
						CursorXEND: editor.cursorX + editor.offsetX,
						CursorY:    CursorPosY,
						CursorYEND: editor.cursorY + editor.offsetY,
						Text:       text,
						remove:     false,
					})
				}
			case termbox.KeyCtrlS:
				editor.SaveFile()
			case termbox.KeyCtrlY:
				editor.Redo()
			case termbox.KeyCtrlZ:
				if len(editor.UndoBuffer) == 0 {
					// No actions to undo
					continue
				}

				// Pop the last action from the UndoBuffer
				action := editor.UndoBuffer[len(editor.UndoBuffer)-1]
				editor.UndoBuffer = editor.UndoBuffer[:len(editor.UndoBuffer)-1]
				// Reverse the action
				if !action.remove {
					if action.Text != "\n" {
						if action.CursorY == action.CursorYEND {
							editor.buffer[action.CursorY] = editor.buffer[action.CursorY][:action.CursorX] + editor.buffer[action.CursorY][action.CursorXEND:]
							if len(action.Text) > editor.width-7 {
								editor.offsetX = len(action.Text) - editor.width - 7 + editor.cursorX
							}
							editor.cursorX = action.CursorX
							//editor.cursorY = action.CursorY
						} else {
							editor.buffer[action.CursorY] = editor.buffer[action.CursorY][:action.CursorX] + editor.buffer[action.CursorYEND][action.CursorXEND:]
							editor.buffer = append(editor.buffer[:action.CursorY+1], editor.buffer[action.CursorYEND+1:]...)
							editor.cursorX = action.CursorX
						}
					} else {
						nextText := editor.buffer[action.CursorYEND][action.CursorXEND:]
						copy(editor.buffer[action.CursorYEND:], editor.buffer[action.CursorYEND+1:])
						editor.buffer = editor.buffer[:len(editor.buffer)-1]
						editor.offsetX = 0
						if editor.cursorX > editor.width-7 {
							editor.offsetX = action.CursorX - editor.width - 7
							editor.cursorX = editor.width - 7
						}
						editor.buffer[action.CursorY+editor.offsetY] += nextText
						editor.cursorY--
						editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
					}
				} else {
					if action.Text != "\n" {
						editor.buffer[action.CursorY+editor.offsetY] = editor.buffer[action.CursorY][:action.CursorXEND] + action.Text + editor.buffer[action.CursorY][action.CursorXEND:]
						if len(action.Text) > editor.width-7 {
							editor.offsetX = len(action.Text) - editor.width - 7 + editor.cursorX
						}
						editor.cursorX = action.CursorXEND
						//editor.cursorY = action.CursorY
					} else {
						editor.Enter()
					}
				}
				if action.CursorY != editor.cursorY+editor.offsetY {
					editor.cursorY = action.CursorY
				}
				if action.CursorX > editor.width-7 {
					editor.offsetX = action.CursorX - editor.width - 7
					editor.cursorX = editor.width - 7
				} else if action.CursorX < editor.width-7+editor.offsetX {
					editor.offsetX = action.CursorX - action.CursorX
				}
				if action.CursorY > editor.height-1 {
					editor.offsetY = action.CursorYEND - editor.height - 1
					editor.cursorY = editor.height - 1
				}

				// Move the action to the RedoBuffer
				editor.RedoBuffer = append(editor.RedoBuffer, action)
			case termbox.KeyEnter:
				editor.Enter()
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if editor.cursorX != 0 || editor.offsetX > 0 {
					var BACKtext string = editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX-1 : editor.cursorX+editor.offsetX]
					editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX+editor.offsetX-1] + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX:]
					editor.cursorX--
					if editor.offsetX > 0 && editor.cursorX == 0 {
						editor.offsetX--
						editor.cursorX++
					}
					editor.UndoBuffer = append(editor.UndoBuffer, Action{
						CursorX:    editor.cursorX + 1,
						CursorXEND: editor.cursorX + editor.offsetX,
						CursorY:    editor.cursorY + editor.offsetY,
						CursorYEND: editor.cursorY + editor.offsetY,
						Text:       string(BACKtext),
						remove:     true,
					})
				} else if editor.cursorY > 0 {
					nextText := editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
					copy(editor.buffer[editor.cursorY+editor.offsetY:], editor.buffer[editor.cursorY+1+editor.offsetY:])
					editor.buffer = editor.buffer[:len(editor.buffer)-1]
					editor.cursorY--
					editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
					editor.offsetX = 0
					editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
					if editor.cursorX > editor.width-7 {
						editor.offsetX = editor.cursorX - editor.width - 7
						editor.cursorX = editor.width - 7
					}
					editor.buffer[editor.cursorY+editor.offsetY] += nextText
					editor.UndoBuffer = append(editor.UndoBuffer, Action{
						CursorX:    editor.cursorX + 1,
						CursorXEND: editor.cursorX + editor.offsetX,
						CursorY:    editor.cursorY + editor.offsetY,
						CursorYEND: editor.cursorY + editor.offsetY,
						Text:       "\n",
						remove:     true,
					})
				}
			case termbox.KeyDelete:
				if editor.cursorX != len(editor.buffer[editor.cursorY+editor.offsetY]) {
					text := editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX]
					editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+1:]
					editor.UndoBuffer = append(editor.UndoBuffer, Action{
						CursorX:    editor.cursorX + 1,
						CursorXEND: editor.cursorX + editor.offsetX,
						CursorY:    editor.cursorY + editor.offsetY,
						CursorYEND: editor.cursorY + editor.offsetY,
						Text:       string(text),
						remove:     true,
					})
				}
			case termbox.KeySpace:
				CursorPosX, CursorPosY := editor.cursorX+editor.offsetX, editor.cursorY+editor.offsetY

				editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + " " + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
				editor.cursorX++
				if editor.cursorX > editor.width-7 {
					editor.offsetX++
					editor.cursorX--
				}
				editor.UndoBuffer = append(editor.UndoBuffer, Action{
					CursorX:    CursorPosX,
					CursorXEND: editor.cursorX + editor.offsetX,
					CursorY:    CursorPosY,
					CursorYEND: editor.cursorY + editor.offsetY,
					Text:       " ",
				})
			case termbox.KeyTab:
				CursorPosX, CursorPosY := editor.cursorX+editor.offsetX, editor.cursorY+editor.offsetY

				editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + "    " + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
				editor.cursorX += len("    ")
				if editor.cursorX > editor.width-7 {
					editor.offsetX++
					editor.cursorX--
				}
				editor.UndoBuffer = append(editor.UndoBuffer, Action{
					CursorX:    CursorPosX,
					CursorXEND: editor.cursorX + editor.offsetX,
					CursorY:    CursorPosY,
					CursorYEND: editor.cursorY + editor.offsetY,
					Text:       "    ",
				})
			case termbox.KeyArrowLeft:
				if editor.cursorX > 0 || editor.offsetX > 0 {
					editor.cursorX--
					if editor.cursorX < 0 {
						editor.offsetX--
						editor.cursorX++
					}
				}
			case termbox.KeyArrowRight:
				if editor.cursorX+editor.offsetX < len(editor.buffer[currentLine+editor.offsetY]) {
					editor.cursorX++
					if editor.cursorX > editor.width-7 {
						editor.offsetX++
						editor.cursorX--
					}
				}
			case termbox.KeyArrowUp:
				if currentLine > 0 || editor.offsetY > 0 {
					if editor.offsetY > 0 && currentLine == 0 {
						editor.offsetY--
					} else {
						editor.cursorY--
					}
					if len(editor.buffer[editor.cursorY+editor.offsetY]) < editor.cursorX+editor.offsetX {
						editor.offsetX = 0
						editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetX])
						editor.offsetX = 0
						editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
						if editor.cursorX > editor.width-7 {
							editor.offsetX = editor.cursorX - editor.width - 7
							editor.cursorX = editor.width - 7
						}
					}

				}
			case termbox.KeyArrowDown:
				if editor.offsetY+editor.cursorY < len(editor.buffer)-1 {
					if editor.cursorY == editor.height-1 {
						editor.offsetX = 0
						editor.offsetY++

					} else {
						editor.cursorY++
					}
					if len(editor.buffer[editor.cursorY+editor.offsetY]) < editor.cursorX+editor.offsetX {
						editor.offsetX = 0
						editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
						if editor.cursorX > editor.width-7 {
							editor.offsetX = editor.cursorX - editor.width - 7
							editor.cursorX = editor.width - 7
						}
					}

				}
			default:
				if ev.Ch != 0 && string(ev.Ch) != "" && string(ev.Ch) != " " {
					editor.AppendCharacter(ev.Ch)
				}
			}
		case termbox.EventResize:
			editor.Render()
		case termbox.EventError:
			panic(ev.Err)
		}

		editor.Render()
	}
}
