package metadata

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"strings"
	"sync"
)

// Allocations accepts reports from executor
type Allocations struct {
	*BaseProducer
	dataMu   *sync.Mutex
	podsData map[string]string
}

func NewAllocation(ctx context.Context, log *logx.Log) (s *Allocations) {
	s = &Allocations{
		BaseProducer: NewBaseProducer(ctx, log, "allocation"),
		dataMu:     &sync.Mutex{},
		podsData:   map[string]string{},
	}
	return
}

func (s *Allocations) Sync(pods []*allocation.Pod) {
	s.dataMu.Lock()
	defer s.dataMu.Unlock()
	s.podsData = map[string]string{}
	for _, v := range pods {
		s.update(v.Name, v, nil)
	}
	s.Store(true, s.podsData)
}

func (s *Allocations) Report(name string, pod *allocation.Pod, failures []error) {
	s.dataMu.Lock()
	defer s.dataMu.Unlock()
	s.update(name, pod, failures)
	s.Store(true, s.podsData)
}

func (s *Allocations) update(name string, pod *allocation.Pod, failures []error) {
	if pod == nil {
		for k := range s.podsData {
			if strings.HasPrefix(k, name+".") {
				delete(s.podsData, k)
			}
		}
		return
	}
	s.podsData[name+".present"] = "true"
	s.podsData[name+".namespace"] = pod.Namespace
	s.podsData[name+".failures"] = fmt.Sprintf("%v", failures)
	var units []string
	for _, unit := range pod.Units {
		units = append(units, unit.UnitName())
	}
	s.podsData[name+".units"] = strings.Join(units, ",")
}
