package cmd

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	proxyClient "github.com/up-zero/my-proxy/client/proxy"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/proxy"
	"strings"
)

type model struct {
	inputs    []textinput.Model
	focused   int
	err       string
	selector  selector
	submitted bool
	pb        *models.ProxyBasic
}

type selector struct {
	options []string
	cursor  int
}

func initialModel(pb *models.ProxyBasic) model {
	m := model{
		inputs:  make([]textinput.Model, 4),
		focused: 0,
		selector: selector{
			options: []string{"TCP", "UDP"},
			cursor:  0,
		},
		pb: pb,
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("201"))
		t.CharLimit = 32
		t.Width = 100
		t.Prompt = ""
		switch i {
		case 0:
			t.Placeholder = "proxy name"
			t.SetValue(pb.Name)
		case 1:
			t.Placeholder = "eg: 8080"
			t.SetValue(pb.ListenPort)
		case 2:
			t.Placeholder = "eg: 192.168.1.8"
			t.SetValue(pb.TargetAddress)
		case 3:
			t.Placeholder = "eg: 8093"
			t.SetValue(pb.TargetPort)
		}
		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.submitted {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs) {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs)
			}

			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focused-1 {
					m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("201"))
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = lipgloss.NewStyle()
				}
			}

			return m, nil

		case "enter":
			if m.focused == 0 { // 如果在选择器上
				// 切换焦点到第一个输入框
				m.focused = 1
				m.inputs[0].Focus()
				m.inputs[0].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("201"))
				return m, nil
			}

			// 校验所有字段
			m.err = ""
			for _, input := range m.inputs {
				if input.Value() == "" {
					m.err = "All fields cannot be empty"
					return m, nil
				}
			}

			// 保存
			if m.pb.Uuid == "" {
				// 新增
				req := &proxy.CreateRequest{
					Name:          m.inputs[0].Value(),
					ListenPort:    m.inputs[1].Value(),
					TargetAddress: m.inputs[2].Value(),
					TargetPort:    m.inputs[3].Value(),
					Type:          m.selector.options[m.selector.cursor],
				}
				if err := proxyClient.Create(req); err != nil {
					m.err = err.Error()
					return m, nil
				}
			} else {
				// 修改
				req := &proxy.EditRequest{
					Uuid:          m.pb.Uuid,
					Name:          m.inputs[0].Value(),
					ListenPort:    m.inputs[1].Value(),
					TargetAddress: m.inputs[2].Value(),
					TargetPort:    m.inputs[3].Value(),
					Type:          m.selector.options[m.selector.cursor],
				}
				if err := proxyClient.Edit(req); err != nil {
					m.err = err.Error()
					return m, nil
				}
			}

			m.submitted = true
			return m, tea.Quit

		case "left", "right":
			if m.focused == 0 { // 仅在选择器聚焦时响应
				if msg.String() == "left" {
					m.selector.cursor--
					if m.selector.cursor < 0 {
						m.selector.cursor = len(m.selector.options) - 1
					}
				} else {
					m.selector.cursor++
					if m.selector.cursor >= len(m.selector.options) {
						m.selector.cursor = 0
					}
				}
			}
		}
	}

	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.submitted {
		var b strings.Builder
		b.WriteString("Config Save Success!\n")
		b.WriteString(strings.Repeat("-----", 10) + "\n")
		b.WriteString(fmt.Sprintf("Type:           %s\n", m.selector.options[m.selector.cursor]))
		b.WriteString(fmt.Sprintf("Name:           %s\n", m.inputs[0].Value()))
		b.WriteString(fmt.Sprintf("Listen Port:    %s\n", m.inputs[1].Value()))
		b.WriteString(fmt.Sprintf("Target Address: %s\n", m.inputs[2].Value()))
		b.WriteString(fmt.Sprintf("Target Port:    %s\n", m.inputs[3].Value()))
		return b.String()
	}

	var b strings.Builder

	b.WriteString("Proxy Config \n\n")

	// 渲染选择器
	selectorStyle := lipgloss.NewStyle().Padding(0, 1)
	var selectorOptions []string
	for i, option := range m.selector.options {
		if i == m.selector.cursor {
			selectorOptions = append(selectorOptions, lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true).Render("["+option+"]"))
		} else {
			selectorOptions = append(selectorOptions, option)
		}
	}
	selectLabel := "Type:            "
	if m.focused == 0 {
		selectLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Render(selectLabel)
	}
	b.WriteString(selectorStyle.Render(selectLabel+strings.Join(selectorOptions, "  ")) + "\n")

	// 渲染输入框
	labels := []string{"Name:          ", "Listen Port:   ", "Target Address:", "Target Port:   "}
	for i, input := range m.inputs {
		style := lipgloss.NewStyle().Padding(0, 1)
		if m.focused == i+1 {
			style = style.Foreground(lipgloss.Color("201"))
		}
		b.WriteString(style.Render(labels[i]) + " " + input.View() + "\n")
	}

	b.WriteString("\n")

	// 渲染错误信息
	if m.err != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString(errorStyle.Render(m.err) + "\n\n")
	}

	b.WriteString("Use 'tab'/'shift+tab' or '↑'/'↓' to switch fields, '←'/'→' to select types, 'enter' to submit, 'q' or 'ctrl+c' to exit.\n")

	return b.String()
}
