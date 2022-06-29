Vim�UnDo� �w�r��\���o���D�F�� 뇣]�P���  )                                   b��     _�                             ����                                                                                                                                                                                                                                                                                                                                                             b�G�     �              )   package event       import (   
	"context"   	"fmt"   	"time"       1	"github.com/prometheus/client_golang/prometheus"   (	"github.com/utilitywarehouse/cx/go/log"   0	"github.com/utilitywarehouse/cx/go/operational"   +	"github.com/utilitywarehouse/cx/go/ticker"   	"github.com/uw-labs/substrate"   	"golang.org/x/sync/errgroup"   )       const (   	defaultBatchSize    = 1000   &	defaultBatchMaxWait = 1 * time.Second   )       type (   B	HandleEventFunc      func(ctx context.Context, event Event) error   E	HandleEventBatchFunc func(ctx context.Context, events []Event) error   ,	ConsumerOption       func(c *EventConsumer)   )       -func WithBatchSize(size int) ConsumerOption {    	return func(c *EventConsumer) {   		c.batchSize = size   	}   }       :func WithBatchMaxWait(wait time.Duration) ConsumerOption {    	return func(c *EventConsumer) {   		c.batchMaxWait = wait   	}   }       ]// WithFilter appends a new Filter to the EventConsumer. Filters are invoked on each consumed   K// Event to determine if it should be passed to the HandlerFunc or ignored.   /func WithFilter(filter Filter) ConsumerOption {    	return func(c *EventConsumer) {   '		c.filters = append(c.filters, filter)   	}   }       M// WithFilterEventObserver appends a new Observer to the EventConsumer. These   L// EventObservers are invoked with each Event that is successfully filtered.   :func WithFilterEventObserver(eo Observer) ConsumerOption {    	return func(c *EventConsumer) {   -		c.filterObservers = append(c.observers, eo)   	}   }       ^// WithConsumerEventObserver appends a new Observer to the EventConsumer. These EventObservers   ?// are invoked with each message that is successfully consumed.   <func WithConsumerEventObserver(eo Observer) ConsumerOption {    	return func(c *EventConsumer) {   '		c.observers = append(c.observers, eo)   	}   }       d// WithConsumeCounter registers a new prometheus counter that is incremented every time a message is   ]// consumed and handled. The metric is labelled with the "type" of the Event payload that was   _// consumed. Note that the fully-qualified name of the metric must be a valid Prometheus metric   // name.   5func WithConsumeCounter(name string) ConsumerOption {   <	counter := prometheus.NewCounterVec(prometheus.CounterOpts{   		Namespace: "event",   		Subsystem: "consumed",   		Name:      name,   @		Help:      "counts number of events consumed, per proto type",   	}, []string{"type"})   "	operational.RemoveMetric(counter)    	operational.AddMetrics(counter)       &	incCounter := func(events ...Event) {   		for _, e := range events {   			counter.WithLabelValues(   9				string(e.Payload.ProtoReflect().Descriptor().Name()),   
			).Inc()   		}   	}       -	return WithConsumerEventObserver(incCounter)   }       Y// WithFilterCounter registers a new prometheus counter that is incremented every time an   Z// Event is consumed and filtered (not handled). Note that the fully-qualified name of the   1// metric must be a valid Prometheus metric name.   4func WithFilterCounter(name string) ConsumerOption {   <	counter := prometheus.NewCounterVec(prometheus.CounterOpts{   		Namespace: "event",   		Subsystem: "filtered",   		Name:      name,   N		Help:      "counts number of events filtered out (ignored), per proto type",   	}, []string{"type"})   "	operational.RemoveMetric(counter)    	operational.AddMetrics(counter)       &	incCounter := func(events ...Event) {   		for _, e := range events {   			counter.WithLabelValues(   9				string(e.Payload.ProtoReflect().Descriptor().Name()),   
			).Inc()   		}   	}       +	return WithFilterEventObserver(incCounter)   }       W// WithUnmarshaler overrides the Unmarshaler func to use when converting protobuf bytes   // to Event.   4func WithUnmarshaler(m Unmarshaler) ConsumerOption {    	return func(c *EventConsumer) {   		c.unmarshaler = m   	}   }       type EventConsumer struct {   %	source  substrate.AsyncMessageSource   	filters []Filter       	observers       []Observer   	filterObservers []Observer       	batchSize    int   	batchMaxWait time.Duration       	unmarshaler Unmarshaler   }       func NewEventConsumer(   %	source substrate.AsyncMessageSource,   	options ...ConsumerOption,   ) *EventConsumer {   	c := &EventConsumer{   		source:       source,   !		batchSize:    defaultBatchSize,   $		batchMaxWait: defaultBatchMaxWait,   		unmarshaler:  Unmarshal,   	}       	for _, o := range options {   		o(c)   	}       		return c   }       C// Consume start reading from a source, unmarshalling into an Event   Ufunc (c *EventConsumer) Consume(ctx context.Context, handler HandleEventFunc) error {   >	syncSource := substrate.NewSynchronousMessageSource(c.source)   _	return syncSource.ConsumeMessages(ctx, func(mCtx context.Context, m substrate.Message) error {   8		if m == nil || m.Data() == nil || len(m.Data()) == 0 {   			return nil   		}   '		event, err := c.unmarshaler(m.Data())   		if err != nil {   			return err   		}       		if c.ignoreEvent(event) {   (			for _, o := range c.filterObservers {   				o(event)   			}   			return nil   		}       		err = handler(ctx, event)   		if err != nil {   			return err   		}       !		for _, o := range c.observers {   			o(event)   		}   		return nil   	})   }       %func (c *EventConsumer) ConsumeBatch(   	ctx context.Context,   	handler HandleEventBatchFunc,   	) error {   %	eg, ctx := errgroup.WithContext(ctx)       %	acks := make(chan substrate.Message)   )	messages := make(chan substrate.Message)       	eg.Go(func() error {   <		return c.handleBatchMessages(ctx, handler, messages, acks)   	})   	eg.Go(func() error {   6		return c.source.ConsumeMessages(ctx, messages, acks)   	})       	return eg.Wait()   }       ,func (c *EventConsumer) handleBatchMessages(   	ctx context.Context,   	handler HandleEventBatchFunc,   #	messages <-chan substrate.Message,   	acks chan<- substrate.Message,   	) error {   #	tick := ticker.New(c.batchMaxWait)   	defer tick.Stop()       '	batch := make([]Event, 0, c.batchSize)   3	toAck := make([]substrate.Message, 0, c.batchSize)       	process := func() error {   		if len(batch) > 0 {   D			log.WithField("batch_size", len(batch)).Debug("processing batch")   .			if err := handler(ctx, batch); err != nil {   )				return fmt.Errorf("handler: %w", err)   			}   		}       		for _, ack := range toAck {   			select {   			case <-ctx.Done():   				return ctx.Err()   			case acks <- ack:   			}   		}       !		for _, o := range c.observers {   			o(batch...)   		}       B		log.WithField("batch_size", len(batch)).Debug("batch processed")   		batch = batch[:0]   		toAck = toAck[:0]   P		tick.Reset() // Reset the ticker to get correct deadline to receive next batch       		return nil   	}       2	putMessage := func(msg substrate.Message) error {   >		if msg == nil || msg.Data() == nil || len(msg.Data()) == 0 {   			toAck = append(toAck, msg)   !			if len(toAck) == c.batchSize {   				return process()   			}   			return nil   		}   %		e, err := c.unmarshaler(msg.Data())   		if err != nil {   *			return fmt.Errorf("unmarshal: %w", err)   		}   		ignore := c.ignoreEvent(e)   		if ignore {   (			for _, o := range c.filterObservers {   				o(e)   			}   		}   		if !ignore {   			batch = append(batch, e)   		}   9		if dMsg, ok := msg.(substrate.DiscardableMessage); ok {   2			// Discard payload when possible to save memory   			dMsg.DiscardPayload()   		}   		toAck = append(toAck, msg)        		if len(toAck) == c.batchSize {   			return process()   		}   		return nil   	}       	for {   
		select {   		case <-ctx.Done():   			return ctx.Err()   		case msg := <-messages:   *			if err := putMessage(msg); err != nil {   				return err   			}   		case <-tick.C:   $			if err := process(); err != nil {   				return err   			}   		}   	}   }       7func (c *EventConsumer) ignoreEvent(event Event) bool {   #	for _, filter := range c.filters {   		if !filter(event.Payload) {   			return true   		}   	}   	return false   }5�5�_�                            ����                                                                                                                                                                                                                                                                                                                                                             b�E�     �              )   package event       import (   
	"context"   	"fmt"   	"time"       1	"github.com/prometheus/client_golang/prometheus"   (	"github.com/utilitywarehouse/cx/go/log"   0	"github.com/utilitywarehouse/cx/go/operational"   +	"github.com/utilitywarehouse/cx/go/ticker"   	"github.com/uw-labs/substrate"   	"golang.org/x/sync/errgroup"   )       const (   	defaultBatchSize    = 1000   &	defaultBatchMaxWait = 1 * time.Second   )       type (   B	HandleEventFunc      func(ctx context.Context, event Event) error   E	HandleEventBatchFunc func(ctx context.Context, events []Event) error   ,	ConsumerOption       func(c *EventConsumer)   )       -func WithBatchSize(size int) ConsumerOption {    	return func(c *EventConsumer) {   		c.batchSize = size   	}   }       :func WithBatchMaxWait(wait time.Duration) ConsumerOption {    	return func(c *EventConsumer) {   		c.batchMaxWait = wait   	}   }       ]// WithFilter appends a new Filter to the EventConsumer. Filters are invoked on each consumed   K// Event to determine if it should be passed to the HandlerFunc or ignored.   /func WithFilter(filter Filter) ConsumerOption {    	return func(c *EventConsumer) {   '		c.filters = append(c.filters, filter)   	}   }       M// WithFilterEventObserver appends a new Observer to the EventConsumer. These   L// EventObservers are invoked with each Event that is successfully filtered.   :func WithFilterEventObserver(eo Observer) ConsumerOption {    	return func(c *EventConsumer) {   -		c.filterObservers = append(c.observers, eo)   	}   }       ^// WithConsumerEventObserver appends a new Observer to the EventConsumer. These EventObservers   ?// are invoked with each message that is successfully consumed.   <func WithConsumerEventObserver(eo Observer) ConsumerOption {    	return func(c *EventConsumer) {   '		c.observers = append(c.observers, eo)   	}   }       d// WithConsumeCounter registers a new prometheus counter that is incremented every time a message is   ]// consumed and handled. The metric is labelled with the "type" of the Event payload that was   _// consumed. Note that the fully-qualified name of the metric must be a valid Prometheus metric   // name.   5func WithConsumeCounter(name string) ConsumerOption {   <	counter := prometheus.NewCounterVec(prometheus.CounterOpts{   		Namespace: "event",   		Subsystem: "consumed",   		Name:      name,   @		Help:      "counts number of events consumed, per proto type",   	}, []string{"type"})   "	operational.RemoveMetric(counter)    	operational.AddMetrics(counter)       &	incCounter := func(events ...Event) {   		for _, e := range events {   			counter.WithLabelValues(   9				string(e.Payload.ProtoReflect().Descriptor().Name()),   
			).Inc()   		}   	}       -	return WithConsumerEventObserver(incCounter)   }       Y// WithFilterCounter registers a new prometheus counter that is incremented every time an   Z// Event is consumed and filtered (not handled). Note that the fully-qualified name of the   1// metric must be a valid Prometheus metric name.   4func WithFilterCounter(name string) ConsumerOption {   <	counter := prometheus.NewCounterVec(prometheus.CounterOpts{   		Namespace: "event",   		Subsystem: "filtered",   		Name:      name,   N		Help:      "counts number of events filtered out (ignored), per proto type",   	}, []string{"type"})   "	operational.RemoveMetric(counter)    	operational.AddMetrics(counter)       &	incCounter := func(events ...Event) {   		for _, e := range events {   			counter.WithLabelValues(   9				string(e.Payload.ProtoReflect().Descriptor().Name()),   
			).Inc()   		}   	}       +	return WithFilterEventObserver(incCounter)   }       W// WithUnmarshaler overrides the Unmarshaler func to use when converting protobuf bytes   // to Event.   4func WithUnmarshaler(m Unmarshaler) ConsumerOption {    	return func(c *EventConsumer) {   		c.unmarshaler = m   	}   }       type EventConsumer struct {   %	source  substrate.AsyncMessageSource   	filters []Filter       	observers       []Observer   	filterObservers []Observer       	batchSize    int   	batchMaxWait time.Duration       	unmarshaler Unmarshaler   }       func NewEventConsumer(   %	source substrate.AsyncMessageSource,   	options ...ConsumerOption,   ) *EventConsumer {   	c := &EventConsumer{   		source:       source,   !		batchSize:    defaultBatchSize,   $		batchMaxWait: defaultBatchMaxWait,   		unmarshaler:  Unmarshal,   	}       	for _, o := range options {   		o(c)   	}       		return c   }       C// Consume start reading from a source, unmarshalling into an Event   Ufunc (c *EventConsumer) Consume(ctx context.Context, handler HandleEventFunc) error {   >	syncSource := substrate.NewSynchronousMessageSource(c.source)   _	return syncSource.ConsumeMessages(ctx, func(mCtx context.Context, m substrate.Message) error {   8		if m == nil || m.Data() == nil || len(m.Data()) == 0 {   			return nil   		}   '		event, err := c.unmarshaler(m.Data())   		if err != nil {   			return err   		}       		if c.ignoreEvent(event) {   (			for _, o := range c.filterObservers {   				o(event)   			}   			return nil   		}       		err = handler(ctx, event)   		if err != nil {   			return err   		}       !		for _, o := range c.observers {   			o(event)   		}   		return nil   	})   }       %func (c *EventConsumer) ConsumeBatch(   	ctx context.Context,   	handler HandleEventBatchFunc,   	) error {   %	eg, ctx := errgroup.WithContext(ctx)       %	acks := make(chan substrate.Message)   )	messages := make(chan substrate.Message)       	eg.Go(func() error {   <		return c.handleBatchMessages(ctx, handler, messages, acks)   	})   	eg.Go(func() error {   6		return c.source.ConsumeMessages(ctx, messages, acks)   	})       	return eg.Wait()   }       ,func (c *EventConsumer) handleBatchMessages(   	ctx context.Context,   	handler HandleEventBatchFunc,   #	messages <-chan substrate.Message,   	acks chan<- substrate.Message,   	) error {   #	tick := ticker.New(c.batchMaxWait)   	defer tick.Stop()       '	batch := make([]Event, 0, c.batchSize)   3	toAck := make([]substrate.Message, 0, c.batchSize)       	process := func() error {   		if len(batch) > 0 {   D			log.WithField("batch_size", len(batch)).Debug("processing batch")   .			if err := handler(ctx, batch); err != nil {   )				return fmt.Errorf("handler: %w", err)   			}   		}       		for _, ack := range toAck {   			select {   			case <-ctx.Done():   				return ctx.Err()   			case acks <- ack:   			}   		}       !		for _, o := range c.observers {   			o(batch...)   		}       B		log.WithField("batch_size", len(batch)).Debug("batch processed")   		batch = batch[:0]   		toAck = toAck[:0]   P		tick.Reset() // Reset the ticker to get correct deadline to receive next batch       		return nil   	}       2	putMessage := func(msg substrate.Message) error {   >		if msg == nil || msg.Data() == nil || len(msg.Data()) == 0 {   			toAck = append(toAck, msg)   !			if len(toAck) == c.batchSize {   				return process()   			}   			return nil   		}   %		e, err := c.unmarshaler(msg.Data())   		if err != nil {   *			return fmt.Errorf("unmarshal: %w", err)   		}   		ignore := c.ignoreEvent(e)   		if ignore {   (			for _, o := range c.filterObservers {   				o(e)   			}   		}   		if !ignore {   			batch = append(batch, e)   		}   9		if dMsg, ok := msg.(substrate.DiscardableMessage); ok {   2			// Discard payload when possible to save memory   			dMsg.DiscardPayload()   		}   		toAck = append(toAck, msg)        		if len(toAck) == c.batchSize {   			return process()   		}   		return nil   	}       	for {   
		select {   		case <-ctx.Done():   			return ctx.Err()   		case msg := <-messages:   *			if err := putMessage(msg); err != nil {   				return err   			}   		case <-tick.C:   $			if err := process(); err != nil {   				return err   			}   		}   	}   }       7func (c *EventConsumer) ignoreEvent(event Event) bool {   #	for _, filter := range c.filters {   		if !filter(event.Payload) {   			return true   		}   	}   	return false   }5�5�_�                             ����                                                                                                                                                                                                                                                                                                                                                             b��     �              )   package event       import (   
	"context"   	"fmt"   	"time"       1	"github.com/prometheus/client_golang/prometheus"   (	"github.com/utilitywarehouse/cx/go/log"   0	"github.com/utilitywarehouse/cx/go/operational"   +	"github.com/utilitywarehouse/cx/go/ticker"   	"github.com/uw-labs/substrate"   	"golang.org/x/sync/errgroup"   )       const (   	defaultBatchSize    = 1000   &	defaultBatchMaxWait = 1 * time.Second   )       type (   B	HandleEventFunc      func(ctx context.Context, event Event) error   E	HandleEventBatchFunc func(ctx context.Context, events []Event) error   ,	ConsumerOption       func(c *EventConsumer)   )       -func WithBatchSize(size int) ConsumerOption {    	return func(c *EventConsumer) {   		c.batchSize = size   	}   }       :func WithBatchMaxWait(wait time.Duration) ConsumerOption {    	return func(c *EventConsumer) {   		c.batchMaxWait = wait   	}   }       ]// WithFilter appends a new Filter to the EventConsumer. Filters are invoked on each consumed   K// Event to determine if it should be passed to the HandlerFunc or ignored.   /func WithFilter(filter Filter) ConsumerOption {    	return func(c *EventConsumer) {   '		c.filters = append(c.filters, filter)   	}   }       M// WithFilterEventObserver appends a new Observer to the EventConsumer. These   L// EventObservers are invoked with each Event that is successfully filtered.   :func WithFilterEventObserver(eo Observer) ConsumerOption {    	return func(c *EventConsumer) {   -		c.filterObservers = append(c.observers, eo)   	}   }       ^// WithConsumerEventObserver appends a new Observer to the EventConsumer. These EventObservers   ?// are invoked with each message that is successfully consumed.   <func WithConsumerEventObserver(eo Observer) ConsumerOption {    	return func(c *EventConsumer) {   '		c.observers = append(c.observers, eo)   	}   }       d// WithConsumeCounter registers a new prometheus counter that is incremented every time a message is   ]// consumed and handled. The metric is labelled with the "type" of the Event payload that was   _// consumed. Note that the fully-qualified name of the metric must be a valid Prometheus metric   // name.   5func WithConsumeCounter(name string) ConsumerOption {   <	counter := prometheus.NewCounterVec(prometheus.CounterOpts{   		Namespace: "event",   		Subsystem: "consumed",   		Name:      name,   @		Help:      "counts number of events consumed, per proto type",   	}, []string{"type"})   "	operational.RemoveMetric(counter)    	operational.AddMetrics(counter)       &	incCounter := func(events ...Event) {   		for _, e := range events {   			counter.WithLabelValues(   9				string(e.Payload.ProtoReflect().Descriptor().Name()),   
			).Inc()   		}   	}       -	return WithConsumerEventObserver(incCounter)   }       Y// WithFilterCounter registers a new prometheus counter that is incremented every time an   Z// Event is consumed and filtered (not handled). Note that the fully-qualified name of the   1// metric must be a valid Prometheus metric name.   4func WithFilterCounter(name string) ConsumerOption {   <	counter := prometheus.NewCounterVec(prometheus.CounterOpts{   		Namespace: "event",   		Subsystem: "filtered",   		Name:      name,   N		Help:      "counts number of events filtered out (ignored), per proto type",   	}, []string{"type"})   "	operational.RemoveMetric(counter)    	operational.AddMetrics(counter)       &	incCounter := func(events ...Event) {   		for _, e := range events {   			counter.WithLabelValues(   9				string(e.Payload.ProtoReflect().Descriptor().Name()),   
			).Inc()   		}   	}       +	return WithFilterEventObserver(incCounter)   }       W// WithUnmarshaler overrides the Unmarshaler func to use when converting protobuf bytes   // to Event.   4func WithUnmarshaler(m Unmarshaler) ConsumerOption {    	return func(c *EventConsumer) {   		c.unmarshaler = m   	}   }       type EventConsumer struct {   %	source  substrate.AsyncMessageSource   	filters []Filter       	observers       []Observer   	filterObservers []Observer       	batchSize    int   	batchMaxWait time.Duration       	unmarshaler Unmarshaler   }       func NewEventConsumer(   %	source substrate.AsyncMessageSource,   	options ...ConsumerOption,   ) *EventConsumer {   	c := &EventConsumer{   		source:       source,   !		batchSize:    defaultBatchSize,   $		batchMaxWait: defaultBatchMaxWait,   		unmarshaler:  Unmarshal,   	}       	for _, o := range options {   		o(c)   	}       		return c   }       C// Consume start reading from a source, unmarshalling into an Event   Ufunc (c *EventConsumer) Consume(ctx context.Context, handler HandleEventFunc) error {   >	syncSource := substrate.NewSynchronousMessageSource(c.source)   _	return syncSource.ConsumeMessages(ctx, func(mCtx context.Context, m substrate.Message) error {   8		if m == nil || m.Data() == nil || len(m.Data()) == 0 {   			return nil   		}   '		event, err := c.unmarshaler(m.Data())   		if err != nil {   			return err   		}       		if c.ignoreEvent(event) {   (			for _, o := range c.filterObservers {   				o(event)   			}   			return nil   		}       		err = handler(ctx, event)   		if err != nil {   			return err   		}       !		for _, o := range c.observers {   			o(event)   		}   		return nil   	})   }       %func (c *EventConsumer) ConsumeBatch(   	ctx context.Context,   	handler HandleEventBatchFunc,   	) error {   %	eg, ctx := errgroup.WithContext(ctx)       %	acks := make(chan substrate.Message)   )	messages := make(chan substrate.Message)       	eg.Go(func() error {   <		return c.handleBatchMessages(ctx, handler, messages, acks)   	})   	eg.Go(func() error {   6		return c.source.ConsumeMessages(ctx, messages, acks)   	})       	return eg.Wait()   }       ,func (c *EventConsumer) handleBatchMessages(   	ctx context.Context,   	handler HandleEventBatchFunc,   #	messages <-chan substrate.Message,   	acks chan<- substrate.Message,   	) error {   #	tick := ticker.New(c.batchMaxWait)   	defer tick.Stop()       '	batch := make([]Event, 0, c.batchSize)   3	toAck := make([]substrate.Message, 0, c.batchSize)       	process := func() error {   		if len(batch) > 0 {   D			log.WithField("batch_size", len(batch)).Debug("processing batch")   .			if err := handler(ctx, batch); err != nil {   )				return fmt.Errorf("handler: %w", err)   			}   		}       		for _, ack := range toAck {   			select {   			case <-ctx.Done():   				return ctx.Err()   			case acks <- ack:   			}   		}       !		for _, o := range c.observers {   			o(batch...)   		}       B		log.WithField("batch_size", len(batch)).Debug("batch processed")   		batch = batch[:0]   		toAck = toAck[:0]   P		tick.Reset() // Reset the ticker to get correct deadline to receive next batch       		return nil   	}       2	putMessage := func(msg substrate.Message) error {   >		if msg == nil || msg.Data() == nil || len(msg.Data()) == 0 {   			toAck = append(toAck, msg)   !			if len(toAck) == c.batchSize {   				return process()   			}   			return nil   		}   %		e, err := c.unmarshaler(msg.Data())   		if err != nil {   *			return fmt.Errorf("unmarshal: %w", err)   		}   		ignore := c.ignoreEvent(e)   		if ignore {   (			for _, o := range c.filterObservers {   				o(e)   			}   		}   		if !ignore {   			batch = append(batch, e)   		}   9		if dMsg, ok := msg.(substrate.DiscardableMessage); ok {   2			// Discard payload when possible to save memory   			dMsg.DiscardPayload()   		}   		toAck = append(toAck, msg)        		if len(toAck) == c.batchSize {   			return process()   		}   		return nil   	}       	for {   
		select {   		case <-ctx.Done():   			return ctx.Err()   		case msg := <-messages:   *			if err := putMessage(msg); err != nil {   				return err   			}   		case <-tick.C:   $			if err := process(); err != nil {   				return err   			}   		}   	}   }       7func (c *EventConsumer) ignoreEvent(event Event) bool {   #	for _, filter := range c.filters {   		if !filter(event.Payload) {   			return true   		}   	}   	return false   }5�5��