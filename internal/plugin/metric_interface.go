package plugin

import (
	"github.com/LNOpenMetrics/go-lnmetrics.reporter/pkg/graphql"

	"github.com/vincenzopalazzo/glightning/glightning"
)

// mapping the internal id with the name of the metrics.
// the id is passed by the plugin RPC name.
var MetricsSupported map[int]string

// 0 = outcoming
// 1 = incoming
// 2 = mutual.
var ChannelDirections map[int]string

func init() {
	MetricsSupported = make(map[int]string)
	MetricsSupported[1] = "metric_one"

	ChannelDirections = make(map[int]string)
	ChannelDirections[0] = "OUTCOMING"
	ChannelDirections[1] = "INCOOMING"
}

// All the metrics need to respect this interface
type Metric interface {
	// return the name of the metric
	MetricName() *string

	// call this to initialized the metric with node
	// information if any error occurs, a not nil value is
	// returned
	OnInit(lightning *glightning.Lightning) error

	// Call this method when the close rpc method is called
	OnClose(msg *Msg, lightning *glightning.Lightning) error

	// Call this method to make the status of the metrics persistent
	MakePersistent() error

	// Method to store the run a callback to upload the content on the server.
	// TODO: Use an interface to generalize the client, it can be also a rest api
	// move accept some interface later.
	UploadOnRepo(client *graphql.Client, lightning *glightning.Lightning) error

	// Method to store the run a callback to init the content on the server
	// the first time that the plugin in ran.
	InitOnRepo(client *graphql.Client, lightning *glightning.Lightning) error

	// Call this method when you want update all the metrics without
	// some particular event throw from c-lightning
	Update(lightning *glightning.Lightning) error

	// Class this method when you want catch some event from
	// c-lightning and make some operation on the metrics data.
	UpdateWithMsg(message *Msg, lightning *glightning.Lightning) error

	// convert the object into a json
	ToJSON() (string, error)

	// Migrate to a new version of the metrics, some new version of the plugin
	// ca contains some format change and for this reason, this method help
	// to migrate the old version to a new version.
	Migrate(payload map[string]interface{}) error
}

// Message struct to pass from the plugin to the metric
type Msg struct {
	// The message is from a command? if not it is nil
	cmd string
	// the map of parameter that the plugin need to feel.
	params map[string]interface{}
}
