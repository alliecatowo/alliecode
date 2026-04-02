package voice

import "testing"

func TestLocalRuntimeStateAndCapabilities(t *testing.T) {
	rt := NewLocalRuntime([]Capability{CapabilityWakeWord, CapabilityDictation})

	if rt.Supports(CapabilityPushToTalk) {
		t.Fatal("did not expect push_to_talk capability")
	}
	if !rt.Supports(CapabilityDictation) {
		t.Fatal("expected dictation capability")
	}

	st := rt.State()
	if st.Enabled || st.Listening {
		t.Fatalf("expected disabled initial state, got %+v", st)
	}
	if st.UpdatedUnix == 0 {
		t.Fatalf("expected initial update timestamp")
	}
}

func TestLocalRuntimeEnableAndListeningGuards(t *testing.T) {
	rt := NewLocalRuntime(nil)

	if ok := rt.SetListening(true); ok {
		t.Fatal("expected listening enable to fail when disabled")
	}

	rt.SetEnabled(true)
	if ok := rt.SetListening(true); !ok {
		t.Fatal("expected listening enable to succeed when enabled")
	}

	st := rt.State()
	if !st.Enabled || !st.Listening {
		t.Fatalf("expected enabled+listening state, got %+v", st)
	}

	rt.SetEnabled(false)
	st = rt.State()
	if st.Listening {
		t.Fatalf("expected listening to clear when disabled, got %+v", st)
	}
}

func TestLocalRuntimePushToTalkWakeWordAndSessionState(t *testing.T) {
	rt := NewLocalRuntime([]Capability{CapabilityPushToTalk, CapabilityWakeWord, CapabilityDictation})
	rt.SetEnabled(true)

	if !rt.BeginPushToTalk() {
		t.Fatal("expected push-to-talk begin to succeed")
	}
	st := rt.State()
	if !st.Listening || !st.PushToTalk {
		t.Fatalf("expected active push-to-talk listening state, got %+v", st)
	}

	rt.EndPushToTalk()
	st = rt.State()
	if st.Listening || st.PushToTalk {
		t.Fatalf("expected push-to-talk release to clear listening, got %+v", st)
	}

	if !rt.ArmWakeWord(80) {
		t.Fatal("expected wake-word arm to succeed")
	}
	st = rt.State()
	if !st.WakeWordArmed || st.WakeWordThreshold != 80 {
		t.Fatalf("expected armed wake-word state, got %+v", st)
	}

	if !rt.StartSession("local") {
		t.Fatal("expected local voice session to start")
	}
	seq := rt.AppendTranscript("hello")
	if seq != 1 {
		t.Fatalf("expected transcript sequence 1, got %d", seq)
	}
	rt.AppendTranscript(" world")
	st = rt.State()
	if !st.SessionActive || st.SessionProvider != "local" {
		t.Fatalf("expected active local session state, got %+v", st)
	}
	if st.SessionCount != 1 {
		t.Fatalf("expected session count increment, got %+v", st)
	}
	if st.Transcript != "hello world" || st.TranscriptSeq != 2 {
		t.Fatalf("expected transcript accumulation, got %+v", st)
	}

	rt.ClearTranscript()
	st = rt.State()
	if st.Transcript != "" || st.TranscriptSeq != 3 {
		t.Fatalf("expected transcript clear with sequence bump, got %+v", st)
	}

	rt.EndSession()
	st = rt.State()
	if st.SessionActive || st.Listening || st.PushToTalk {
		t.Fatalf("expected session end to clear runtime activity, got %+v", st)
	}
}
