package stats

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ricochet2200/go-disk-usage/du"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
	"time"
)

type Stats struct {
	plugins.Base
}

type StatsResponse struct {
	Results []StatsResult `json:"results"`
}
type StatsResult struct {
	Mount string `json:"mount"`
	Size  uint64 `json:"size"`
	Used  uint64 `json:"used"`
}
type StatsArgs struct {
	Mounts []string `json:"mounts"`
}
type CpuResponse struct {
	Times []cpu.TimesStat `json:"times"`
	Info  []cpu.InfoStat  `json:"info"`
}
type MemoryResponse struct {
	Memory *mem.VirtualMemoryStat `json:"memory"`
	Swap   *mem.SwapMemoryStat    `json:"swap"`
}

func (self *Stats) RegisterRPC(server *rpc.Server) {
	server.Register(self)
}

func New() *Stats {
	return &Stats{plugins.Base{Name: "stats", Version: 1}}
}

type NetResult struct {
	Device string `json:"device"`
	TxRate uint64 `json:"tx_rate"`
	RxRate uint64 `json:"rx_rate"`
}

func (s *Stats) Net(args int, res *[]*NetResult) error {
	results, err := net.IOCounters(true)
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	results2, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	b := []*NetResult{}
	for i, r := range results {
		netResult := NetResult{Device: r.Name}
		netResult.TxRate = results2[i].BytesSent - r.BytesSent
		netResult.RxRate = results2[i].BytesRecv - r.BytesRecv
		b = append(b, &netResult)
	}

	*res = b

	return nil
}

func (s *Stats) Load(args int, res *load.AvgStat) error {
	l, err := load.Avg()
	*res = *l
	if err != nil {
		return err
	}
	return nil
}

func (s *Stats) Host(args int, res *host.InfoStat) error {
	h, err := host.Info()
	*res = *h
	if err != nil {
		return err
	}
	return nil
}

func (s *Stats) Cpu(args int, res *CpuResponse) error {
	log.WithFields(log.Fields{
		"plugin": s.GetName(),
		"method": "Cpu",
	}).Debug("Gathering CPU stats")

	t, err := cpu.Times(false)
	i, err := cpu.Info()
	if err != nil {
		return err
	}

	res.Times = t
	res.Info = i

	return nil
}

func (s *Stats) Memory(args *MemoryResponse, res *MemoryResponse) error {
	log.WithFields(log.Fields{
		"plugin": s.GetName(),
		"method": "Memory",
	}).Debug("Gathering MEM stats")

	v, err := mem.VirtualMemory()
	swap, err := mem.SwapMemory()

	if err != nil {
		return err
	}

	res.Memory = v
	res.Swap = swap
	return nil
}

func (s *Stats) DiskSpace(args *StatsArgs, res *StatsResponse) error {
	log.WithFields(log.Fields{
		"plugin": s.GetName(),
		"method": "Diskspace",
	}).Debug("Gathering Diskspace stats")

	if len(args.Mounts) > 0 {
		for _, m := range args.Mounts {
			res.Results = append(res.Results, s.diskspaceResult(m))
		}
	} else {
		res.Results = append(res.Results, s.diskspaceResult("/"))
	}
	return nil
}

func (s *Stats) diskspaceResult(mount string) StatsResult {
	usage := du.NewDiskUsage(mount)
	return StatsResult{Mount: mount, Size: usage.Size(), Used: usage.Used()}
}
