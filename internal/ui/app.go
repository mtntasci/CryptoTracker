package ui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"crypto-tracker/internal/attribution"
	"crypto-tracker/internal/chain"
	"crypto-tracker/internal/chain/account"
	"crypto-tracker/internal/chain/utxo"
	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
	"crypto-tracker/internal/tracer"
	"crypto-tracker/internal/ui/views"
)

// AppState represents the active screen in the Bubbletea state machine.
type AppState int

const (
	StateInput AppState = iota
	StateTracing
	StateResult
)

type traceEventMsg struct {
	Event tracer.TraceEvent
	Ok    bool
}

// AppModel is the root Bubbletea model for CryptoTracker.
type AppModel struct {
	cfg           *config.Config
	state         AppState
	selectedChain model.ChainType
	textInput     textinput.Model
	spinner       spinner.Model
	errorMsg      string

	// Active tracing state
	cancelTrace    context.CancelFunc
	eventCh        <-chan tracer.TraceEvent
	rootAddress    string
	currentAddress string
	currentDepth   int
	visitedCount   int
	queueSize      int
	statusMessage  string
	steps          []model.TraceStep
	skipped        []model.SkippedOutput
	exchangeHits   []model.TraceStep

	// Result view state
	finalResult  *model.TraceResult
	scrollOffset int
}

// NewAppModel initializes the root Bubbletea model. If initialAddress is non-empty,
// it immediately launches tracing on startup.
func NewAppModel(cfg *config.Config, initialAddress string, initialChain model.ChainType) *AppModel {
	ti := textinput.New()
	ti.Placeholder = "e.g. bc1q..., 0x..., or T..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 68

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(views.ColorPrimary)

	selected := initialChain
	if selected == "" {
		selected = model.ChainBitcoin
	}

	m := &AppModel{
		cfg:           cfg,
		state:         StateInput,
		selectedChain: selected,
		textInput:     ti,
		spinner:       sp,
	}

	if strings.TrimSpace(initialAddress) != "" {
		m.textInput.SetValue(strings.TrimSpace(initialAddress))
		if detected, err := model.DetectChainFromAddress(initialAddress); err == nil && initialChain == "" {
			m.selectedChain = detected
		}
	}

	return m
}

func (m *AppModel) Init() tea.Cmd {
	if strings.TrimSpace(m.textInput.Value()) != "" {
		cmd := m.startTracingSession(strings.TrimSpace(m.textInput.Value()))
		return tea.Batch(textinput.Blink, m.spinner.Tick, cmd)
	}
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.cancelTrace != nil {
				m.cancelTrace()
			}
			return m, tea.Quit

		case "q":
			if m.state != StateInput {
				if m.cancelTrace != nil {
					m.cancelTrace()
				}
				return m, tea.Quit
			}
		}

		switch m.state {
		case StateInput:
			switch msg.String() {
			case "tab":
				m.cycleChain()
				return m, nil
			case "enter":
				addr := strings.TrimSpace(m.textInput.Value())
				if addr == "" {
					m.errorMsg = "Please enter a valid blockchain address."
					return m, nil
				}
				if detected, err := model.DetectChainFromAddress(addr); err == nil {
					m.selectedChain = detected
				} else {
					m.errorMsg = err.Error()
					return m, nil
				}
				m.errorMsg = ""
				cmd := m.startTracingSession(addr)
				return m, tea.Batch(m.spinner.Tick, cmd)
			}

			var tiCmd tea.Cmd
			m.textInput, tiCmd = m.textInput.Update(msg)
			// Auto-detect chain while typing
			if detected, err := model.DetectChainFromAddress(m.textInput.Value()); err == nil {
				m.selectedChain = detected
				m.errorMsg = ""
			}
			return m, tiCmd

		case StateResult:
			switch msg.String() {
			case "up", "k":
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			case "down", "j":
				if m.finalResult != nil && m.scrollOffset+8 < len(m.finalResult.Steps) {
					m.scrollOffset++
				}
			case "n", "r":
				m.state = StateInput
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.finalResult = nil
				m.scrollOffset = 0
				return m, textinput.Blink
			}
		}

	case spinner.TickMsg:
		if m.state == StateTracing {
			var spCmd tea.Cmd
			m.spinner, spCmd = m.spinner.Update(msg)
			cmds = append(cmds, spCmd)
		}

	case traceEventMsg:
		if !msg.Ok {
			m.state = StateResult
			return m, nil
		}

		ev := msg.Event
		m.statusMessage = ev.Message
		if ev.CurrentAddr != "" {
			m.currentAddress = ev.CurrentAddr
		}
		if ev.Depth > 0 {
			m.currentDepth = ev.Depth
		}
		m.visitedCount = ev.VisitedCnt
		m.queueSize = ev.QueueSize

		switch ev.Type {
		case tracer.EventHopDiscovered:
			if ev.Step != nil {
				m.steps = append(m.steps, *ev.Step)
			}
		case tracer.EventExchangeHit:
			if ev.Step != nil {
				m.steps = append(m.steps, *ev.Step)
				m.exchangeHits = append(m.exchangeHits, *ev.Step)
			}
		case tracer.EventOutputSkipped:
			if ev.Skipped != nil {
				m.skipped = append(m.skipped, *ev.Skipped)
			}
		case tracer.EventComplete:
			m.finalResult = ev.Result
			m.state = StateResult
			return m, nil
		}

		cmds = append(cmds, waitForNextTraceEvent(m.eventCh))
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) View() string {
	switch m.state {
	case StateInput:
		return views.RenderInputView(views.InputViewProps{
			SelectedChain: m.selectedChain,
			TextInputView: m.textInput.View(),
			ErrorMessage:  m.errorMsg,
			Config:        m.cfg,
		})

	case StateTracing:
		return views.RenderTracingView(views.TracingViewProps{
			Chain:          m.selectedChain,
			RootAddress:    m.rootAddress,
			SpinnerView:    m.spinner.View(),
			CurrentDepth:   m.currentDepth,
			MaxDepth:       m.cfg.MaxDepth,
			VisitedCount:   m.visitedCount,
			QueueSize:      m.queueSize,
			CurrentAddress: m.currentAddress,
			StatusMessage:  m.statusMessage,
			Steps:          m.steps,
			Skipped:        m.skipped,
			ExchangeHits:   m.exchangeHits,
		})

	case StateResult:
		return views.RenderResultView(views.ResultViewProps{
			Result:       m.finalResult,
			ScrollOffset: m.scrollOffset,
		})

	default:
		return ""
	}
}

func (m *AppModel) cycleChain() {
	switch m.selectedChain {
	case model.ChainBitcoin:
		m.selectedChain = model.ChainEthereum
	case model.ChainEthereum:
		m.selectedChain = model.ChainTron
	default:
		m.selectedChain = model.ChainBitcoin
	}
}

func (m *AppModel) startTracingSession(address string) tea.Cmd {
	if m.cancelTrace != nil {
		m.cancelTrace()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelTrace = cancel
	m.state = StateTracing
	m.rootAddress = address
	m.currentAddress = address
	m.currentDepth = 0
	m.visitedCount = 0
	m.queueSize = 1
	m.statusMessage = "Initializing chain provider and attribution engine..."
	m.steps = nil
	m.skipped = nil
	m.exchangeHits = nil
	m.finalResult = nil

	var provider chain.Provider
	switch m.selectedChain {
	case model.ChainBitcoin:
		provider = utxo.NewBitcoinProvider(m.cfg)
	case model.ChainEthereum:
		provider = account.NewEthereumProvider(m.cfg)
	case model.ChainTron:
		provider = account.NewTronProvider(m.cfg)
	default:
		provider = utxo.NewBitcoinProvider(m.cfg)
	}

	attrEngine := attribution.NewEngine(m.cfg)
	traceEngine := tracer.NewEngine(m.cfg, provider, attrEngine)

	ch := make(chan tracer.TraceEvent, 64)
	m.eventCh = ch

	go func() {
		_, _ = traceEngine.Trace(ctx, address, ch)
	}()

	return waitForNextTraceEvent(ch)
}

func waitForNextTraceEvent(ch <-chan tracer.TraceEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		return traceEventMsg{
			Event: ev,
			Ok:    ok,
		}
	}
}

