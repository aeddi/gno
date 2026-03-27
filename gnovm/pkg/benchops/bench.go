package benchops

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ---- nesting tracer -------------------------------------------------------
//
// Records every StartStore/StopStore call while a nesting event is active,
// then prints the full call tree when the outer StopStore closes.
//
// For each event the full Go callstack is captured (inner-to-outer order).
// When printing, the inner StartStore's stack already contains the outer
// caller's frames deeper in the chain, so a single reversed stack gives the
// complete path from top-level caller down to the nested call site.

const maxNestingEvents = 5 // stop after this many events to avoid flooding

type storeCallEvent struct {
	kind      string // "StartStore" or "StopStore"
	storeCode byte
	depth     int
	frames    []string // inner-to-outer: [0] = direct caller, [-1] = e.g. main
}

var (
	nestDepth      int
	nestEventCount int
	nestBuf        []storeCallEvent // buffered events for current outer window
	outerFrame     string           // function name of outer StartStore's direct caller
)

// collectStack captures up to n application frames, filtering out runtime and
// benchops internals. Returns them inner-to-outer: frames[0] is the direct
// caller of StartStore or StopStore.
func collectStack(n int) []string {
	pcs := make([]uintptr, 128)
	count := runtime.Callers(2, pcs) // skip runtime.Callers + collectStack
	frames := runtime.CallersFrames(pcs[:count])
	var out []string
	for len(out) < n {
		f, more := frames.Next()
		if !strings.Contains(f.Function, "gnovm/pkg/benchops") &&
			!strings.HasPrefix(f.Function, "runtime.") {
			short := f.File
			if i := strings.Index(short, "gnovm/"); i >= 0 {
				short = short[i:]
			}
			out = append(out, fmt.Sprintf("%s (%s:%d)", f.Function, short, f.Line))
		}
		if !more {
			break
		}
	}
	return out
}

func onStartStore(code byte) {
	nestDepth++
	ev := storeCallEvent{
		kind:      "StartStore",
		storeCode: code,
		depth:     nestDepth,
		frames:    collectStack(24),
	}
	if nestDepth == 1 {
		// Outer call: start a fresh buffer and note the outer caller.
		nestBuf = []storeCallEvent{ev}
		if len(ev.frames) > 0 {
			outerFrame = ev.frames[0] // direct caller of outer StartStore
		}
	} else {
		// Inner call: nesting detected — append to buffer.
		nestBuf = append(nestBuf, ev)
	}
}

func onStopStore(code byte) {
	ev := storeCallEvent{
		kind:      "StopStore",
		storeCode: code,
		depth:     nestDepth,
		frames:    collectStack(24),
	}
	nestBuf = append(nestBuf, ev)
	nestDepth--

	if nestDepth == 0 {
		printNestingEvent()
		nestBuf = nil
		outerFrame = ""
	}
}

func printNestingEvent() {
	// Only print if nesting actually occurred.
	hasInner := false
	for _, e := range nestBuf {
		if e.depth > 1 {
			hasInner = true
			break
		}
	}
	if !hasInner || nestEventCount >= maxNestingEvents {
		return
	}
	nestEventCount++
	fmt.Printf("\n━━━ nesting event #%d ━━━\n\n", nestEventCount)

	// Find the outer StartStore (depth == 1).
	var outerEv *storeCallEvent
	for i := range nestBuf {
		if nestBuf[i].depth == 1 && nestBuf[i].kind == "StartStore" {
			outerEv = &nestBuf[i]
			break
		}
	}
	if outerEv == nil {
		return
	}

	// Reverse outer frames: outermost caller first.
	outerRev := make([]string, len(outerEv.frames))
	for i, f := range outerEv.frames {
		outerRev[len(outerEv.frames)-1-i] = f
	}

	// Print outer call chain with increasing indentation (2 spaces per level).
	for i, f := range outerRev {
		fmt.Printf("%s%s\n", strings.Repeat("  ", i+1), f)
	}
	// Arrow length scales with the depth of the last frame so the `> ` tip
	// lands 2 columns past the deepest frame's indent.
	// last_frame_indent = 2 * len(outerRev); target_width = last + 2
	// marker prefix "  outer " = 8 chars; dashes fill the remainder.
	outerDashes := max(4, 2*len(outerRev)-8)
	fmt.Printf("  outer %s> StartStore(%s)\n", strings.Repeat("-", outerDashes), StoreCodeString(outerEv.storeCode))

	// outerIndentLevel is the indent level of the outer's deepest frame.
	// Subsequent events start their unique frames at this same level.
	outerIndentLevel := len(outerRev)

	// Extract outer function name for junction detection in later stacks.
	outerFn := ""
	if outerFrame != "" {
		if idx := strings.IndexByte(outerFrame, ' '); idx >= 0 {
			outerFn = outerFrame[:idx]
		}
	}

	// Print all remaining events: inner events (depth > 1) and outer StopStore.
	// The outer StartStore is the only depth-1 event already handled above.
	for _, ev := range nestBuf {
		if ev.depth == 1 && ev.kind == "StartStore" {
			continue
		}

		// Reverse frames: outermost caller first.
		reversed := make([]string, len(ev.frames))
		for i, f := range ev.frames {
			reversed[len(ev.frames)-1-i] = f
		}

		// Find the junction frame (the outerFn function at a different call
		// site) and display from that frame onward so the reader sees exactly
		// where inside the outer function this event originates.
		display := reversed
		if outerFn != "" {
			for i, f := range reversed {
				if strings.HasPrefix(f, outerFn) {
					display = reversed[i:]
					break
				}
			}
		}

		fmt.Println()
		for i, f := range display {
			fmt.Printf("%s%s\n", strings.Repeat("  ", outerIndentLevel+i), f)
		}

		// Arrow scales to align `> ` two columns past the last frame's indent.
		dashes := max(4, 2*(outerIndentLevel+len(display)-1)-8)
		label := "inner"
		if ev.depth == 1 {
			label = "outer"
		}
		annotation := ""
		if ev.kind == "StartStore" {
			annotation = "  ← curStart overwritten here"
		}
		fmt.Printf("  %s %s> %s(%s)%s\n", label, strings.Repeat("-", dashes), ev.kind, StoreCodeString(ev.storeCode), annotation)
	}
	fmt.Println()
}

const (
	invalidCode = byte(0x00)
)

var measure bench

type bench struct {
	// Opcode timing: single timeline, always one op active.
	opCounts   [256]int64
	opAccumDur [256]time.Duration
	curOpCode  byte
	curStart   time.Time
	timeZero   time.Time

	// Store timing: own accumulators, uses SwitchOpCode for handoff.
	storeCounts    [256]int64
	storeAccumDur  [256]time.Duration
	storeAccumSize [256]int64

	// Native timing.
	nativeCounts   [256]int64
	nativeAccumDur [256]time.Duration
}

func InitMeasure() {
	measure = bench{
		curOpCode: invalidCode,
	}
}

// finalizeCurrentOp attributes elapsed time since curStart to
// curOpCode and returns the snapshot time.
func finalizeCurrentOp() time.Time {
	now := time.Now()
	if measure.curOpCode != invalidCode && measure.curStart != measure.timeZero {
		measure.opAccumDur[measure.curOpCode] += now.Sub(measure.curStart)
	}
	measure.curStart = measure.timeZero
	return now
}

// SwitchOpCode finalizes the current op's elapsed time and
// starts timing a new op. Returns the old op code so the
// caller can pass it to resumeOpCode when done.
func SwitchOpCode(code byte) byte {
	if code == invalidCode {
		panic("the OpCode is invalid")
	}
	old := measure.curOpCode
	now := finalizeCurrentOp()
	measure.curOpCode = code
	measure.curStart = now
	measure.opCounts[code]++
	return old
}

// resumeOpCode resumes a previous op without incrementing its count.
// Used by StopStore/StopNative to hand back to the parent op.
func resumeOpCode(code byte) {
	if measure.curOpCode != invalidCode {
		panic("resumeOpCode called with active op")
	}
	measure.curOpCode = code
	measure.curStart = time.Now()
}

// StopOpCode finalizes the current op. Used at OpHalt/return.
func StopOpCode() {
	finalizeCurrentOp()
	measure.curOpCode = invalidCode
}

// ---- Store operations ----

// StartStore suspends the current VM op timer and begins a
// store operation. Returns the old op code.
func StartStore(storeCode byte) byte {
	onStartStore(storeCode)
	old := measure.curOpCode
	// Finalize the VM op's time up to now.
	now := finalizeCurrentOp()
	// Park the timeline — store tracks its own duration.
	measure.curOpCode = invalidCode
	measure.storeCounts[storeCode]++
	// Store the start time in curStart temporarily;
	// StopStore will read it.
	measure.curStart = now
	return old
}

// StopStore ends the store operation, records its duration
// and size, then resumes the previous VM op.
func StopStore(storeCode byte, old byte, size int) {
	onStopStore(storeCode)
	now := time.Now()
	if measure.curStart != measure.timeZero {
		measure.storeAccumDur[storeCode] += now.Sub(measure.curStart)
	}
	measure.storeAccumSize[storeCode] += int64(size)
	resumeOpCode(old)
}

// ---- Native operations ----

func StartNative(nativeCode byte) byte {
	if nativeCode == invalidCode {
		panic("the NativeCode is invalid")
	}
	old := measure.curOpCode
	finalizeCurrentOp() // finalize previous op BEFORE GC
	runtime.GC()
	now := time.Now() // fresh timestamp after GC
	measure.curOpCode = invalidCode
	measure.nativeCounts[nativeCode]++
	measure.curStart = now
	return old
}

func StopNative(nativeCode byte, old byte) {
	now := time.Now()
	if measure.curStart != measure.timeZero {
		measure.nativeAccumDur[nativeCode] += now.Sub(measure.curStart)
	}
	resumeOpCode(old)
}
