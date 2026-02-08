package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/logicossoftware/go-mdocx"
	"github.com/mattn/go-sixel"
	"golang.org/x/image/draw"
)

func runBrowseTUI(doc *mdocx.Document, header *headerInfo, theme string, noImages bool) error {
	supportsImages := supportsSixel() && !noImages
	model, err := newBrowseModel(doc, header, theme, supportsImages)
	if err != nil {
		return err
	}
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}

type browseModel struct {
	doc            *mdocx.Document
	header         *headerInfo
	tabs           []string
	activeTab      int
	markdownList   list.Model
	mediaList      list.Model
	viewport       viewport.Model
	metadataView   string
	headerView     string
	supportsImages bool
	theme          string
	width          int
	height         int
	listWidth      int
	contentWidth   int
	activeStyle    lipgloss.Style
	inactiveStyle  lipgloss.Style
	renderer       *glamour.TermRenderer
	viewingImage   bool
	imageContent   string
}

type markdownItem struct {
	path string
	size int
}

func (m markdownItem) Title() string       { return m.path }
func (m markdownItem) Description() string { return fmt.Sprintf("%d bytes", m.size) }
func (m markdownItem) FilterValue() string { return m.path }

type mediaItem struct {
	id       string
	path     string
	mimeType string
	size     int
	data     []byte
}

func (m mediaItem) Title() string {
	if m.path != "" {
		return m.path
	}
	return m.id
}
func (m mediaItem) Description() string { return fmt.Sprintf("%s (%d bytes)", m.mimeType, m.size) }
func (m mediaItem) FilterValue() string { return m.id }

func newBrowseModel(doc *mdocx.Document, header *headerInfo, theme string, supportsImages bool) (browseModel, error) {
	mdItems := make([]list.Item, 0, len(doc.Markdown.Files))
	for _, mf := range doc.Markdown.Files {
		mdItems = append(mdItems, markdownItem{path: mf.Path, size: len(mf.Content)})
	}
	mediaItems := make([]list.Item, 0, len(doc.Media.Items))
	for _, mi := range doc.Media.Items {
		mediaItems = append(mediaItems, mediaItem{id: mi.ID, path: mi.Path, mimeType: mi.MIMEType, size: len(mi.Data), data: mi.Data})
	}

	mdList := list.New(mdItems, list.NewDefaultDelegate(), 0, 0)
	mdList.SetShowHelp(false)
	mdList.SetFilteringEnabled(false)
	mdList.Title = "Markdown"

	mList := list.New(mediaItems, list.NewDefaultDelegate(), 0, 0)
	mList.SetShowHelp(false)
	mList.SetFilteringEnabled(false)
	mList.Title = "Media"

	metadataView := "(no metadata)"
	if doc.Metadata != nil {
		b, _ := json.MarshalIndent(doc.Metadata, "", "  ")
		metadataView = string(b)
	}

	headerView := "(header unavailable)"
	if header != nil {
		headerView = fmt.Sprintf("Magic: %s\nMagic Valid: %t\nVersion: %d\nHeader Flags: 0x%04x\nFixed Header Size: %d\nMetadata Length: %d\nReserved Clean: %t\n",
			header.MagicHex,
			header.MagicValid,
			header.Version,
			header.HeaderFlags,
			header.FixedHdrSize,
			header.MetadataLength,
			header.ReservedClean,
		)
	}

	renderer, err := buildRenderer(theme, 80)
	if err != nil {
		return browseModel{}, err
	}

	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3
	vp.SetContent("")

	model := browseModel{
		doc:            doc,
		header:         header,
		tabs:           []string{"Markdown", "Media", "Metadata", "Header"},
		activeTab:      0,
		markdownList:   mdList,
		mediaList:      mList,
		viewport:       vp,
		metadataView:   metadataView,
		headerView:     headerView,
		supportsImages: supportsImages,
		theme:          theme,
		activeStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Padding(0, 1),
		inactiveStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1),
		renderer:       renderer,
	}
	model.refreshContent()
	return model, nil
}

func (m browseModel) Init() tea.Cmd {
	return tea.SetWindowTitle("MDOCX Browser")
}

func (m browseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If viewing an image, any key exits image view
	if m.viewingImage {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.viewingImage = false
			m.imageContent = ""
			return m, nil
		}
		if _, ok := msg.(tea.MouseMsg); ok {
			m.viewingImage = false
			m.imageContent = ""
			return m, nil
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "v":
			// View image if on Media tab and an image is selected
			if m.activeTab == 1 && m.supportsImages {
				if item, ok := m.mediaList.SelectedItem().(mediaItem); ok {
					if strings.HasPrefix(strings.ToLower(item.mimeType), "image/") {
						m.imageContent = m.renderSixelImage(item.data)
						if m.imageContent != "" {
							m.viewingImage = true
							return m, nil
						}
					}
				}
			}
			return m, nil
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			m.refreshContent()
			return m, nil
		case "shift+tab", "left":
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
			m.refreshContent()
			return m, nil
		case "pgdown", "pgup", "home", "end":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	case tea.MouseMsg:
		// Handle mouse wheel scrolling
		if msg.Button == tea.MouseButtonWheelUp {
			m.viewport.LineUp(3)
			return m, nil
		}
		if msg.Button == tea.MouseButtonWheelDown {
			m.viewport.LineDown(3)
			return m, nil
		}
		// Handle mouse clicks on tabs
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft && msg.Y == 0 {
			// Clicked on the tab bar - determine which tab was clicked
			tabX := 0
			for i, t := range m.tabs {
				tabWidth := len(t) + 2 // padding of 1 on each side
				if msg.X >= tabX && msg.X < tabX+tabWidth {
					if m.activeTab != i {
						m.activeTab = i
						m.refreshContent()
					}
					return m, nil
				}
				tabX += tabWidth
			}
		}
		// Pass mouse events to child components
		switch m.activeTab {
		case 0:
			prev := m.markdownList.Index()
			var cmd tea.Cmd
			m.markdownList, cmd = m.markdownList.Update(msg)
			if m.markdownList.Index() != prev {
				m.refreshContent()
			}
			return m, cmd
		case 1:
			prev := m.mediaList.Index()
			var cmd tea.Cmd
			m.mediaList, cmd = m.mediaList.Update(msg)
			if m.mediaList.Index() != prev {
				m.refreshContent()
			}
			return m, cmd
		default:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.reflow()
		m.refreshContent()
		return m, nil
	}

	switch m.activeTab {
	case 0:
		prev := m.markdownList.Index()
		var cmd tea.Cmd
		m.markdownList, cmd = m.markdownList.Update(msg)
		if m.markdownList.Index() != prev {
			m.refreshContent()
		}
		return m, cmd
	case 1:
		prev := m.mediaList.Index()
		var cmd tea.Cmd
		m.mediaList, cmd = m.mediaList.Update(msg)
		if m.mediaList.Index() != prev {
			m.refreshContent()
		}
		return m, cmd
	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
}

func (m browseModel) View() string {
	// If viewing an image, show it full screen
	if m.viewingImage {
		return m.imageContent + "\n\n(Press any key to return)"
	}

	var tabViews []string
	for i, t := range m.tabs {
		if i == m.activeTab {
			tabViews = append(tabViews, m.activeStyle.Render(t))
		} else {
			tabViews = append(tabViews, m.inactiveStyle.Render(t))
		}
	}
	tabsLine := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)

	body := ""
	switch m.activeTab {
	case 0:
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.markdownList.View(), m.viewport.View())
	case 1:
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.mediaList.View(), m.viewport.View())
	default:
		body = m.viewport.View()
	}

	return tabsLine + "\n" + body
}

func (m *browseModel) reflow() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	m.listWidth = max(24, m.width/3)
	m.contentWidth = max(20, m.width-m.listWidth-1)
	listHeight := max(5, m.height-3)
	m.markdownList.SetSize(m.listWidth, listHeight)
	m.mediaList.SetSize(m.listWidth, listHeight)
	m.viewport.Height = listHeight
	m.applyViewportSize()

	renderer, err := buildRenderer(m.theme, m.contentWidth)
	if err == nil {
		m.renderer = renderer
	}
}

func (m *browseModel) refreshContent() {
	m.applyViewportSize()
	switch m.activeTab {
	case 0:
		m.viewport.SetContent(m.renderMarkdownSelection())
	case 1:
		m.viewport.SetContent(m.renderMediaSelection())
	case 2:
		m.viewport.SetContent(m.metadataView)
	case 3:
		m.viewport.SetContent(m.headerView)
	}
	m.viewport.GotoTop()
}

func (m *browseModel) applyViewportSize() {
	if m.width == 0 || m.height == 0 {
		return
	}
	if m.activeTab >= 2 {
		m.viewport.Width = max(20, m.width-2)
		m.viewport.Height = max(5, m.height-3)
		return
	}
	m.viewport.Width = m.contentWidth
}

func (m *browseModel) renderMarkdownSelection() string {
	item, ok := m.markdownList.SelectedItem().(markdownItem)
	if !ok {
		return ""
	}
	var content []byte
	for _, mf := range m.doc.Markdown.Files {
		if mf.Path == item.path {
			content = mf.Content
			break
		}
	}
	if len(content) == 0 {
		return "(empty markdown)"
	}
	if m.renderer == nil {
		return string(content)
	}
	out, err := m.renderer.Render(string(content))
	if err != nil {
		return string(content)
	}
	return out
}

func (m *browseModel) renderMediaSelection() string {
	item, ok := m.mediaList.SelectedItem().(mediaItem)
	if !ok {
		return ""
	}
	var b strings.Builder
	b.WriteString("ID: ")
	b.WriteString(item.id)
	b.WriteString("\n")
	if item.path != "" {
		b.WriteString("Path: ")
		b.WriteString(item.path)
		b.WriteString("\n")
	}
	if item.mimeType != "" {
		b.WriteString("MIME: ")
		b.WriteString(item.mimeType)
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("Size: %d bytes\n", item.size))

	// Show image dimensions if it's an image
	if strings.HasPrefix(strings.ToLower(item.mimeType), "image/") {
		if cfg, _, err := image.DecodeConfig(bytes.NewReader(item.data)); err == nil {
			b.WriteString(fmt.Sprintf("Dimensions: %d x %d\n", cfg.Width, cfg.Height))
		}
		if m.supportsImages {
			b.WriteString("\nPress 'v' to view image")
		}
	}
	return b.String()
}

func buildRenderer(theme string, width int) (*glamour.TermRenderer, error) {
	opts := []glamour.TermRendererOption{glamour.WithWordWrap(width)}
	if strings.TrimSpace(theme) != "" {
		opts = append(opts, glamour.WithStylePath(theme))
	} else {
		opts = append(opts, glamour.WithAutoStyle())
	}
	return glamour.NewTermRenderer(opts...)
}

// renderSixelImage renders an image to a Sixel string for full-screen display.
func (m *browseModel) renderSixelImage(data []byte) string {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return ""
	}

	// Scale to fit terminal (leave room for status line)
	maxWidthPx := m.width * 8
	maxHeightPx := (m.height - 2) * 16
	if maxWidthPx < 80 {
		maxWidthPx = 80
	}
	if maxHeightPx < 100 {
		maxHeightPx = 100
	}
	scaled := scaleImage(img, maxWidthPx, maxHeightPx)

	var buf bytes.Buffer
	if err := sixel.NewEncoder(&buf).Encode(scaled); err != nil {
		return ""
	}
	return buf.String()
}

func supportsSixel() bool {
	// Check TERM for known Sixel-capable terminals
	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	termProgram := strings.ToLower(strings.TrimSpace(os.Getenv("TERM_PROGRAM")))
	wtSession := os.Getenv("WT_SESSION")

	// Windows Terminal supports Sixel
	if wtSession != "" {
		return true
	}
	// iTerm2, WezTerm, mintty, foot, contour support Sixel
	if strings.Contains(termProgram, "iterm") ||
		strings.Contains(termProgram, "wezterm") ||
		strings.Contains(termProgram, "mintty") ||
		strings.Contains(termProgram, "contour") {
		return true
	}
	if term == "" {
		return false
	}
	if strings.Contains(term, "sixel") ||
		strings.Contains(term, "xterm") ||
		strings.Contains(term, "mlterm") ||
		strings.Contains(term, "foot") {
		return true
	}
	return false
}

// scaleImage scales an image to fit within maxWidth x maxHeight while preserving aspect ratio.
func scaleImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	if origWidth <= 0 || origHeight <= 0 {
		return img
	}

	if origWidth <= maxWidth && origHeight <= maxHeight {
		return img
	}

	// Calculate scale factor
	scaleW := float64(maxWidth) / float64(origWidth)
	scaleH := float64(maxHeight) / float64(origHeight)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}

	newWidth := int(float64(origWidth) * scale)
	newHeight := int(float64(origHeight) * scale)
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}
