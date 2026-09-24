package ping

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/flily/etherwind/winds"
)

type PingResult struct {
	Result *winds.PingResult
	Record *PingRecord
	Err    error
}

func makePingRoutine(params *Params, ch chan<- *PingResult, record *PingRecord, pinger *winds.Pinger) {
	addr := record.IP

	payloadBase := winds.DefaultPingPayloadBase

	count := params.Count
	seq := 1
	id := os.Getpid() & 0xffff
	for count < 0 || seq <= count {
		// fmt.Fprintf(os.Stderr, "PING %s count=%d seq=%d\n", addr, count, seq)
		result, err := pinger.Ping(addr, id, seq, payloadBase)
		msg := &PingResult{
			Result: result,
			Record: record,
			Err:    err,
		}
		ch <- msg

		seq++
		time.Sleep(1 * time.Second)
	}
}

func makePingReceiver(ch <-chan *PingResult) tea.Cmd {
	return func() tea.Msg {
		item := <-ch
		// fmt.Fprintf(os.Stderr, "RECV: %+v\n", item)
		return item
	}
}

type model struct {
	ch      chan *PingResult
	params  *Params
	records []*PingRecord
}

func newModel(params *Params, records []*PingRecord) *model {
	m := &model{
		ch:      make(chan *PingResult, 10),
		params:  params,
		records: records,
	}

	return m
}

func (m *model) Init() tea.Cmd {
	ch := m.ch

	for _, record := range m.records {
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

		go makePingRoutine(m.params, ch, record, pinger)
	}

	return makePingReceiver(ch)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		}

	case *PingResult:
		record := msg.Record
		// fmt.Fprintf(os.Stderr, "UPDATE: recv %+v\n", msg)
		result, err := msg.Result, msg.Err
		if err != nil {
			// fmt.Fprintf(os.Stderr, "!!!!!!!!!!\n")
			record.AddFailure()

		} else {
			pingTimeMs := float64(result.Duration) / float64(time.Millisecond)
			record.Add(pingTimeMs)
		}
		return m, makePingReceiver(m.ch)
	}

	return m, nil
}

func (m *model) View() tea.View {
	buf := make([]string, 0, len(m.records))
	for _, record := range m.records {
		ms := 0.0
		if len(record.Records) > 0 {
			ms = record.Records[len(record.Records)-1]
		}
		s := fmt.Sprintf("ping %s  ->>  %d / %d  latest=%.2fms",
			record.IP, len(record.Records), record.Count, ms)

		// fmt.Fprintf(os.Stderr, "[%d] VIEW: r = %+v\n", i, record)
		buf = append(buf, s)
	}
	return tea.NewView(strings.Join(buf, "\n") + "\n")
}

func runTuiPing(params *Params, records []*PingRecord) {
	m := newModel(params, records)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("Failed to run TUI: %s\n", err.Error())
	}
}

func MainTui(params *Params) {
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
