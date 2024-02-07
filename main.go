//A terminal based text editor


package main

import (
    "golang.design/x/clipboard"
    "github.com/nsf/termbox-go"
    "io/ioutil"
    "fmt"
    "strings"
    "strconv"
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
    remove  bool
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
    width -= 7
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
    // Clear the screen and set the padding for the lines
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    maxLineLength := e.width - 2
    lineCountDigits := len(strconv.Itoa(len(e.buffer)))
    lineCountWidth := lineCountDigits + 1 // +1 for the space between line count and '>'

    for i, line := range e.buffer {
        var side rune = ' '
        if i < len(e.buffer) - e.offsetY {
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
                termbox.SetCell(j+lineCountWidth+2, i, rune(paddedLine[j+e.offsetX]), termbox.ColorDefault, termbox.ColorDefault)
            }
        }
    }

    termbox.SetCursor(e.cursorX+lineCountWidth+2, e.cursorY)
    termbox.Flush()
}
//add character to line
func (e *Editor) AppendCharacter(char rune) {
    // Get the current line and cursor position
    lineIndex := e.cursorY + e.offsetY
    cursorPositionX := e.cursorX + e.offsetX
    cursorPositionY := e.cursorY + e.offsetY

        e.buffer[lineIndex] = e.buffer[lineIndex][:cursorPositionX] + string(char) + e.buffer[lineIndex][cursorPositionX:]

    // Update the cursor position
    e.cursorX++
    if e.cursorX > e.width {
        e.offsetX++
        e.cursorX = e.width
    }

    // Record the action in the UndoBuffer
    e.UndoBuffer = append(e.UndoBuffer, Action{
        CursorX: cursorPositionX,
        CursorXEND: e.cursorX+e.offsetX,
        CursorY: cursorPositionY,
        CursorYEND: e.cursorY+e.offsetY,
        Text:    string(char),
        remove:  false,
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
        if action.Text != "\n"{
        e.buffer[action.CursorY] = e.buffer[action.CursorY][:action.CursorX-1] + e.buffer[action.CursorY][action.CursorX:]
        } else{
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
func (editor *Editor)Enter(){
    var nextText string = ""
    CursorPosX, CursorPosY := editor.cursorX+editor.offsetX, editor.cursorY+editor.offsetY
    if editor.cursorX+editor.offsetX < len(editor.buffer[editor.cursorY+editor.offsetY]) {
        nextText = editor.buffer[editor.cursorY][editor.cursorX:]
        editor.buffer[editor.cursorY] = editor.buffer[editor.cursorY][:editor.cursorX]
    }
    editor.buffer = append(editor.buffer[:editor.cursorY+1+editor.offsetY], append([]string{""}, editor.buffer[editor.cursorY+1+editor.offsetY:]...)...)
    editor.cursorY++
    editor.buffer[editor.cursorY+editor.offsetY] = nextText
    editor.cursorX = 0
    editor.offsetX = 0
    if editor.cursorY>editor.height{
        editor.offsetY += 1
        editor.cursorY = editor.height 
    }
    editor.UndoBuffer = append(editor.UndoBuffer, Action{
        CursorX: CursorPosX,
        CursorXEND: editor.cursorX+editor.offsetX,
        CursorY: CursorPosY,
        CursorYEND: editor.cursorY+editor.offsetY,
        Text:    "\n",
        remove: false,
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
        action := editor.UndoBuffer[len(editor.UndoBuffer)-1]
        fmt.Println(editor.cursorY, action.CursorY, editor.height)
        fmt.Println(action.CursorY, editor.cursorY+editor.offsetY)
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
                    remove: false,
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
                if action.CursorXEND > editor.width{
                    editor.offsetX = action.CursorXEND-editor.width
                    editor.cursorX = editor.width
                } 
                if action.CursorYEND > editor.height{
                    editor.offsetY = action.CursorYEND-editor.height
                    editor.cursorY = editor.height
                } else if action.CursorYEND-editor.offsetY<0{
                    editor.offsetY = action.CursorYEND
                    editor.cursorY = 0
                }
                // Reverse the action
                if !action.remove{
                    if action.Text != "\n"{
                        editor.buffer[action.CursorY] = editor.buffer[action.CursorY][:action.CursorX]+editor.buffer[action.CursorY][action.CursorXEND:]
                        if len(action.Text)>editor.width{
                          editor.offsetX = len(action.Text)-editor.width+editor.cursorX
                        }
                        editor.cursorX = action.CursorX
                        //editor.cursorY = action.CursorY
                    } else{
                            nextText := editor.buffer[action.CursorYEND][action.CursorXEND:]
                            copy(editor.buffer[action.CursorYEND:], editor.buffer[action.CursorYEND+1:])
                            editor.buffer = editor.buffer[:len(editor.buffer)-1]
                            editor.offsetX = 0
                            if editor.cursorX > editor.width{
                                editor.offsetX = action.CursorX-editor.width
                                editor.cursorX = editor.width
                            }
                            editor.buffer[action.CursorY+editor.offsetY] += nextText 
                            editor.cursorY--
                            editor.cursorX = len(editor.buffer[editor.cursorY+editor.offsetY])
            }
            } else{
                if action.Text!="\n"{
                editor.buffer[action.CursorY+editor.offsetY] = editor.buffer[action.CursorY][:action.CursorXEND]+action.Text+editor.buffer[action.CursorY][action.CursorXEND:]
                if len(action.Text)>editor.width{
                    editor.offsetX = len(action.Text)-editor.width+editor.cursorX
                }
                editor.cursorX = action.CursorXEND
                //editor.cursorY = action.CursorY
            } else{
                editor.Enter()
            }
            }
            if action.CursorY != editor.cursorY+editor.offsetY{
                editor.cursorY = action.CursorY
            }
                if action.CursorX > editor.width{
                    editor.offsetX = action.CursorX-editor.width
                    editor.cursorX = editor.width
                } else if action.CursorX<editor.width+editor.offsetX{
                    editor.offsetX = action.CursorX-action.CursorX
                }
                if action.CursorY > editor.height{
                    editor.offsetY = action.CursorYEND-editor.height
                    editor.cursorY = 0
                }
            
                // Move the action to the RedoBuffer
                editor.RedoBuffer = append(editor.RedoBuffer, action)
            case termbox.KeyEnter:
                editor.Enter()
            case termbox.KeyBackspace, termbox.KeyBackspace2:
                if editor.cursorX != 0 || editor.offsetX>0{
                    var BACKtext string = editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX-1:editor.cursorX+editor.offsetX]
                    editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX+editor.offsetX-1] + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX:]
                    editor.cursorX--
                    if editor.offsetX>0&&editor.cursorX==0{
                        editor.offsetX--
                        editor.cursorX++
                    }
                    editor.UndoBuffer = append(editor.UndoBuffer, Action{
                        CursorX: editor.cursorX+1,
                        CursorXEND: editor.cursorX+editor.offsetX,
                        CursorY: editor.cursorY+editor.offsetY,
                        CursorYEND: editor.cursorY+editor.offsetY,
                        Text:    string(BACKtext),
                        remove:  true,
                    })
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
                    editor.UndoBuffer = append(editor.UndoBuffer, Action{
                        CursorX: editor.cursorX+1,
                        CursorXEND: editor.cursorX+editor.offsetX,
                        CursorY: editor.cursorY+editor.offsetY,
                        CursorYEND: editor.cursorY+editor.offsetY,
                        Text:    "\n",
                        remove:  true,
                    })
                }
            case termbox.KeyDelete:
                if editor.cursorX != len(editor.buffer[editor.cursorY+editor.offsetY]) {
                    text := editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+editor.offsetX]
                    editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX+1:]
                    editor.UndoBuffer = append(editor.UndoBuffer, Action{
                        CursorX: editor.cursorX+1,
                        CursorXEND: editor.cursorX+editor.offsetX,
                        CursorY: editor.cursorY+editor.offsetY,
                        CursorYEND: editor.cursorY+editor.offsetY,
                        Text:    string(text),
                        remove:  true,
                    })
                }
            case termbox.KeySpace:
                CursorPosX,CursorPosY  := editor.cursorX+editor.offsetX, editor.cursorY+editor.offsetY

                editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + " " + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
                editor.cursorX++
                if editor.cursorX > editor.width{
                    editor.offsetX++
                    editor.cursorX--
                }
                editor.UndoBuffer = append(editor.UndoBuffer, Action{
                    CursorX: CursorPosX,
                    CursorXEND: editor.cursorX+editor.offsetX,
                    CursorY: CursorPosY,
                    CursorYEND: editor.cursorY+editor.offsetY,
                    Text:    " ",
                })
            case termbox.KeyTab:
                CursorPosX,CursorPosY  := editor.cursorX+editor.offsetX, editor.cursorY+editor.offsetY

                editor.buffer[editor.cursorY+editor.offsetY] = editor.buffer[editor.cursorY+editor.offsetY][:editor.cursorX] + "    " + editor.buffer[editor.cursorY+editor.offsetY][editor.cursorX:]
                editor.cursorX += len("    ")
                if editor.cursorX > editor.width{
                    editor.offsetX++
                    editor.cursorX--
                }
                editor.UndoBuffer = append(editor.UndoBuffer, Action{
                    CursorX: CursorPosX,
                    CursorXEND: editor.cursorX+editor.offsetX,
                    CursorY: CursorPosY,
                    CursorYEND: editor.cursorY+editor.offsetY,
                    Text:    "    ",
                })
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