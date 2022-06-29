Vim�UnDo� ��G��9�2��Xy�Zth�cf�{��g�{�P9                                     b���     _�                     *        ����                                                                                                                                                                                                                                                                                                                                                             b�T     �               �   package event       import (   	"fmt"   	"time"       	"github.com/google/uuid"   	"github.com/pkg/errors"   <	envelope "github.com/utilitywarehouse/event-envelope-proto"   #	"google.golang.org/protobuf/proto"   3	"google.golang.org/protobuf/reflect/protoregistry"   /	"google.golang.org/protobuf/types/known/anypb"   5	"google.golang.org/protobuf/types/known/timestamppb"       )	"github.com/utilitywarehouse/cx/go/meta"   )       Y// Observer is called whenever something happens to an event. E.g. it's consumed, sank or   // filtered.   #type Observer func(events ...Event)       D// Unmarshaler converts protobuf encoded bytes into the Event struct   1type Unmarshaler func(data []byte) (Event, error)       type (   	Event struct {   		ID        string   		Payload   proto.Message   		Timestamp time.Time   		AppliesAt time.Time   		Comment   string       		identifier        string   		requireIdentifier bool       		Sender   	}       	Sender struct {   		Domain      string   		Application string   	}   )       type EventOption func(*Event)       8// WithAppliesAt sets the applies_at field of the Event.   V// Use this to change the default `time.Now()` when constructing and Event with `New`.   5func WithAppliesAt(appliesAt time.Time) EventOption {   	return func(e *Event) {   		e.AppliesAt = appliesAt   	}   }       *// WithIdentifier sets the identifier when   I// [protoc-gen-uwentity](github.com/utilitywarehouse/protoc-gen-uwentity)   &// isn't defined on the proto message.   4func WithIdentifier(identifier string) EventOption {   	return func(e *Event) {   		e.identifier = identifier   	}   }       B// WithoutRequiringIdentifier removes the check for an identifier.   j// Prefer using [protoc-gen-uwentity](github.com/utilitywarehouse/protoc-gen-uwentity) or setting one with   // `WithIdentifier` option.   /func WithoutRequiringIdentifier() EventOption {   	return func(e *Event) {   		e.requireIdentifier = false   	}   }       type identifiable interface {   	GetEntityIdentifier() string   }       ;func New(msg proto.Message, options ...EventOption) Event {   	e := Event{   )		ID:                uuid.New().String(),    		Timestamp:         time.Now(),    		AppliesAt:         time.Now(),   		Payload:           msg,   		requireIdentifier: true,   	}       	if meta.Owner != "" {   		e.Domain = meta.Owner   	}        	if meta.ApplicationName != "" {   "		e.Comment = meta.ApplicationName   &		e.Application = meta.ApplicationName   	}       /	if ident, ok := e.Payload.(identifiable); ok {   ,		e.identifier = ident.GetEntityIdentifier()   	}       	for _, o := range options {   		o(&e)   	}       		return e   }       ,func Unmarshal(data []byte) (Event, error) {   	var env envelope.Event   #	err := proto.Unmarshal(data, &env)   	if err != nil {   ;		return Event{}, fmt.Errorf("envelope unmarshal: %w", err)   	}       3	if err := env.Timestamp.CheckValid(); err != nil {   =		return Event{}, fmt.Errorf("timestamp is invalid: %w", err)   	}   	ts := env.Timestamp.AsTime()       3	if err := env.AppliesAt.CheckValid(); err != nil {   >		return Event{}, fmt.Errorf("applies_at is invalid: %w", err)   	}   	at := env.AppliesAt.AsTime()       +	payload, err := env.Payload.UnmarshalNew()   	if err != nil {   -		if errors.Is(err, protoregistry.NotFound) {   B			return Event{}, fmt.Errorf("proto payload not linked: %w", err)   		}   :		return Event{}, fmt.Errorf("payload unmarshal: %w", err)   	}       	var (   		application string   		domain      string   	)   )	if sender := env.Sender; sender != nil {   "		application = sender.Application   		domain = sender.Domain   	}       	var identifier string   -	if ident, ok := payload.(identifiable); ok {   *		identifier = ident.GetEntityIdentifier()   	}       	return Event{   		ID:         env.Id,   		Timestamp:  ts,   		AppliesAt:  at,   		Payload:    payload,   		Comment:    env.Comment,   ?		Sender:     Sender{Domain: domain, Application: application},   		identifier: identifier,   	}, nil   }       !func (e Event) Validate() error {   	if e.ID == "" {   +		return errors.New("event ID is required")   	}   ,	if _, err := uuid.Parse(e.ID); err != nil {   <		return fmt.Errorf("event id is not a valid uuid: %w", err)   	}   	if e.Payload == nil {   0		return errors.New("event payload is required")   	}   	if e.Timestamp.IsZero() {   2		return errors.New("event timestamp is required")   	}   	if e.AppliesAt.IsZero() {   2		return errors.New("event appliesAt is required")   	}       /	if e.identifier == "" && e.requireIdentifier {   3		return errors.New("an identifier is required. " +   \			"consider using github.com/utilitywarehouse/protoc-gen-uwentity in your proto message " +   >			"or add one manually by using event.WithIdentifier option",   		)   	}       	return nil   }       *func (e Event) Marshal() ([]byte, error) {   %	if err := e.Validate(); err != nil {   		return nil, err   	}       !	any, err := anypb.New(e.Payload)   	if err != nil {   		return nil, err   	}       #	ts := timestamppb.New(e.Timestamp)   (	if err := ts.CheckValid(); err != nil {   ?		return nil, fmt.Errorf("marshal: invalid timestamp: %w", err)   	}       #	at := timestamppb.New(e.AppliesAt)   (	if err := at.CheckValid(); err != nil {   J		return nil, fmt.Errorf("marshal: invalid applies at timestamp: %w", err)   	}       	env := &envelope.Event{   		Id:        e.ID,   		Payload:   any,   		Timestamp: ts,   		AppliesAt: at,   		Comment:   e.Comment,   		Sender: &envelope.Sender{   			Domain:      e.Domain,   			Application: e.Application,   		},   	}       	return proto.Marshal(env)   }       &func (e Event) PartitionKey() string {   	return e.identifier   }       @// SetIdentifier sets the identifier for an Event and returns it   8func SetIdentifier(evt Event, identifier string) Event {   	evt.identifier = identifier   	return evt   }5�5�_�                             ����                                                                                                                                                                                                                                                                                                                                                             b���     �                 package event       import (   	"fmt"   	"time"       	"github.com/google/uuid"   	"github.com/pkg/errors"   <	envelope "github.com/utilitywarehouse/event-envelope-proto"   #	"google.golang.org/protobuf/proto"   3	"google.golang.org/protobuf/reflect/protoregistry"   /	"google.golang.org/protobuf/types/known/anypb"   5	"google.golang.org/protobuf/types/known/timestamppb"       )	"github.com/utilitywarehouse/cx/go/meta"   )       Y// Observer is called whenever something happens to an event. E.g. it's consumed, sank or   // filtered.   #type Observer func(events ...Event)       D// Unmarshaler converts protobuf encoded bytes into the Event struct   1type Unmarshaler func(data []byte) (Event, error)       type (   	Event struct {   		ID        string   		Payload   proto.Message   		Timestamp time.Time   		AppliesAt time.Time   		Comment   string       		identifier        string   		requireIdentifier bool       		Sender   	}       	Sender struct {   		Domain      string   		Application string   		User        *SenderUser   	}       	SenderUser struct {   		ID        string   		Reference string   		Type      string   	}   )       type EventOption func(*Event)       8// WithAppliesAt sets the applies_at field of the Event.   V// Use this to change the default `time.Now()` when constructing and Event with `New`.   5func WithAppliesAt(appliesAt time.Time) EventOption {   	return func(e *Event) {   		e.AppliesAt = appliesAt   	}   }       *// WithIdentifier sets the identifier when   I// [protoc-gen-uwentity](github.com/utilitywarehouse/protoc-gen-uwentity)   &// isn't defined on the proto message.   4func WithIdentifier(identifier string) EventOption {   	return func(e *Event) {   		e.identifier = identifier   	}   }       B// WithoutRequiringIdentifier removes the check for an identifier.   j// Prefer using [protoc-gen-uwentity](github.com/utilitywarehouse/protoc-gen-uwentity) or setting one with   // `WithIdentifier` option.   /func WithoutRequiringIdentifier() EventOption {   	return func(e *Event) {   		e.requireIdentifier = false   	}   }       1// WithSender sets the sender field of the event.   Kfunc WithSender(application, domain string, user *SenderUser) EventOption {   	return func(e *Event) {   		e.Sender = Sender{   			Domain:      domain,   			Application: application,   			User:        user,   		}   	}   }       type identifiable interface {   	GetEntityIdentifier() string   }       ;func New(msg proto.Message, options ...EventOption) Event {   	e := Event{   )		ID:                uuid.New().String(),    		Timestamp:         time.Now(),    		AppliesAt:         time.Now(),   		Payload:           msg,   		requireIdentifier: true,   	}       	if meta.Owner != "" {   		e.Domain = meta.Owner   	}        	if meta.ApplicationName != "" {   "		e.Comment = meta.ApplicationName   &		e.Application = meta.ApplicationName   	}       /	if ident, ok := e.Payload.(identifiable); ok {   ,		e.identifier = ident.GetEntityIdentifier()   	}       	for _, o := range options {   		o(&e)   	}       		return e   }       ,func Unmarshal(data []byte) (Event, error) {   	var env envelope.Event   #	err := proto.Unmarshal(data, &env)   	if err != nil {   ;		return Event{}, fmt.Errorf("envelope unmarshal: %w", err)   	}       3	if err := env.Timestamp.CheckValid(); err != nil {   =		return Event{}, fmt.Errorf("timestamp is invalid: %w", err)   	}   	ts := env.Timestamp.AsTime()       3	if err := env.AppliesAt.CheckValid(); err != nil {   >		return Event{}, fmt.Errorf("applies_at is invalid: %w", err)   	}   	at := env.AppliesAt.AsTime()       +	payload, err := env.Payload.UnmarshalNew()   	if err != nil {   -		if errors.Is(err, protoregistry.NotFound) {   B			return Event{}, fmt.Errorf("proto payload not linked: %w", err)   		}   :		return Event{}, fmt.Errorf("payload unmarshal: %w", err)   	}       	var (   		es = Sender{}   	)   )	if sender := env.Sender; sender != nil {   %		es.Application = sender.Application   		es.Domain = sender.Domain       		if sender.User != nil {   			es.User = &SenderUser{   (				ID:        sender.GetUser().GetId(),   /				Reference: sender.GetUser().GetReference(),   *				Type:      sender.GetUser().GetType(),   			}   		}   	}       	var identifier string   -	if ident, ok := payload.(identifiable); ok {   *		identifier = ident.GetEntityIdentifier()   	}       	return Event{   		ID:         env.Id,   		Timestamp:  ts,   		AppliesAt:  at,   		Payload:    payload,   		Comment:    env.Comment,   		Sender:     es,   		identifier: identifier,   	}, nil   }       !func (e Event) Validate() error {   	if e.ID == "" {   +		return errors.New("event ID is required")   	}   ,	if _, err := uuid.Parse(e.ID); err != nil {   <		return fmt.Errorf("event id is not a valid uuid: %w", err)   	}   	if e.Payload == nil {   0		return errors.New("event payload is required")   	}   	if e.Timestamp.IsZero() {   2		return errors.New("event timestamp is required")   	}   	if e.AppliesAt.IsZero() {   2		return errors.New("event appliesAt is required")   	}       /	if e.identifier == "" && e.requireIdentifier {   3		return errors.New("an identifier is required. " +   \			"consider using github.com/utilitywarehouse/protoc-gen-uwentity in your proto message " +   >			"or add one manually by using event.WithIdentifier option",   		)   	}       	return nil   }       *func (e Event) Marshal() ([]byte, error) {   %	if err := e.Validate(); err != nil {   		return nil, err   	}       !	any, err := anypb.New(e.Payload)   	if err != nil {   		return nil, err   	}       #	ts := timestamppb.New(e.Timestamp)   (	if err := ts.CheckValid(); err != nil {   ?		return nil, fmt.Errorf("marshal: invalid timestamp: %w", err)   	}       #	at := timestamppb.New(e.AppliesAt)   (	if err := at.CheckValid(); err != nil {   J		return nil, fmt.Errorf("marshal: invalid applies at timestamp: %w", err)   	}       	env := &envelope.Event{   		Id:        e.ID,   		Payload:   any,   		Timestamp: ts,   		AppliesAt: at,   		Comment:   e.Comment,   		Sender: &envelope.Sender{   			Domain:      e.Domain,   			Application: e.Application,   &			User:        mapSenderUser(e.User),   		},   	}       	return proto.Marshal(env)   }       3func mapSenderUser(in *SenderUser) *envelope.User {   	if in == nil {   		return nil   	}   	return &envelope.User{   		Type:      in.Type,   		Reference: in.Reference,   		Id:        in.ID,   	}   }       &func (e Event) PartitionKey() string {   	return e.identifier   }       @// SetIdentifier sets the identifier for an Event and returns it   8func SetIdentifier(evt Event, identifier string) Event {   	evt.identifier = identifier   	return evt   }5�5��