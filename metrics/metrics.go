package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	ipPortAddress                          string
	logger                                 logging.Logger
	numAggregatedResponses                 prometheus.Counter
	numAggregatorReceivedTasks             prometheus.Counter
	numOperatorTaskResponses               prometheus.Counter
	aggregatorGasCostPaidForBatcherTotal   prometheus.Gauge
	aggregatorNumTimesPaidForBatcher       prometheus.Counter
	numBumpedGasPriceForAggregatedResponse prometheus.Counter
}

const alignedNamespace = "aligned"

func NewMetrics(ipPortAddress string, reg prometheus.Registerer, logger logging.Logger) *Metrics {
	return &Metrics{
		ipPortAddress: ipPortAddress,
		logger:        logger,
		numAggregatedResponses: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Namespace: alignedNamespace,
			Name:      "aggregated_responses",
			Help:      "Number of aggregated responses sent to the Aligned Service Manager",
		}),
		numOperatorTaskResponses: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Namespace: alignedNamespace,
			Name:      "operator_responses",
			Help:      "Number of proof verified by the operator and sent to the Aligned Service Manager",
		}),
		numAggregatorReceivedTasks: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Namespace: alignedNamespace,
			Name:      "aggregator_received_tasks",
			Help:      "Number of tasks received by the Service Manager",
		}),
		aggregatorGasCostPaidForBatcherTotal: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Namespace: alignedNamespace,
			Name:      "aggregator_gas_cost_paid_for_batcher",
			Help:      "Accumulated gas cost the aggregator paid for the batcher when the tx cost was higher than the respondToTaskFeeLimit",
		}),
		aggregatorNumTimesPaidForBatcher: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Namespace: alignedNamespace,
			Name:      "aggregator_num_times_paid_for_batcher",
			Help:      "Number of times the aggregator paid for the batcher when the tx cost was higher than the respondToTaskFeeLimit",
		}),
		numBumpedGasPriceForAggregatedResponse: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Namespace: alignedNamespace,
			Name:      "respond_to_task_gas_price_bumped",
			Help:      "Number of times gas price was bumped while sending aggregated response",
		}),
	}
}

// Start creates a http handler for reg and starts the prometheus server in a goroutine, listening at m.ipPortAddress.
// reg needs to be the prometheus registry that was passed in the NewMetrics constructor
func (m *Metrics) Start(ctx context.Context, reg prometheus.Gatherer) <-chan error {
	m.logger.Infof("Starting metrics server at port %v", m.ipPortAddress)
	errC := make(chan error, 1)

	server := http.Server{
		Addr:           m.ipPortAddress,
		Handler:        http.NewServeMux(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // This is 1MB
	}

	server.Handler.(*http.ServeMux).Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	))

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			errC <- errors.New("prometheus server failed")
		} else {
			errC <- nil
		}
	}()
	return errC
}

func (m *Metrics) IncAggregatorReceivedTasks() {
	m.numAggregatorReceivedTasks.Inc()
}

func (m *Metrics) IncAggregatedResponses() {
	m.numAggregatedResponses.Inc()
}

func (m *Metrics) IncOperatorTaskResponses() {
	m.numOperatorTaskResponses.Inc()
}

func (m *Metrics) IncAggregatorPaidForBatcher() {
	m.aggregatorNumTimesPaidForBatcher.Inc()
}

func (m *Metrics) AddAggregatorGasPaidForBatcher(value float64) {
	m.aggregatorGasCostPaidForBatcherTotal.Add(value)
}

func (m *Metrics) IncBumpedGasPriceForAggregatedResponse() {
	m.numBumpedGasPriceForAggregatedResponse.Inc()
}
