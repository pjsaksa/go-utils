package csp

import "fmt"

const (
	None constValue = iota
	//
	InlineSpeculationRules
	ReportSample
	Self
	StrictDynamic
	TrustedTypesEval
	UnsafeEval
	UnsafeHashes
	UnsafeInline
	WasmUnsafeEval
)

func (v constValue) cspValue() string {
	switch v {
	case None:
		return "'none'"
	case InlineSpeculationRules:
		return "'inline-speculation-rules'"
	case ReportSample:
		return "'report-sample'"
	case Self:
		return "'self'"
	case StrictDynamic:
		return "'strict-dynamic'"
	case TrustedTypesEval:
		return "'trusted-types-eval'"
	case UnsafeEval:
		return "'unsafe-eval'"
	case UnsafeHashes:
		return "'unsafe-hashes'"
	case UnsafeInline:
		return "'unsafe-inline'"
	case WasmUnsafeEval:
		return "'wasm-unsafe-eval'"
	}
	panic(fmt.Errorf("constValue.String() called with invalid value: %#v", v))
}
