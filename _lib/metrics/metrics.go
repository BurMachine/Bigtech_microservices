package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics содержит все метрики для сервиса
type Metrics struct {
	// gRPC Server метрики
	GRPCRequestsTotal    *prometheus.CounterVec
	GRPCRequestDuration  *prometheus.HistogramVec
	GRPCRequestsInFlight *prometheus.GaugeVec

	// HTTP Server метрики
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight *prometheus.GaugeVec

	// gRPC Client метрики
	GRPCClientRequestsTotal   *prometheus.CounterVec
	GRPCClientRequestDuration *prometheus.HistogramVec

	// Database метрики
	DBQueryDuration   *prometheus.HistogramVec
	DBQueryTotal      *prometheus.CounterVec
	DBConnectionsOpen *prometheus.GaugeVec
	DBConnectionsIdle *prometheus.GaugeVec
	DBConnectionsUsed *prometheus.GaugeVec

	// Kafka Consumer метрики
	KafkaMessagesConsumed *prometheus.CounterVec
	KafkaMessagesErrors   *prometheus.CounterVec
	KafkaConsumerLag      *prometheus.GaugeVec
	KafkaBatchDuration    *prometheus.HistogramVec

	// Worker метрики (Outbox/Inbox)
	WorkerBatchProcessed *prometheus.CounterVec
	WorkerBatchErrors    *prometheus.CounterVec
	WorkerBatchDuration  *prometheus.HistogramVec
	WorkerQueueSize      *prometheus.GaugeVec
}

// Config конфигурация метрик
type Config struct {
	ServiceName string
	Namespace   string // Например "messenger"
	Subsystem   string // Например "gateway", "auth", "users"
}

// New создает новый набор метрик для сервиса
func New(cfg Config) *Metrics {
	if cfg.Namespace == "" {
		cfg.Namespace = "messenger"
	}
	if cfg.Subsystem == "" {
		cfg.Subsystem = cfg.ServiceName
	}

	return &Metrics{
		// ========== gRPC Server ==========
		GRPCRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "grpc_requests_total",
				Help:      "Total number of gRPC requests",
			},
			[]string{"service", "method", "code"},
		),

		GRPCRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "grpc_request_duration_seconds",
				Help:      "gRPC request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"service", "method"},
		),

		GRPCRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "grpc_requests_in_flight",
				Help:      "Number of gRPC requests currently being processed",
			},
			[]string{"service", "method"},
		),

		// ========== HTTP Server ==========
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"service", "method", "path", "code"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"service", "method", "path"},
		),

		HTTPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "http_requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
			[]string{"service", "method", "path"},
		),

		// ========== gRPC Client ==========
		GRPCClientRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "grpc_client_requests_total",
				Help:      "Total number of gRPC client requests",
			},
			[]string{"target_service", "method", "code"},
		),

		GRPCClientRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "grpc_client_request_duration_seconds",
				Help:      "gRPC client request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"target_service", "method"},
		),

		// ========== Database ==========
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.0001, .0005, .001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation", "table"},
		),

		DBQueryTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		DBConnectionsOpen: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "db_connections_open",
				Help:      "Number of open database connections",
			},
			[]string{"database"},
		),

		DBConnectionsIdle: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "db_connections_idle",
				Help:      "Number of idle database connections",
			},
			[]string{"database"},
		),

		DBConnectionsUsed: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "db_connections_in_use",
				Help:      "Number of database connections in use",
			},
			[]string{"database"},
		),

		// ========== Kafka Consumer ==========
		KafkaMessagesConsumed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "kafka_messages_consumed_total",
				Help:      "Total number of Kafka messages consumed",
			},
			[]string{"topic", "consumer_group"},
		),

		KafkaMessagesErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "kafka_messages_errors_total",
				Help:      "Total number of Kafka message processing errors",
			},
			[]string{"topic", "consumer_group", "error_type"},
		),

		KafkaConsumerLag: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "kafka_consumer_lag",
				Help:      "Kafka consumer lag (messages behind)",
			},
			[]string{"topic", "partition", "consumer_group"},
		),

		KafkaBatchDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "kafka_batch_duration_seconds",
				Help:      "Kafka batch processing duration in seconds",
				Buckets:   []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
			},
			[]string{"topic", "consumer_group"},
		),

		// ========== Worker (Outbox/Inbox) ==========
		WorkerBatchProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "worker_batch_processed_total",
				Help:      "Total number of worker batches processed",
			},
			[]string{"worker_type", "status"}, // worker_type: "outbox" или "inbox"
		),

		WorkerBatchErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "worker_batch_errors_total",
				Help:      "Total number of worker batch errors",
			},
			[]string{"worker_type", "error_type"},
		),

		WorkerBatchDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "worker_batch_duration_seconds",
				Help:      "Worker batch processing duration in seconds",
				Buckets:   []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
			},
			[]string{"worker_type"},
		),

		WorkerQueueSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "worker_queue_size",
				Help:      "Current size of worker queue",
			},
			[]string{"worker_type"},
		),
	}
}

// RecordGRPCRequest записывает метрики для gRPC запроса
func (m *Metrics) RecordGRPCRequest(service, method, code string, duration time.Duration) {
	m.GRPCRequestsTotal.WithLabelValues(service, method, code).Inc()
	m.GRPCRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// RecordHTTPRequest записывает метрики для HTTP запроса
func (m *Metrics) RecordHTTPRequest(service, method, path, code string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(service, method, path, code).Inc()
	m.HTTPRequestDuration.WithLabelValues(service, method, path).Observe(duration.Seconds())
}

// RecordGRPCClientRequest записывает метрики для gRPC client запроса
func (m *Metrics) RecordGRPCClientRequest(targetService, method, code string, duration time.Duration) {
	m.GRPCClientRequestsTotal.WithLabelValues(targetService, method, code).Inc()
	m.GRPCClientRequestDuration.WithLabelValues(targetService, method).Observe(duration.Seconds())
}

// RecordDBQuery записывает метрики для БД запроса
func (m *Metrics) RecordDBQuery(operation, table, status string, duration time.Duration) {
	m.DBQueryTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdateDBStats обновляет метрики пула соединений БД
func (m *Metrics) UpdateDBStats(database string, open, idle, inUse int) {
	m.DBConnectionsOpen.WithLabelValues(database).Set(float64(open))
	m.DBConnectionsIdle.WithLabelValues(database).Set(float64(idle))
	m.DBConnectionsUsed.WithLabelValues(database).Set(float64(inUse))
}

// RecordKafkaMessage записывает метрики для Kafka сообщения
func (m *Metrics) RecordKafkaMessage(topic, consumerGroup string, success bool, duration time.Duration) {
	if success {
		m.KafkaMessagesConsumed.WithLabelValues(topic, consumerGroup).Inc()
	} else {
		m.KafkaMessagesErrors.WithLabelValues(topic, consumerGroup, "processing_error").Inc()
	}
	m.KafkaBatchDuration.WithLabelValues(topic, consumerGroup).Observe(duration.Seconds())
}

// UpdateKafkaLag обновляет метрику lag для Kafka consumer
func (m *Metrics) UpdateKafkaLag(topic, partition, consumerGroup string, lag int64) {
	m.KafkaConsumerLag.WithLabelValues(topic, partition, consumerGroup).Set(float64(lag))
}

// RecordWorkerBatch записывает метрики для worker batch
func (m *Metrics) RecordWorkerBatch(workerType, status string, duration time.Duration) {
	m.WorkerBatchProcessed.WithLabelValues(workerType, status).Inc()
	m.WorkerBatchDuration.WithLabelValues(workerType).Observe(duration.Seconds())
}

// UpdateWorkerQueueSize обновляет размер очереди worker
func (m *Metrics) UpdateWorkerQueueSize(workerType string, size int) {
	m.WorkerQueueSize.WithLabelValues(workerType).Set(float64(size))
}
