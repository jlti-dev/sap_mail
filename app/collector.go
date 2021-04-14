package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sync"
	"time"
	"net/http"
	"fmt"
	"log"
)
type MailCollector struct {
	mutexLastLoop		sync.Mutex
	mutexLastFetch		sync.Mutex
	mutexNumberAtt		sync.Mutex
	mutexNumberErrAtt	sync.Mutex
	mutexNumberMail		sync.Mutex
	mutexNumberErrMail	sync.Mutex

	namespace		string

	lastLoopDuration	map[string]time.Duration
	lastFetchDuration	map[string]time.Duration
	numberOfAttachments	map[string]int
	numberOfErrAttachments	map[string]int
	numberOfMails		map[string]int
	numberOfErrMails	map[string]int

	lastLoopDurationDesc	*prometheus.Desc
	lastFetchDurationDesc	*prometheus.Desc
	numberOfAttDesc		*prometheus.Desc
	numberOfErrAttDesc	*prometheus.Desc
	numberOfMailDesc	*prometheus.Desc
	numberOfErrMailDesc	*prometheus.Desc

}
func NewMailCollector() *MailCollector{
	ns := "Mail_"


	ret := &MailCollector{
		mutexLastLoop: sync.Mutex{},
		mutexLastFetch: sync.Mutex{},
		mutexNumberAtt: sync.Mutex{},
		mutexNumberErrAtt: sync.Mutex{},
		mutexNumberMail: sync.Mutex{},
		mutexNumberErrMail: sync.Mutex{},

		namespace: ns,

		lastLoopDuration: make(map[string]time.Duration),
		lastFetchDuration: make(map[string]time.Duration),
		numberOfAttachments: make(map[string]int),
		numberOfErrAttachments: make(map[string]int),
		numberOfMails: make(map[string]int),
		numberOfErrMails: make(map[string]int),

		lastLoopDurationDesc: prometheus.NewDesc(
			ns + "last_loop_duration_nanoseconds",
			"Loop Duration in Nanoseconds",
			[]string{"SystemMandant"}, nil,
		),
		lastFetchDurationDesc: prometheus.NewDesc(
			ns + "last_fetch_duration_nanoseconds",
			"Time used for last fetch command (without attachments) in Nanoseconds",
			[]string{"SystemMandant"}, nil,
		),
		numberOfAttDesc: prometheus.NewDesc(
			ns + "number_of_attachments_loaded",
			"Number Of Attachments successfully loaded since last scrape",
			[]string{"SystemMandant"}, nil,
		),
		numberOfErrAttDesc: prometheus.NewDesc(
			ns + "number_of_error_attachments_loaded",
			"Number Of Attachments not successfully loaded since last scrape",
			[]string{"SystemMandant"}, nil,
		),
		numberOfMailDesc: prometheus.NewDesc(
			ns + "number_of_mails_sent",
			"Number Of mail successfully sent (including error Mails) since last scrape",
			[]string{"SystemMandant"}, nil,
		),
		numberOfErrMailDesc: prometheus.NewDesc(
			ns + "number_of_error_mails_sent",
			"Number Of mails not successfully sent since last scrape",
			[]string{"SystemMandant"}, nil,
		),
	}
	return ret
}
func (m *MailCollector) initVars(){
	m.lastLoopDuration = make(map[string]time.Duration)
	m.lastFetchDuration = make(map[string]time.Duration)
	m.numberOfAttachments = make(map[string]int)
	m.numberOfErrAttachments = make(map[string]int)
	m.numberOfMails = make(map[string]int)
	m.numberOfErrMails = make(map[string]int)
}
func (m *MailCollector) Start(port int){
	http.Handle("/metrics", promhttp.Handler())
	prometheus.MustRegister(m)
	m.initVars()
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
func (m *MailCollector) Describe(ch chan<- *prometheus.Desc){
	ch <- m.lastLoopDurationDesc
	ch <- m.lastFetchDurationDesc
	ch <- m.numberOfAttDesc
	ch <- m.numberOfErrAttDesc
	ch <- m.numberOfMailDesc
	ch <- m.numberOfErrMailDesc
}
func (m *MailCollector) Collect(ch chan<- prometheus.Metric){
	for k, v := range m.lastLoopDuration{
		ch <- prometheus.MustNewConstMetric(
			m.lastLoopDurationDesc, //Description
			prometheus.GaugeValue, //Type
			float64(v * time.Nanosecond), //value
			k,
		)
	}
	for k, v := range m.lastFetchDuration{
		ch <- prometheus.MustNewConstMetric(
			m.lastFetchDurationDesc, //Description
			prometheus.GaugeValue, //Type
			float64(v * time.Nanosecond), //value
			k,
		)
	}
	for k, v := range m.numberOfAttachments{
		ch <- prometheus.MustNewConstMetric(
			m.numberOfAttDesc, //Description
			prometheus.GaugeValue, //Type
			float64(v), //value
			k,
		)
	}
	for k, v := range m.numberOfErrAttachments {
		ch <- prometheus.MustNewConstMetric(
			m.numberOfErrAttDesc, //Description
			prometheus.GaugeValue, //Type
			float64(v), //value
			k,
		)
	}
	for k, v := range m.numberOfMails {
		ch <- prometheus.MustNewConstMetric(
			m.numberOfMailDesc, //Description
			prometheus.GaugeValue, //Type
			float64(v), //value
			k,
		)
	}
	for k, v := range m.numberOfErrMails {
		ch <- prometheus.MustNewConstMetric(
			m.numberOfErrMailDesc, //Description
			prometheus.GaugeValue, //Type
			float64(v), //value
			k,
		)
	}
	//m.initVars()
}
func (m *MailCollector) setLastLoopDuration(sys string, mandt string, loopDuration time.Duration){
	m.mutexLastLoop.Lock()
	defer m.mutexLastLoop.Unlock()

	m.lastLoopDuration[fmt.Sprintf("%s-%s", sys, mandt)] = loopDuration
}
func (m *MailCollector) setLastFetchDuration(sys string, mandt string, fetchDuration time.Duration){
	m.mutexLastFetch.Lock()
	defer m.mutexLastFetch.Unlock()

	m.lastFetchDuration[fmt.Sprintf("%s-%s", sys, mandt)] = fetchDuration
}
func (m *MailCollector) incMails(sys string, mandt string) {
	m.mutexNumberMail.Lock()
	defer m.mutexNumberMail.Unlock()

	m.numberOfMails[fmt.Sprintf("%s-%s", sys, mandt)] ++
}
func (m *MailCollector) incErrMails(sys string, mandt string,) {
	m.mutexNumberErrMail.Lock()
	defer m.mutexNumberErrMail.Unlock()

	m.numberOfErrMails[fmt.Sprintf("%s-%s", sys, mandt)] ++
}

func (m *MailCollector) incAttachments(sys string, mandt string,) {
	m.mutexNumberAtt.Lock()
	defer m.mutexNumberAtt.Unlock()

	m.numberOfAttachments[fmt.Sprintf("%s-%s", sys, mandt)] ++
}
func (m *MailCollector) incErrAttachments(sys string, mandt string,) {
	m.mutexNumberErrAtt.Lock()
	defer m.mutexNumberErrAtt.Unlock()

	m.numberOfErrAttachments[fmt.Sprintf("%s-%s", sys, mandt)] ++
}
