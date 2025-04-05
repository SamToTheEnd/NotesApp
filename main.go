package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Note struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Todo struct {
	Task      string    `json:"task"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	notes       []Note
	todos       []Todo
	noteFile    = "notes.json"
	todoFile    = "todos.json"
	currentNote = -1
	currentTodo = -1
)

func main() {
	loadData()

	a := app.New()
	w := a.NewWindow("Terminal Notes v1.0")
	w.Resize(fyne.NewSize(800, 600))

	a.Settings().SetTheme(&TerminalTheme{})

	// Create data bindings
	notesList := binding.BindStringList(&[]string{})
	todosList := binding.BindStringList(&[]string{})
	noteInput := widget.NewMultiLineEntry()
	todoInput := widget.NewEntry()

	refreshLists(notesList, todosList)

	// Create lists
	notesWidget := widget.NewListWithData(notesList,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	todosWidget := widget.NewListWithData(todosList,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	// Note management buttons
	noteButtons := container.NewGridWithColumns(4,
		newTerminalButton("Add Note", func() {
			if noteInput.Text != "" {
				notes = append(notes, Note{
					Content:   noteInput.Text,
					CreatedAt: time.Now(),
				})
				noteInput.SetText("")
				refreshLists(notesList, todosList)
				saveData()
			}
		}),
		newTerminalButton("Edit Note", func() {
			if currentNote >= 0 {
				noteInput.SetText(notes[currentNote].Content)
				notes = append(notes[:currentNote], notes[currentNote+1:]...)
				refreshLists(notesList, todosList)
				currentNote = -1
			}
		}),
		newTerminalButton("Delete Note", func() {
			if currentNote >= 0 {
				notes = append(notes[:currentNote], notes[currentNote+1:]...)
				refreshLists(notesList, todosList)
				currentNote = -1
				saveData()
			}
		}),
		newTerminalButton("Clear Notes", func() {
			notes = []Note{}
			refreshLists(notesList, todosList)
			saveData()
		}),
	)

	// Todo management buttons
	todoButtons := container.NewGridWithColumns(4,
		newTerminalButton("Add Todo", func() {
			if todoInput.Text != "" {
				todos = append(todos, Todo{
					Task:      todoInput.Text,
					CreatedAt: time.Now(),
				})
				todoInput.SetText("")
				refreshLists(notesList, todosList)
				saveData()
			}
		}),
		newTerminalButton("Toggle Done", func() {
			if currentTodo >= 0 {
				todos[currentTodo].Done = !todos[currentTodo].Done
				refreshLists(notesList, todosList)
				saveData()
			}
		}),
		newTerminalButton("Delete Todo", func() {
			if currentTodo >= 0 {
				todos = append(todos[:currentTodo], todos[currentTodo+1:]...)
				refreshLists(notesList, todosList)
				currentTodo = -1
				saveData()
			}
		}),
		newTerminalButton("Clear Todos", func() {
			todos = []Todo{}
			refreshLists(notesList, todosList)
			saveData()
		}),
	)

	// Selection handlers
	notesWidget.OnSelected = func(id int) {
		currentNote = id
		currentTodo = -1
	}

	todosWidget.OnSelected = func(id int) {
		currentTodo = id
		currentNote = -1
	}

	// Layout
	notesPanel := container.NewBorder(
		widget.NewLabel("Notes (Ctrl+N)"),
		noteButtons,
		nil, nil,
		container.NewVScroll(notesWidget),
	)

	todosPanel := container.NewBorder(
		widget.NewLabel("Todos (Ctrl+T)"),
		todoButtons,
		nil, nil,
		container.NewVScroll(todosWidget),
	)

	inputPanel := container.NewVBox(
		widget.NewLabel("Input (Ctrl+S to save)"),
		noteInput,
		todoInput,
	)

	mainSplit := container.NewHSplit(
		container.NewVSplit(
			notesPanel,
			todosPanel,
		),
		inputPanel,
	)

	// Keyboard shortcuts
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		switch e.Name {
		case fyne.KeyN:
			w.Canvas().Focus(noteInput)
		case fyne.KeyT:
			w.Canvas().Focus(todoInput)
		case fyne.KeyS:
			saveData()
			dialog.ShowInformation("Saved", "Data successfully saved!", w)
		}
	})

	w.SetContent(mainSplit)
	w.ShowAndRun()
}

func newTerminalButton(text string, action func()) *widget.Button {
	return widget.NewButton(text, action)
}

func refreshLists(notesList binding.StringList, todosList binding.StringList) {
	noteStrs := make([]string, len(notes))
	for i, note := range notes {
		noteStrs[i] = fmt.Sprintf("[%s] %s",
			note.CreatedAt.Format("2006-01-02 15:04"),
			note.Content)
	}

	todoStrs := make([]string, len(todos))
	for i, todo := range todos {
		status := "[ ]"
		if todo.Done {
			status = "[x]"
		}
		todoStrs[i] = fmt.Sprintf("%s [%s] %s",
			status,
			todo.CreatedAt.Format("2006-01-02 15:04"),
			todo.Task)
	}

	notesList.Set(noteStrs)
	todosList.Set(todoStrs)
}

func loadData() {
	loadFromFile(noteFile, &notes)
	loadFromFile(todoFile, &todos)
}

func loadFromFile(filename string, data interface{}) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	json.Unmarshal(file, data)
}

func saveData() {
	saveToFile(noteFile, notes)
	saveToFile(todoFile, todos)
}

func saveToFile(filename string, data interface{}) {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling data: %v\n", err)
		return
	}

	err = os.WriteFile(filename, file, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	}
}

type TerminalTheme struct{}

func (t TerminalTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return &color.RGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 0xff}
	case theme.ColorNameForeground:
		return color.RGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff}
	case theme.ColorNameButton:
		return color.RGBA{R: 0x2e, G: 0x2e, B: 0x2e, A: 0xff}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t TerminalTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t TerminalTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t TerminalTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 12
	}
	return theme.DefaultTheme().Size(name)
}
