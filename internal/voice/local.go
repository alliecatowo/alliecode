package voice

import (
	"sort"
	"sync"
	"time"
)

type Capability string

const (
	CapabilityWakeWord   Capability = "wake_word"
	CapabilityDictation  Capability = "dictation"
	CapabilityPushToTalk Capability = "push_to_talk"
)

type State struct {
	Enabled           bool
	Listening         bool
	PushToTalk        bool
	SessionActive     bool
	SessionProvider   string
	WakeWordArmed     bool
	WakeWordThreshold int
	Transcript        string
	TranscriptSeq     int
	UpdatedUnix       int64
	SessionCount      int
	Capabilities      []Capability
}

type LocalRuntime struct {
	mu                sync.RWMutex
	enabled           bool
	listening         bool
	pushToTalk        bool
	sessionActive     bool
	sessionProvider   string
	wakeWordArmed     bool
	wakeWordThreshold int
	transcript        string
	transcriptSeq     int
	lastUpdated       time.Time
	sessionCount      int
	capabilities      map[Capability]bool
}

func NewLocalRuntime(caps []Capability) *LocalRuntime {
	m := map[Capability]bool{}
	for _, c := range caps {
		m[c] = true
	}
	return &LocalRuntime{capabilities: m, wakeWordThreshold: 62, lastUpdated: time.Now()}
}

func (r *LocalRuntime) State() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	capabilities := make([]Capability, 0, len(r.capabilities))
	for c, ok := range r.capabilities {
		if ok {
			capabilities = append(capabilities, c)
		}
	}
	sort.Slice(capabilities, func(i, j int) bool { return capabilities[i] < capabilities[j] })
	return State{
		Enabled:           r.enabled,
		Listening:         r.listening,
		PushToTalk:        r.pushToTalk,
		SessionActive:     r.sessionActive,
		SessionProvider:   r.sessionProvider,
		WakeWordArmed:     r.wakeWordArmed,
		WakeWordThreshold: r.wakeWordThreshold,
		Transcript:        r.transcript,
		TranscriptSeq:     r.transcriptSeq,
		UpdatedUnix:       r.lastUpdated.Unix(),
		SessionCount:      r.sessionCount,
		Capabilities:      capabilities,
	}
}

func (r *LocalRuntime) Supports(cap Capability) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.capabilities[cap]
}

func (r *LocalRuntime) SetEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = enabled
	if !enabled {
		r.listening = false
		r.pushToTalk = false
		r.sessionActive = false
		r.wakeWordArmed = false
	}
	r.lastUpdated = time.Now()
}

func (r *LocalRuntime) SetListening(listening bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled {
		r.listening = false
		r.pushToTalk = false
		return false
	}
	r.listening = listening
	if !listening {
		r.pushToTalk = false
	}
	r.lastUpdated = time.Now()
	return true
}

func (r *LocalRuntime) BeginPushToTalk() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled || !r.capabilities[CapabilityPushToTalk] {
		r.pushToTalk = false
		return false
	}
	r.listening = true
	r.pushToTalk = true
	r.lastUpdated = time.Now()
	return true
}

func (r *LocalRuntime) EndPushToTalk() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pushToTalk = false
	r.listening = false
	r.lastUpdated = time.Now()
}

func (r *LocalRuntime) ArmWakeWord(threshold int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled || !r.capabilities[CapabilityWakeWord] {
		r.wakeWordArmed = false
		return false
	}
	if threshold <= 0 {
		threshold = r.wakeWordThreshold
	}
	r.wakeWordThreshold = threshold
	r.wakeWordArmed = true
	r.lastUpdated = time.Now()
	return true
}

func (r *LocalRuntime) DisarmWakeWord() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wakeWordArmed = false
	r.lastUpdated = time.Now()
}

func (r *LocalRuntime) StartSession(provider string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled {
		return false
	}
	r.sessionActive = true
	r.sessionProvider = provider
	r.sessionCount++
	r.lastUpdated = time.Now()
	return true
}

func (r *LocalRuntime) EndSession() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessionActive = false
	r.sessionProvider = ""
	r.pushToTalk = false
	r.listening = false
	r.lastUpdated = time.Now()
}

func (r *LocalRuntime) AppendTranscript(delta string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if delta == "" {
		return r.transcriptSeq
	}
	r.transcript += delta
	r.transcriptSeq++
	r.lastUpdated = time.Now()
	return r.transcriptSeq
}

func (r *LocalRuntime) ClearTranscript() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transcript = ""
	r.transcriptSeq++
	r.lastUpdated = time.Now()
}
