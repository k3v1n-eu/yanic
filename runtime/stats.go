package runtime

const (
	DISABLED_AUTOUPDATER = "disabled"
	GLOBAL_SITE          = "global"
)

// CounterMap to manage multiple values
type CounterMap map[string]uint32

// GlobalStats struct
type GlobalStats struct {
	Clients       uint32
	ClientsWifi   uint32
	ClientsWifi24 uint32
	ClientsWifi5  uint32
	Gateways      uint32
	Nodes         uint32

	Firmwares   CounterMap
	Models      CounterMap
	Autoupdater CounterMap
}

//NewGlobalStats returns global statistics for InfluxDB
func NewGlobalStats(nodes *Nodes, sites []string) (result map[string]*GlobalStats) {
	result = make(map[string]*GlobalStats)

	result[GLOBAL_SITE] = &GlobalStats{
		Firmwares:   make(CounterMap),
		Models:      make(CounterMap),
		Autoupdater: make(CounterMap),
	}

	for _, site := range sites {
		result[site] = &GlobalStats{
			Firmwares:   make(CounterMap),
			Models:      make(CounterMap),
			Autoupdater: make(CounterMap),
		}
	}

	nodes.RLock()
	for _, node := range nodes.List {
		if node.Online {
			result[GLOBAL_SITE].Add(node)

			if info := node.Nodeinfo; info != nil {
				site := info.System.SiteCode
				if _, exist := result[site]; exist {
					result[site].Add(node)
				}
			}
		}
	}
	nodes.RUnlock()
	return
}

// Add values to GlobalStats
// if node is online
func (s *GlobalStats) Add(node *Node) {
	s.Nodes++
	if stats := node.Statistics; stats != nil {
		s.Clients += stats.Clients.Total
		s.ClientsWifi24 += stats.Clients.Wifi24
		s.ClientsWifi5 += stats.Clients.Wifi5
		s.ClientsWifi += stats.Clients.Wifi
	}
	if node.IsGateway() {
		s.Gateways++
	}
	if info := node.Nodeinfo; info != nil {
		s.Models.Increment(info.Hardware.Model)
		s.Firmwares.Increment(info.Software.Firmware.Release)
		if info.Software.Autoupdater.Enabled {
			s.Autoupdater.Increment(info.Software.Autoupdater.Branch)
		} else {
			s.Autoupdater.Increment(DISABLED_AUTOUPDATER)
		}
	}
}

// Increment counter in the map by one
// if the value is not empty
func (m CounterMap) Increment(key string) {
	if key != "" {
		m[key]++
	}
}
