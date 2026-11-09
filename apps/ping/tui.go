package ping

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/flily/etherwind/winds"
)

var (
	columns = []table.Column{
		{Title: "target", Width: 16},
		{Title: "icmp_seq", Width: 8},
		{Title: "ttl", Width: 4},
		{Title: "time", Width: 12},
		{Title: "Rx / Tx", Width: 12},
		{Title: "loss", Width: 8},
		{Title: "total_time", Width: 15},
		{Title: "min", Width: 12},
		{Title: "max", Width: 12},
		{Title: "avg", Width: 12},
	}
)

type PingResult struct {
	Result *winds.PingResult
	Record *PingRecord
	Err    error
}

type model struct {
	table   table.Model
	ch      chan *PingResult
	params  *Params
	running atomic.Int32
	records []*PingRecord
}

func newModel(params *Params, records []*PingRecord) *model {
	t := table.New(
		table.WithColumns(columns),
		table.WithHeight(1+len(records)),
		table.WithWidth(tableWidth(columns)),
		table.WithFocused(true),
	)
	s := table.DefaultStyles()
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := &model{
		table:   t,
		ch:      make(chan *PingResult, len(records)),
		params:  params,
		running: atomic.Int32{},
		records: records,
	}

	return m
}

func tableWidth(columns []table.Column) int {
	width := 0
	for _, col := range columns {
		width += col.Width
	}

	if len(columns) > 1 {
		// two-space border by default
		width += (2 * len(columns)) + 1
	}

	return width
}

func (m *model) makePingRoutine(params *Params, group int, ch chan<- *PingResult, record *PingRecord, pinger *winds.Pinger) {
	addr := record.IP
	payloadBase := winds.DefaultPingPayloadBase

	defer func() {
		_ = pinger.Close()
		ch <- nil
	}()

	count := params.Count
	seq := 1
	id := (os.Getpid() & 0x00ff) | ((group & 0x00ff) << 8)
	for count < 0 || seq <= count {
		// fmt.Fprintf(os.Stderr, "PING %s count=%d seq=%d\n", addr, count, seq)
		result, err := pinger.Ping(addr, id, seq, payloadBase)
		msg := &PingResult{
			Result: result,
			Record: record,
			Err:    err,
		}
		ch <- msg

		seq = (seq + 1) & 0xffff
		time.Sleep(1 * params.Interval)
	}
}

func (m *model) makePingReceiver(ch <-chan *PingResult) tea.Cmd {
	return func() tea.Msg {
		item := <-ch
		// fmt.Fprintf(os.Stderr, "RECV: %+v\n", item)
		return item
	}
}

func (m *model) Init() tea.Cmd {
	ch := m.ch

	m.running.Store(int32(len(m.records)))
	for i, record := range m.records {
		addr := record.IP
		network := winds.NetworkIPv4
		if addr.To4() == nil {
			network = winds.NetworkIPv6
		}

		pinger, err := winds.NewPinger(network)
		if err != nil {
			errMessage := rootError(err)
			if isPermissionDenied(err) {
				fmt.Printf("No permission to start ping: %s\n", errMessage.Error())
				if runtime.GOOS == "linux" {
					fmt.Printf("Hint: use `sudo setcap cap_net_raw=+ep etherwind` or `sudo` to get root privileges\n")
				}

			} else {
				fmt.Printf("Failed to create pinger: %s\n", errMessage.Error())
			}

			return tea.Quit
		}

		go m.makePingRoutine(m.params, i, ch, record, pinger)
	}

	return m.makePingReceiver(ch)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var nextCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tea.QuitMsg:
		fmt.Printf("Quit message received\n")
		return m, tea.Quit

	case *PingResult:
		if msg == nil {
			// ping goroutine finished
			current := m.running.Add(-1)
			if current > 0 {
				return m, m.makePingReceiver(m.ch)
			}

			return m, tea.Quit
		}

		record := msg.Record
		// fmt.Fprintf(os.Stderr, "UPDATE: recv %+v\n", msg)
		result, err := msg.Result, msg.Err
		if err != nil {
			// fmt.Fprintf(os.Stderr, "!!!!!!!!!!\n")
			record.AddFailure()

		} else {
			pingTimeMs := float64(result.Duration) / float64(time.Millisecond)
			record.Add(pingTimeMs, result.TTL)
		}

		return m, m.makePingReceiver(m.ch)
	}

	m.table, nextCmd = m.table.Update(msg)
	return m, nextCmd
}

func (m *model) makeRow(record *PingRecord) table.Row {
	row := make([]string, 0, len(columns))

	row = append(row,
		record.IP.String(),                                   // target
		fmt.Sprintf("%d", len(record.Records)),               // icmp_seq
		fmt.Sprintf("%d", record.LastTTL),                    // ttl
		fmt.Sprintf("%.3f ms", record.LastTime),              // time
		fmt.Sprintf("%d / %d", record.Success, record.Count), // transmitted / received
		fmt.Sprintf("%.1f %%", record.PacketLoss()),          // loss
		fmt.Sprintf("%.1f ms", record.TimeCostMs()),          // total_time
		fmt.Sprintf("%.3f ms", record.Min()),                 // min
		fmt.Sprintf("%.3f ms", record.Max()),                 // max
		fmt.Sprintf("%.3f ms", record.Average()),             // avg
	)

	return table.Row(row)
}

func (m *model) makeRows() []table.Row {
	rows := make([]table.Row, 0, len(m.records))
	for _, record := range m.records {
		rows = append(rows, m.makeRow(record))
	}

	return rows
}

func (m *model) View() tea.View {
	rows := m.makeRows()
	m.table.SetRows(rows)

	return tea.NewView(m.table.View() + "\n")
}

func runTuiPing(params *Params, records []*PingRecord) {
	m := newModel(params, records)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("Failed to run TUI: %s\n", err.Error())
	}
}

func MainTui(params *Params) {
	if len(params.Target) > 256 {
		fmt.Printf("error: too many targets (maximum 256)\n")
		return
	}

	records := make([]*PingRecord, 0, len(params.Target))
	for _, target := range params.Target {
		ip := net.ParseIP(target)
		if ip == nil {
			fmt.Printf("Invalid IP address: %s\n", target)
			return
		}
		records = append(records, NewPingRecordWithIP(ip, DefaultPingRecordCapacity))
	}

	runTuiPing(params, records)
}
