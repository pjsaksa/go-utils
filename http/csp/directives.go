package csp

import "fmt"

const (
	DefaultSrc Directive = iota
	//
	BaseUri
	ChildSrc
	ConnectSrc
	FontSrc
	FormAction
	FrameAncestors
	FrameSrc
	ImgSrc
	ManifestSrc
	MediaSrc
	ObjectSrc
	ReportTo
	RequireTrustedTypesFor
	Sandbox
	ScriptSrc
	ScriptSrcAttr
	ScriptSrcElem
	StyleSrc
	StyleSrcAttr
	StyleSrcElem
	TrustedTypes
	UpgradeInsecureRequests
	WorkerSrc
	//
	DirectiveCount
)

func (d Directive) String() string {
	switch d {
	case DefaultSrc:
		return "default-src"
	case BaseUri:
		return "base-uri"
	case ChildSrc:
		return "child-src"
	case ConnectSrc:
		return "connect-src"
	case FontSrc:
		return "font-src"
	case FormAction:
		return "form-action"
	case FrameAncestors:
		return "frame-ancestors"
	case FrameSrc:
		return "frame-src"
	case ImgSrc:
		return "img-src"
	case ManifestSrc:
		return "manifest-src"
	case MediaSrc:
		return "media-src"
	case ObjectSrc:
		return "object-src"
	case ReportTo:
		return "report-to"
	case RequireTrustedTypesFor:
		return "require-trusted-types-for"
	case Sandbox:
		return "sandbox"
	case ScriptSrc:
		return "script-src"
	case ScriptSrcAttr:
		return "script-src-attr"
	case ScriptSrcElem:
		return "script-src-elem"
	case StyleSrc:
		return "style-src"
	case StyleSrcAttr:
		return "style-src-attr"
	case StyleSrcElem:
		return "style-src-elem"
	case TrustedTypes:
		return "trusted-types"
	case UpgradeInsecureRequests:
		return "update-insecure-requests"
	case WorkerSrc:
		return "worker-src"
	}
	panic(fmt.Errorf("Directive.String() called with invalid value: %#v", d))
}
