package plugin

import (
	"errors"
	"fmt"
	"time"

	"github.com/niftynei/glightning/glightning"
	"github.com/robfig/cron/v3"

	"github.com/OpenLNMetrics/go-metrics-reported/pkg/graphql"
	"github.com/OpenLNMetrics/go-metrics-reported/pkg/log"
)

type MetricsPlugin struct {
	Plugin  *glightning.Plugin
	Metrics map[int]Metric
	Rpc     *glightning.Lightning
	Cron    *cron.Cron
	Server  *graphql.Client
}

func (plugin *MetricsPlugin) HendlerRPCMessage(event *glightning.RpcCommandEvent) error {
	command := event.Cmd
	switch command.MethodName {
	case "stop":
		// Share to all the metrics, so we need a global method that iterate over the metrics map
		params := make(map[string]interface{})
		params["timestamp"] = time.Now()
		msg := Msg{"stop", params}
		for _, metric := range plugin.Metrics {
			go plugin.callOnStopOnMetrics(metric, &msg)
		}
		plugin.Cron.Stop()
		log.GetInstance().Info("Close command received")
	default:
		return nil
	}
	return nil
}

func (plugin *MetricsPlugin) RegisterMetrics(id int, metric Metric) error {
	_, ok := plugin.Metrics[id]
	if ok {
		log.GetInstance().Error(fmt.Sprintf("Metrics with is %d already registered.", id))
		return errors.New(fmt.Sprintf("Metrics with is %d already registered.", id))
	}
	plugin.Metrics[id] = metric
	return nil
}

func (plugin *MetricsPlugin) RegisterMethods() {
	method := NewMetricPlugin()
	rpcMethod := glightning.NewRpcMethod(method, "Show diagnostic node")
	rpcMethod.LongDesc = "Show the diagnostic data of the lightning network node"
	rpcMethod.Category = "metrics"
	plugin.Plugin.RegisterMethod(rpcMethod)

	infoMethod := NewPluginRpcMethod()
	infoRpcMethod := glightning.NewRpcMethod(infoMethod, "Show go-lnmetrics-reporter info")
	infoRpcMethod.Category = "metrics"
	infoRpcMethod.LongDesc = "Return a map where the key is the id of the method and the value is the payload of the metric. The metrics_id is a string that conatins the id divided by a comma. An example is \"diagnostic \"1,2,3\"\""
	plugin.Plugin.RegisterMethod(infoRpcMethod)
}

func (instance *MetricsPlugin) callUpdateOnMetric(metric Metric, msg *Msg) {
	metric.UpdateWithMsg(msg, instance.Rpc)
}

func (instance *MetricsPlugin) callOnStopOnMetrics(metric Metric, msg *Msg) {
	err := metric.OnClose(msg, instance.Rpc)
	if err != nil {
		log.GetInstance().Error(err)
	}
}

func (instance *MetricsPlugin) callUpdateOnMetricNoMsg(metric Metric) {
	log.GetInstance().Debug("Calling Update on metrics")
	err := metric.Update(instance.Rpc)
	if err != nil {
		log.GetInstance().Error(fmt.Sprintf("Error %s", err))
	}
}

func (instance *MetricsPlugin) RegisterRecurrentEvt(after string) {
	instance.Cron = cron.New()
	instance.Cron.AddFunc(after, func() {
		log.GetInstance().Debug("Calling recurrent")
		for _, metric := range instance.Metrics {
			go instance.callUpdateOnMetricNoMsg(metric)
		}
	})
}

func (instance *MetricsPlugin) RegisterOneTimeEvt(after string) {
	duration, err := time.ParseDuration(after)
	if err != nil {
		log.GetInstance().Error(fmt.Sprintf("Error in the on time evt: %s", err))
		return
	}
	time.AfterFunc(duration, func() {
		log.GetInstance().Debug("Calling on time function function")
		// TODO: Should C-Lightning send a on init event like notification?
		for _, metric := range instance.Metrics {
			go metric.OnInit(instance.Rpc)
		}
	})
}
