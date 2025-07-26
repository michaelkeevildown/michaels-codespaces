package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

// componentItem implements list.Item
type componentItem struct {
	component Component
}

func (i componentItem) FilterValue() string {
	return i.component.Name
}

func (i componentItem) Title() string {
	check := "[ ]"
	if i.component.Selected {
		check = "[x]"
	}
	return fmt.Sprintf("%s %s %s", check, i.component.Emoji, i.component.Name)
}

func (i componentItem) Description() string {
	return i.component.Description
}

// itemDelegate implements list.ItemDelegate
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(componentItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", i.Title(), i.Description())

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// Model represents the component selector
type Model struct {
	list     list.Model
	items    []list.Item
	quitting bool
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// Save selections and quit
			for i, item := range m.items {
				if comp, ok := item.(componentItem); ok {
					Registry[i].Selected = comp.component.Selected
				}
			}
			m.quitting = true
			return m, tea.Quit

		case " ":
			// Toggle selection
			i, ok := m.list.SelectedItem().(componentItem)
			if ok {
				i.component.Selected = !i.component.Selected
				m.items[m.list.Index()] = i
				m.list.SetItem(m.list.Index(), i)
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the model
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// NewSelector creates a new component selector model
func NewSelector() Model {
	items := []list.Item{}
	for _, comp := range Registry {
		items = append(items, componentItem{component: comp})
	}

	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "🚀 Select Components to Install"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys(" "),
				key.WithHelp("space", "toggle"),
			),
		}
	}

	return Model{
		list:  l,
		items: items,
	}
}

// SelectComponents runs the interactive component selector
func SelectComponents() ([]Component, error) {
	p := tea.NewProgram(NewSelector(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return nil, fmt.Errorf("failed to run component selector: %w", err)
	}
	return GetSelected(), nil
}

// Simple key binding helper
type key struct{}

func (k key) NewBinding(opts ...func(*keyBinding)) keyBinding {
	kb := keyBinding{}
	for _, opt := range opts {
		opt(&kb)
	}
	return kb
}

func (k key) WithKeys(keys ...string) func(*keyBinding) {
	return func(kb *keyBinding) {
		kb.keys = keys
	}
}

func (k key) WithHelp(key, desc string) func(*keyBinding) {
	return func(kb *keyBinding) {
		kb.help = []string{key, desc}
	}
}

type keyBinding struct {
	keys []string
	help []string
}

func (kb keyBinding) Keys() []string      { return kb.keys }
func (kb keyBinding) Help() (string, string) { return kb.help[0], kb.help[1] }