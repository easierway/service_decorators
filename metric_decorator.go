package service_decorators

import (
	"time"

	"github.com/easierway/g_met"
)

const (
	//TimeSpent is the metric item name of the time spent
	TimeSpent = "time_spent"
	//OccurredError is metric item name of the occurred error
	OccurredError = "occurred_error"
)

//ErrorClassifier is to decide the type/class of the error
//the return values are
//string: the type/class
//bool: if the error needs to be put into the metrics
type ErrorClassifier func(err error) (string, bool)

//MetricDecoratorConfig is the configuration MetricDecorator
type MetricDecoratorConfig struct {
	errorClassifier         ErrorClassifier
	needsRecordingTimeSpent bool
	metricSender            g_met.GMet
}

//MetricDecorator is to introduce the metrics of the service.
//It is based on GMet (https://github.com/easierway/g_met)
type MetricDecorator struct {
	config *MetricDecoratorConfig
}

//CreateMetricDecorator is the helper method of
//creating CreateMetricDecorator instance.
//The settings can be defined by WithXX method chain
func CreateMetricDecorator(gmetInstance g_met.GMet) *MetricDecoratorConfig {
	return &MetricDecoratorConfig{metricSender: gmetInstance}
}

//WithErrorClassifier is to set the ErrorClassifier
func (config *MetricDecoratorConfig) WithErrorClassifier(errClassifier ErrorClassifier) *MetricDecoratorConfig {
	config.errorClassifier = errClassifier
	return config
}

//NeedsRecordingTimeSpent is to turn on the time spent metrics
func (config *MetricDecoratorConfig) NeedsRecordingTimeSpent() *MetricDecoratorConfig {
	config.needsRecordingTimeSpent = true
	return config
}

//Build is to create a CreateMetricDecorator instance according to the settings
func (config *MetricDecoratorConfig) Build() *MetricDecorator {
	return &MetricDecorator{config}
}

//Decorate is to add the metrics logic to the inner service function
func (dec *MetricDecorator) Decorate(innerFn ServiceFunc) ServiceFunc {
	return func(req Request) (Response, error) {
		startT := time.Now()
		resp, err := innerFn(req)
		timeSpent := time.Since(startT)
		mItems := make([]g_met.MetricItem, 0, 3)

		if dec.config.errorClassifier != nil && err != nil {
			typeOfErr, needsToRecord := dec.config.errorClassifier(err)
			if needsToRecord {
				mItems = append(mItems, g_met.Metric(OccurredError, typeOfErr))
			}
		}
		if dec.config.needsRecordingTimeSpent {
			mItems = append(mItems, g_met.Metric(TimeSpent, timeSpent))
		}
		if len(mItems) > 0 {
			dec.config.metricSender.Send(mItems...)
		}
		return resp, err
	}
}
