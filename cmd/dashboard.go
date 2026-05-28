package cmd

import (
	"fmt"
	"math"
	"strings"
	"time"

	dashboardClient "github.com/up-zero/my-proxy/client/dashboard"
	"github.com/up-zero/my-proxy/models"
	serviceDashboard "github.com/up-zero/my-proxy/service/dashboard"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var dashboardInterval time.Duration

var (
	dashboardFrameStyle = lipgloss.NewStyle().Padding(1, 2)
	dashboardTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	dashboardHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	dashboardErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dashboardOkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	dashboardWarnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	dashboardInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	dashboardMutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	panelTitleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	panelSubtitleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	metricValueStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
)

type dashboardModel struct {
	width    int
	height   int
	interval time.Duration
	loading  bool
	err      string
	overview *serviceDashboard.OverviewResponse
}

type overviewLoadedMsg struct {
	overview *serviceDashboard.OverviewResponse
	err      error
}

type dashboardTickMsg time.Time

type summaryCard struct {
	title string
	value string
	sub   string
	color lipgloss.Color
}

func newDashboardModel(interval time.Duration) dashboardModel {
	return dashboardModel{
		width:    120,
		height:   36,
		interval: interval,
		loading:  true,
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return tea.Batch(fetchOverviewCmd(), dashboardTickCmd(m.interval))
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			m.err = ""
			return m, fetchOverviewCmd()
		}
	case overviewLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.err = ""
		m.overview = msg.overview
		return m, nil
	case dashboardTickMsg:
		m.loading = true
		return m, tea.Batch(fetchOverviewCmd(), dashboardTickCmd(m.interval))
	}
	return m, nil
}

func (m dashboardModel) View() string {
	if m.width < 72 || m.height < 20 {
		return dashboardFrameStyle.Render("终端窗口过小，请放大后查看资源监控面板。\n建议至少使用 72x20 的终端大小。")
	}

	contentWidth := maxInt(40, m.width-4)
	header := m.renderHeader(contentWidth)

	if m.overview == nil {
		message := "正在拉取仪表盘数据..."
		if m.err != "" {
			message = dashboardErrorStyle.Render("拉取失败：" + m.err)
		}
		body := renderPanel(contentWidth, "实时总览", "终端资源监控 UI", lipgloss.Color("63"), message)
		return dashboardFrameStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", body))
	}

	summary := renderSummaryCards(m.overview, contentWidth)
	main := m.renderMainLayout(contentWidth)

	return dashboardFrameStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", summary, "", main))
}

func (m dashboardModel) renderHeader(width int) string {
	statusText := dashboardHintStyle.Render("按 r 立即刷新 · 按 q 退出")
	if m.loading {
		statusText = dashboardWarnStyle.Render("采样中...") + "  " + statusText
	} else if m.err != "" {
		statusText = dashboardErrorStyle.Render("最近刷新失败："+m.err) + "  " + statusText
	} else if m.overview != nil {
		statusText = dashboardOkStyle.Render(updatedAtText(m.overview.Summary.UpdatedAt)) + "  " + statusText
	}

	left := dashboardTitleStyle.Render("my-proxy 终端仪表盘")
	right := lipgloss.NewStyle().Align(lipgloss.Right).Width(maxInt(0, width-lipgloss.Width(left)-2)).Render(statusText)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m dashboardModel) renderMainLayout(width int) string {
	trendPanel := renderTrendPanel(m.overview, width)

	if width >= 132 {
		leftWidth := maxInt(68, int(math.Round(float64(width)*0.57)))
		rightWidth := maxInt(36, width-leftWidth-1)
		left := renderTrendPanel(m.overview, leftWidth)
		right := renderNodePanel(m.overview, rightWidth, m.height)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		trendPanel,
		"",
		renderNodePanel(m.overview, width, m.height),
	)
}

func fetchOverviewCmd() tea.Cmd {
	return func() tea.Msg {
		overview, err := dashboardClient.Overview()
		return overviewLoadedMsg{overview: overview, err: err}
	}
}

func dashboardTickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return dashboardTickMsg(t)
	})
}

func renderSummaryCards(overview *serviceDashboard.OverviewResponse, width int) string {
	bodyWidth := panelContentWidth(width)
	cards := []summaryCard{
		{
			title: "代理总数",
			value: formatNumber(int64(overview.Summary.ProxyTotal)),
			sub:   fmt.Sprintf("运行 %s / 停止 %s", formatNumber(int64(overview.Summary.ProxyRunning)), formatNumber(int64(overview.Summary.ProxyStopped))),
			color: lipgloss.Color("39"),
		},
		{
			title: "当前连接数",
			value: formatNumber(overview.Summary.TotalConnections),
			sub:   "活跃 TCP / UDP / HTTP 请求",
			color: lipgloss.Color("214"),
		},
		{
			title: "累计入站",
			value: formatBytesInt64(overview.Summary.TotalTrafficIn),
			sub:   "实时 " + formatRate(overview.Summary.InboundRate),
			color: lipgloss.Color("45"),
		},
		{
			title: "累计出站",
			value: formatBytesInt64(overview.Summary.TotalTrafficOut),
			sub:   "实时 " + formatRate(overview.Summary.OutboundRate),
			color: lipgloss.Color("42"),
		},
		{
			title: "CPU / 内存",
			value: fmt.Sprintf("%s / %s", formatPercent(overview.System.CPUPercent), formatPercent(overview.System.MemoryPercent)),
			sub:   fmt.Sprintf("Go 堆 %s · Goroutines %d", formatBytesUint64(overview.System.GoMemoryAlloc), overview.System.Goroutines),
			color: lipgloss.Color("99"),
		},
		{
			title: "服务运行时长",
			value: formatUptime(overview.Summary.UptimeSeconds),
			sub:   updatedAtText(overview.Summary.UpdatedAt),
			color: lipgloss.Color("135"),
		},
	}

	columns := 1
	switch {
	case bodyWidth >= 156:
		columns = 4
	case bodyWidth >= 112:
		columns = 3
	case bodyWidth >= 78:
		columns = 2
	}
	gap := 1
	cardWidth := maxInt(18, (bodyWidth-(columns-1)*gap)/columns)
	rows := make([]string, 0, (len(cards)+columns-1)/columns)
	for start := 0; start < len(cards); start += columns {
		end := minInt(len(cards), start+columns)
		items := make([]string, 0, end-start)
		for _, card := range cards[start:end] {
			items = append(items, renderSummaryCard(card, cardWidth))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, items...))
	}
	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return renderPanel(width, "摘要指标", "", lipgloss.Color("63"), body)
}

func renderSummaryCard(card summaryCard, width int) string {
	title := lipgloss.NewStyle().Foreground(card.color).Bold(true).Render(card.title)
	value := metricValueStyle.Render(card.value)
	sub := dashboardHintStyle.Render(card.sub)
	body := strings.Join([]string{title, value, sub}, "\n")
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(card.color).
		Padding(0, 1).
		Height(3)
	return style.Width(contentWidthForStyle(width, style)).Render(body)
}

func renderTrendPanel(overview *serviceDashboard.OverviewResponse, width int) string {
	bodyWidth := panelContentWidth(width)
	sections := []string{
		renderTrafficPanel(overview, bodyWidth),
		renderConnectionPanel(overview, bodyWidth),
		renderSystemPanel(overview, bodyWidth),
	}
	return renderPanel(width, "趋势概览", "流量、连接与系统资源等", lipgloss.Color("99"), strings.Join(sections, "\n\n"))
}

func renderTrafficPanel(overview *serviceDashboard.OverviewResponse, width int) string {
	contentWidth := subPanelContentWidth(width)
	inValues := make([]float64, 0, len(overview.Charts.Traffic))
	outValues := make([]float64, 0, len(overview.Charts.Traffic))
	for _, item := range overview.Charts.Traffic {
		inValues = append(inValues, item.InboundRate)
		outValues = append(outValues, item.OutboundRate)
	}

	body := []string{
		renderMetricSeriesBlock("入站", inValues, contentWidth, formatRate(overview.Summary.InboundRate), lipgloss.Color("45")),
		renderMetricSeriesBlock("出站", outValues, contentWidth, formatRate(overview.Summary.OutboundRate), lipgloss.Color("42")),
	}
	return renderSubPanel(width, "流量速率", "入站 / 出站实时趋势", lipgloss.Color("45"), strings.Join(body, "\n"))
}

func renderConnectionPanel(overview *serviceDashboard.OverviewResponse, width int) string {
	contentWidth := subPanelContentWidth(width)
	values := make([]float64, 0, len(overview.Charts.Connections))
	for _, item := range overview.Charts.Connections {
		values = append(values, float64(item.Connections))
	}

	body := []string{
		renderMetricSeriesBlock("连接", values, contentWidth, formatNumber(overview.Summary.TotalConnections), lipgloss.Color("214")),
	}
	return renderSubPanel(width, "连接活跃度", "当前连接走势与瞬时活跃度", lipgloss.Color("214"), strings.Join(body, "\n"))
}

func renderSystemPanel(overview *serviceDashboard.OverviewResponse, width int) string {
	contentWidth := subPanelContentWidth(width)
	cpuValues := make([]float64, 0, len(overview.Charts.System))
	memoryValues := make([]float64, 0, len(overview.Charts.System))
	for _, item := range overview.Charts.System {
		cpuValues = append(cpuValues, item.CPUPercent)
		memoryValues = append(memoryValues, item.MemoryPercent)
	}

	body := []string{
		renderMetricSeriesBlock("CPU", cpuValues, contentWidth, formatPercent(overview.System.CPUPercent), lipgloss.Color("39")),
		renderMetricSeriesBlock("内存", memoryValues, contentWidth, fmt.Sprintf("%s (%s / %s)", formatPercent(overview.System.MemoryPercent), formatBytesUint64(overview.System.MemoryUsed), formatBytesUint64(overview.System.MemoryTotal)), lipgloss.Color("99")),
	}
	return renderSubPanel(width, "系统资源", "CPU / 内存使用趋势", lipgloss.Color("99"), strings.Join(body, "\n"))
}

func renderNodePanel(overview *serviceDashboard.OverviewResponse, width, height int) string {
	if len(overview.Nodes) == 0 {
		return renderPanel(width, "活跃节点", "按连接数与实时速率综合排序", lipgloss.Color("81"), dashboardHintStyle.Render("暂无代理节点"))
	}
	bodyWidth := panelContentWidth(width)

	limit := 6
	if bodyWidth < 50 {
		limit = 4
	}
	if height < 30 {
		limit = 4
	}
	if height < 24 {
		limit = 3
	}
	limit = minInt(limit, len(overview.Nodes))

	items := make([]string, 0, limit)
	for index, node := range overview.Nodes[:limit] {
		items = append(items, renderNodeItem(index+1, node, bodyWidth))
		if index < limit-1 {
			items = append(items, renderDivider(bodyWidth))
		}
	}
	body := strings.Join(items, "\n")
	return renderPanel(width, fmt.Sprintf("活跃节点 Top %d", limit), "按连接数与实时速率综合排序", lipgloss.Color("81"), body)
}

func renderNodeItem(rank int, node serviceDashboard.NodeLoadMetric, width int) string {
	badge := dashboardErrorStyle.Render("STOPPED")
	if node.State == models.ProxyStateRunning {
		badge = dashboardOkStyle.Render("RUNNING")
	}

	tagText := "未打标签"
	if len(node.TagList) > 0 {
		names := make([]string, 0, len(node.TagList))
		for _, tag := range node.TagList {
			names = append(names, tag.Name)
		}
		tagText = strings.Join(names, " / ")
	}

	nameWidth := maxInt(10, width-20)
	head := fmt.Sprintf("%d. %s", rank, truncateText(node.Name, nameWidth))
	typeText := dashboardInfoStyle.Render("[" + node.Type + "]")
	endpoint := truncateText(fmt.Sprintf("%s:%s", node.ListenAddress, node.ListenPort), maxInt(12, width/2))
	stats := fmt.Sprintf("连接 %s   ↑ %s   ↓ %s", formatNumber(node.ActiveConnections), formatRate(node.InboundRate), formatRate(node.OutboundRate))
	totals := fmt.Sprintf("累计 ↑ %s   ↓ %s", formatBytesInt64(node.TrafficIn), formatBytesInt64(node.TrafficOut))
	tags := truncateText("标签: "+tagText, maxInt(14, width-4))

	lines := []string{
		lipgloss.JoinHorizontal(lipgloss.Top, dashboardInfoStyle.Render(head), " ", typeText, " ", badge),
		dashboardHintStyle.Render(endpoint + "   " + stats),
		dashboardMutedStyle.Render(tags + "   " + totals),
	}

	if width < 58 {
		lines = []string{
			lipgloss.JoinHorizontal(lipgloss.Top, dashboardInfoStyle.Render(head), " ", typeText, " ", badge),
			dashboardHintStyle.Render(endpoint),
			dashboardHintStyle.Render(stats),
			dashboardMutedStyle.Render(tags),
		}
	}

	return strings.Join(lines, "\n")
}

func renderMetricSeriesBlock(label string, values []float64, width int, current string, color lipgloss.Color) string {
	headerLabel := lipgloss.NewStyle().Foreground(color).Bold(true).Render(label)
	headerValue := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Align(lipgloss.Right).Render(current)
	contentWidth := metricBlockWidth(width)
	labelWidth := maxInt(6, contentWidth/3)
	head := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(headerLabel),
		lipgloss.NewStyle().Align(lipgloss.Right).Width(maxInt(10, contentWidth-labelWidth)).Render(headerValue),
	)
	spark := lipgloss.NewStyle().Foreground(color).Render(renderSparkline(values, contentWidth))
	return strings.Join([]string{head, spark}, "\n")
}

func renderSparkline(values []float64, width int) string {
	if width <= 0 {
		return ""
	}
	if len(values) == 0 {
		return strings.Repeat("·", width)
	}

	sampled := make([]float64, width)
	for i := 0; i < width; i++ {
		idx := i * len(values) / width
		if idx >= len(values) {
			idx = len(values) - 1
		}
		sampled[i] = values[idx]
	}

	minValue, maxValue := sampled[0], sampled[0]
	for _, value := range sampled[1:] {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}

	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	if maxValue-minValue < 0.0001 {
		block := '▄'
		if maxValue <= 0 {
			block = '▁'
		}
		return strings.Repeat(string(block), width)
	}

	var builder strings.Builder
	for _, value := range sampled {
		ratio := (value - minValue) / (maxValue - minValue)
		index := int(math.Round(ratio * float64(len(blocks)-1)))
		if index < 0 {
			index = 0
		}
		if index >= len(blocks) {
			index = len(blocks) - 1
		}
		builder.WriteRune(blocks[index])
	}
	return builder.String()
}

func renderPanel(width int, title, subtitle string, borderColor lipgloss.Color, body string) string {
	header := panelTitleStyle.Render(title)
	if subtitle != "" {
		header += "\n" + panelSubtitleStyle.Render(subtitle)
	}
	content := header + "\n\n" + body
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Align(lipgloss.Left)
	return style.Width(contentWidthForStyle(width, style)).Render(content)
}

func renderSubPanel(width int, title, subtitle string, borderColor lipgloss.Color, body string) string {
	header := panelTitleStyle.Render(title)
	if subtitle != "" {
		header += "\n" + panelSubtitleStyle.Render(subtitle) + "\n"
	}
	content := header + "\n" + body
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Align(lipgloss.Left)
	return style.Width(contentWidthForStyle(width, style)).Render(content)
}

func renderDivider(width int) string {
	if width <= 0 {
		return ""
	}
	return dashboardMutedStyle.Render(strings.Repeat("─", maxInt(8, width-2)))
}

func contentWidthForStyle(totalWidth int, style lipgloss.Style) int {
	return maxInt(0, totalWidth-style.GetHorizontalFrameSize())
}

func panelContentWidth(totalWidth int) int {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	return contentWidthForStyle(totalWidth, style)
}

func subPanelContentWidth(totalWidth int) int {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)
	return contentWidthForStyle(totalWidth, style)
}

func metricBlockWidth(totalWidth int) int {
	// Keep a tiny right-side breathing room for sparkline blocks so they don't
	// touch the panel edge or wrap in terminals with stricter width handling.
	return maxInt(12, totalWidth-2)
}

func formatBytesInt64(value int64) string {
	return formatBytesFloat64(float64(value))
}

func formatBytesUint64(value uint64) string {
	return formatBytesFloat64(float64(value))
}

func formatBytesFloat64(value float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	if value < 0 {
		value = 0
	}
	unitIndex := 0
	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}
	return fmt.Sprintf("%.2f%s", value, units[unitIndex])
}

func formatRate(value float64) string {
	return formatBytesFloat64(value) + "/s"
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}

func formatNumber(value int64) string {
	negative := value < 0
	if negative {
		value = -value
	}
	text := fmt.Sprintf("%d", value)
	if len(text) <= 3 {
		if negative {
			return "-" + text
		}
		return text
	}
	var parts []string
	for len(text) > 3 {
		parts = append([]string{text[len(text)-3:]}, parts...)
		text = text[:len(text)-3]
	}
	parts = append([]string{text}, parts...)
	result := strings.Join(parts, ",")
	if negative {
		return "-" + result
	}
	return result
}

func formatUptime(totalSeconds int64) string {
	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	if days > 0 {
		return fmt.Sprintf("%d天 %d时", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%d时 %d分", hours, minutes)
	}
	return fmt.Sprintf("%d分 %d秒", minutes, seconds)
}

func updatedAtText(timestamp int64) string {
	if timestamp <= 0 {
		return "等待采样"
	}
	return "更新于 " + time.UnixMilli(timestamp).Format("15:04:05")
}

func truncateText(value string, limit int) string {
	runes := []rune(value)
	if limit <= 0 || len(runes) <= limit {
		return value
	}
	if limit <= 1 {
		return string(runes[:limit])
	}
	return string(runes[:limit-1]) + "…"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var dashboardCmd = &cobra.Command{
	Use:     "stats",
	Aliases: []string{"dashboard", "monitor", "top"},
	Short:   "show terminal stats for real-time proxy and system metrics",
	Long:    "",
	Example: "my-proxy stats\nmy-proxy stats --interval 2s\nmy-proxy dashboard # backward-compatible alias",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if dashboardInterval <= 0 {
			return fmt.Errorf("interval must be greater than 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		program := tea.NewProgram(newDashboardModel(dashboardInterval), tea.WithAltScreen())
		if _, err := program.Run(); err != nil {
			fmt.Printf("stats ui error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().DurationVarP(&dashboardInterval, "interval", "i", 3*time.Second, "refresh interval")
}
