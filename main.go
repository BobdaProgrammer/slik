//A terminal based text editor




package main

import (
    "golang.design/x/clipboard"
    "github.com/nsf/termbox-go"
    "io/ioutil"
    "fmt"
    "strings"
    "os"
)

var filename string = ""

// defining the structure of the text editor
type Action struct {
    CursorX int
    CursorXEND int
    CursorY int
    CursorYEND int
    Text    string
}

type Editor struct {
    buffer       []string
    UndoBuffer   []Action
    RedoBuffer   []Action
    cursorX      int
    cursorY      int
    offsetX      int
    offsetY      int
    width        int
    height       int
}

// creating the editor
func NewEditor() *Editor {
    var width, height int = termbox.Size()
    width -= 3
    height -= 1
    return &Editor{
        buffer:       []string{""},
        UndoBuffer:   []Action{},
        RedoBuffer:   []Action{},
        cursorX:      0,
        cursorY:      0,
        offsetX:      0,
        offsetY:      0,
        width:        width,
        height:       height,
    }
}

//Saving the text to the file
func (e *Editor) SaveFile(){
    if filename!=""{
        //if the file already exists just truncate
        file, err := os.Create(filename)
        if err != nil {
            panic(err)
        }
        defer file.Close()
        //join all the lines into one string and writing to the file
        text := strings.Join(e.buffer, "\n")
        _, err = file.WriteString(text)
        if err != nil{
            panic(err)
        }

    } else{
        //if no file is specified then create untitled.txt
        file, err := os.Create("untitled.txt")
        if err != nil {
            panic(err)
        }
        defer file.Close()
        text := strings.Join(e.buffer, "\n")
        _, err = file.WriteString(text)
        if err != nil{
            panic(err)
        }
    }
}

//adding text from file to screen
func (e *Editor) writeEditor(data string){
    data = strings.Replace(data, "\t", "    ", -1)
    //turn data from string into an array and display to screen
    items := strings.Split(data, "\n")
    e.buffer = items
}

//reading file from command line arguments
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

//rendering the text on the screen
func (e *Editor) Render() {
    //clear the screen and set the padding for the lines
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    maxLineLength := e.width - 2
    for i, line := range e.buffer {
        var side rune = ' '
        if i < len(e.buffer) - e.offsetY{
            line = e.buffer[i+e.offsetY]
        }
        // Pad the line with spaces to reach the maximum line length
        paddedLine := fmt.Sprintf("%-*s", maxLineLength, line)
        if e.cursorY == i{
            side = '>'
        }
        termbox.SetCell(0, i, side, termbox.ColorYellow, termbox.ColorDefault)
        termbox.SetCell(1, i, ' ', termbox.ColorDefault, termbox.ColorDefault)
        //change characters based on line length
        for j, _ := range paddedLine {
            if j < len(paddedLine)-e.offsetX{
            termbox.SetCell(j+2, i, rune(paddedLine[j+e.offsetX]), termbox.ColorDefault, termbox.ColorDefault)
            }
        }
    }

    termbox.SetCursor(e.cursorX+2, e.cursorY)
    termbox.Flush()
}
//add character to line
func (e *Editor) AppendCharacter(char rune) {
    // Get the current line and cursor position
    lineIndex := e.cursorY + e.offsetY
    cursorPosition := e.cursorX + e.offsetX

        e.buffer[lineIndex] = e.buffer[lineIndex][:cursorPosition] + string(char) + e.buffer[lineIndex][cursorPosition:]

    // Update the cursor position
    e.cursorX++
    if e.cursorX > e.width {
        e.offsetX++
        e.cursorX = e.width
    }

    // Record the action in the UndoBuffer
    e.UndoBuffer = append(e.UndoBuffer, Action{
        CursorX: cursorPosition,
        CursorXEND: e.cursorX+e.offsetX,
        CursorY: e.cursorY,
        CursorYEND: e.cursorY+e.offsetY,
        Text:    string(char),
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
    if err != nil{
        panic(err)
    }
    //Check if a file is specified
    if len(os.Args) >1{
        editor.ReadFile(os.Args[1])
        editor.Render()
        filename = os.Args[1]
    }
    defer func() {
        termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
        termbox.Close()
        fmt.Println(editor.UndoBuffer)
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
                CursorPosX,CursorPosY  := editor.cursorX+editor.offsetX, editor.cursorY+editor.offsetY
                if text != ""{
                editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX+editor.offsetX] + text + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX:]
                if editor.cursorX == editor.width{
                    editor.offsetX+= len(text)
                } else{
                    editor.cursorX+= len(text)
                    if editor.cursorX > editor.width{
                        editor.offsetX = editor.cursorX-editor.width
                        editor.cursorX = editor.width
                    }
                }
                editor.UndoBuffer = append(editor.UndoBuffer, Action{
                    CursorX: CursorPosX,
                    CursorXEND: editor.cursorX+editor.offsetX,
                    CursorY: CursorPosY,
                    CursorYEND: editor.cursorY+editor.offsetY,
                    Text:    text,
                })
            }
            case termbox.KeyCtrlS:
                editor.SaveFile()
            case termbox.KeyCtrlZ:
                if len(editor.UndoBuffer) == 0 {
                    // No actions to undo
                    continue
                }
            
                // Pop the last action from the UndoBuffer
                action := editor.UndoBuffer[len(editor.UndoBuffer)-1]
                editor.UndoBuffer = editor.UndoBuffer[:len(editor.UndoBuffer)-1]
            
                // Reverse the action
                editor.buffer[action.CursorY+editor.offsetY] = editor.buffer[action.CursorY][:action.CursorX]+editor.buffer[action.CursorY][action.CursorXEND:]
                if len(action.Text)>editor.width{
                    editor.offsetX = len(action.Text)-editor.width+editor.cursorX
                }
                editor.cursorX = action.CursorX
                editor.cursorY = action.CursorY
            
                // Move the action to the RedoBuffer
                editor.RedoBuffer = append(editor.RedoBuffer, action)
            case termbox.KeyEnter:
                //editor.UndoBuffer = append(editor.UndoBuffer, []int{editor.cursorX + editor.offsetX, editor.cursorY + editor.offsetY, []string{"\n"}})
                var nextText string = ""
                if editor.cursorX+editor.offsetX < len(editor.buffer[editor.cursorY+editor.offsetY]) {
                    nextText = editor.buffer[editor.cursorY][editor.cursorX:]
                    editor.buffer[editor.cursorY] = editor.buffer[editor.cursorY][:editor.cursorX]
                }
                editor.buffer = append(editor.buffer[:editor.cursorY+1+editor.offsetY], append([]string{""}, editor.buffer[editor.cursorY+1+editor.offsetY:]...)...)
                editor.cursorY++
                editor.buffer[editor.cursorY+editor.offsetY] = nextText
                editor.cursorX = 0
                editor.offsetX = 0
            case termbox.KeyBackspace, termbox.KeyBackspace2:
                if editor.cursorX != 0 || editor.offsetX>0{
                    editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX-1] + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
                    editor.cursorX--
                    if editor.offsetX>0&&editor.cursorX==0{
                        editor.offsetX--
                        editor.cursorX++
                    }
                } else if editor.cursorY > 0 {
                    nextText := editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
                    copy(editor.buffer[editor.cursorY+editor.offsetY:], editor.buffer[editor.cursorY+1+editor.offsetY:])
                    editor.buffer = editor.buffer[:len(editor.buffer)-1]
                    editor.cursorY--
                    editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
                    editor.offsetX = 0
                    editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
                    if editor.cursorX > editor.width{
                        editor.offsetX = editor.cursorX-editor.width
                        editor.cursorX = editor.width
                    }
                    editor.buffer[editor.cursorY+editor.offsetY] += nextText 
                }
            case termbox.KeyDelete:
                if editor.cursorX != len(editor.buffer[editor.cursorY+editor.offsetY]) {
                    editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+1:]
                }
            case termbox.KeySpace:
                editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + " " + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
                editor.cursorX++
                if editor.cursorX > editor.width{
                    editor.offsetX++
                    editor.cursorX--
                }
            case termbox.KeyTab:
                editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + "    " + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
                editor.cursorX += len("    ")
                if editor.cursorX > editor.width{
                    editor.offsetX++
                    editor.cursorX--
                }
            case termbox.KeyArrowLeft:
                if editor.cursorX > 0||editor.offsetX>0 {
                    editor.cursorX--
                    if editor.cursorX < 0{
                        editor.offsetX--
                        editor.cursorX++
                    }
                }
            case termbox.KeyArrowRight:
                if editor.cursorX+editor.offsetX < len(editor.buffer[currentLine+editor.offsetY]) {
                    editor.cursorX++
                    if editor.cursorX > editor.width{
                        editor.offsetX++
                        editor.cursorX--
                    }
                }
            case termbox.KeyArrowUp:
                if currentLine > 0 || editor.offsetY>0{
                    if editor.offsetY > 0&&currentLine==0{
                        editor.offsetY--
                    } else{
                        editor.cursorY--
                        if len(editor.buffer[editor.cursorY+editor.offsetY]) < editor.cursorX+editor.offsetX {
                            editor.offsetX = 0
                            editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetX])
                            editor.offsetX = 0
                            editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
                            if editor.cursorX > editor.width{
                                editor.offsetX = editor.cursorX-editor.width
                                editor.cursorX = editor.width
                            }
                        }
                    }
                    
                }
            case termbox.KeyArrowDown:
                if editor.offsetY+editor.cursorY < len(editor.buffer)-1 {
                    if editor.cursorY == editor.height{
                            editor.offsetX = 0
                            editor.offsetY++

                    }else{
                    editor.cursorY++
                    if len(editor.buffer[editor.cursorY+editor.offsetY]) < editor.cursorX+editor.offsetX {
                        editor.offsetX = 0
                        editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
                        if editor.cursorX > editor.width{
                            editor.offsetX = editor.cursorX-editor.width
                            editor.cursorX = editor.width
                        }
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