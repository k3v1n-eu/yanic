package influxdb

import (
	"testing"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/stretchr/testify/assert"

	"github.com/FreifunkBremen/yanic/tree/master/data"
	"github.com/FreifunkBremen/yanic/tree/master/runtime"
)

func TestToInflux(t *testing.T) {
	assert := assert.New(t)

	node := &runtime.Node{
		Statistics: &data.Statistics{
			NodeID:      "deadbeef",
			LoadAverage: 0.5,
			Wireless: data.WirelessStatistics{
				&data.WirelessAirtime{Frequency: 5500},
			},
			Traffic: struct {
				Tx      *data.Traffic `json:"tx"`
				Rx      *data.Traffic `json:"rx"`
				Forward *data.Traffic `json:"forward"`
				MgmtTx  *data.Traffic `json:"mgmt_tx"`
				MgmtRx  *data.Traffic `json:"mgmt_rx"`
			}{
				Tx:      &data.Traffic{Dropped: 1321},
				Rx:      &data.Traffic{Bytes: 1213},
				Forward: &data.Traffic{Bytes: 1322},
				MgmtTx:  &data.Traffic{Packets: 2327},
				MgmtRx:  &data.Traffic{Bytes: 2331},
			},
			MeshVPN: &data.MeshVPN{
				Groups: map[string]*data.MeshVPNPeerGroup{
					"ffhb": {
						Peers: map[string]*data.MeshVPNPeerLink{
							"vpn01": {Established: 3},
							"vpn02": {},
							"trash": nil,
							"vpn03": {Established: 0},
						},
					},
				},
			},
		},
		Nodeinfo: &data.NodeInfo{
			NodeID: "deadbeef",
			Owner: &data.Owner{
				Contact: "nobody",
			},
			System: data.System{
				SiteCode: "ffxx",
			},
			Wireless: &data.Wireless{
				TxPower24: 3,
				Channel24: 4,
			},
			Network: data.Network{
				Mac: "DEADMAC",
			},
			Software: data.Software{
				Autoupdater: struct {
					Enabled bool   `json:"enabled,omitempty"`
					Branch  string `json:"branch,omitempty"`
				}{
					Enabled: true,
					Branch:  "testing",
				},
			},
		},
		Neighbours: &data.Neighbours{
			NodeID: "deadbeef",
			Batadv: map[string]data.BatadvNeighbours{
				"a-interface": {
					Neighbours: map[string]data.BatmanLink{
						"BAFF1E5": {
							Tq: 204,
						},
					},
				},
			},
			LLDP: map[string]data.LLDPNeighbours{
				"b-interface": {},
			},
		},
	}

	neighbour := &runtime.Node{
		Nodeinfo: &data.NodeInfo{
			NodeID: "foobar",
			Network: data.Network{
				Mac: "BAFF1E5",
			},
			Software: data.Software{
				Autoupdater: struct {
					Enabled bool   `json:"enabled,omitempty"`
					Branch  string `json:"branch,omitempty"`
				}{
					Enabled: false,
				},
			},
		},
		Statistics: &data.Statistics{
			NodeID: "foobar",
		},
	}

	// do not add a empty statistics of a node
	droppednode := &runtime.Node{
		Nodeinfo: &data.NodeInfo{
			NodeID: "notfound",
			Network: data.Network{
				Mac: "instats",
			},
		},
		Statistics: &data.Statistics{},
	}

	points := testPoints(node, neighbour, droppednode)
	var fields map[string]interface{}
	var tags map[string]string

	assert.Len(points, 3)

	// first point contains the neighbour
	sPoint := points[0]
	tags = sPoint.Tags()
	fields, _ = sPoint.Fields()

	assert.EqualValues("deadbeef", tags["nodeid"])
	assert.EqualValues("nobody", tags["owner"])
	assert.EqualValues("testing", tags["autoupdater"])
	assert.EqualValues("ffxx", tags["site"])
	assert.EqualValues(0.5, fields["load"])
	assert.EqualValues(0, fields["neighbours.lldp"])
	assert.EqualValues(1, fields["neighbours.batadv"])
	assert.EqualValues(1, fields["neighbours.vpn"])
	assert.EqualValues(1, fields["neighbours.total"])

	assert.EqualValues(uint32(3), fields["wireless.txpower24"])
	assert.EqualValues(uint32(5500), fields["airtime11a.frequency"])
	assert.EqualValues("", tags["frequency5500"])

	assert.EqualValues(int64(1213), fields["traffic.rx.bytes"])
	assert.EqualValues(float64(1321), fields["traffic.tx.dropped"])
	assert.EqualValues(int64(1322), fields["traffic.forward.bytes"])
	assert.EqualValues(int64(2331), fields["traffic.mgmt_rx.bytes"])
	assert.EqualValues(float64(2327), fields["traffic.mgmt_tx.packets"])

	// second point contains the link
	nPoint := points[1]
	tags = nPoint.Tags()
	fields, _ = nPoint.Fields()
	assert.EqualValues("link", nPoint.Name())
	assert.EqualValues(map[string]string{
		"source.id":  "deadbeef",
		"source.mac": "a-interface",
		"target.id":  "foobar",
		"target.mac": "BAFF1E5",
	}, tags)
	assert.EqualValues(80, fields["tq"])

	// third point contains the neighbour
	nPoint = points[2]
	tags = nPoint.Tags()
	assert.EqualValues("disabled", tags["autoupdater"])
}

// Processes data and returns the InfluxDB points
func testPoints(nodes ...*runtime.Node) (points []*client.Point) {
	// Create dummy client
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{Addr: "http://127.0.0.1"})
	if err != nil {
		panic(err)
	}

	nodesList := runtime.NewNodes(&runtime.NodesConfig{})

	// Create dummy connection
	conn := &Connection{
		points: make(chan *client.Point),
		client: influxClient,
	}

	for _, node := range nodes {
		nodesList.AddNode(node)
	}

	// Process data
	go func() {
		for _, node := range nodes {
			conn.InsertNode(node)
			if node.Neighbours != nil {
				for _, link := range nodesList.NodeLinks(node) {
					conn.InsertLink(&link, node.Lastseen.GetTime())
				}
			}
		}
		conn.Close()
	}()

	// Read points
	for point := range conn.points {
		points = append(points, point)
	}

	return
}
