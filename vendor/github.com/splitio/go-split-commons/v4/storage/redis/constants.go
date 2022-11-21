package redis

// Split redis keys, fields & TTLs
const (
	KeySplit            = "SPLITIO.split.{split}"                                    // split object
	KeySplitTill        = "SPLITIO.splits.till"                                      // last split fetch
	KeySegment          = "SPLITIO.segment.{segment}"                                // segment object
	KeySegmentTill      = "SPLITIO.segment.{segment}.till"                           // last segment fetch
	KeyEvents           = "SPLITIO.events"                                           // events LIST key
	KeyImpressionsQueue = "SPLITIO.impressions"                                      // impressions LIST key
	KeyTrafficType      = "SPLITIO.trafficType.{trafficType}"                        // traffic Type fetch
	KeyAPIKeyHash       = "SPLITIO.hash"                                             // hash key
	KeyConfig           = "SPLITIO.telemetry.config"                                 // config Key
	KeyLatency          = "SPLITIO.telemetry.latencies"                              // latency Key
	KeyException        = "SPLITIO.telemetry.exceptions"                             // exception Key
	KeyUniquekeys       = "SPLITIO.uniquekeys"                                       // Unique keys
	KeyImpressionsCount = "SPLITIO.impressions.count"                                // impressions count
	FieldLatency        = "{sdkVersion}/{machineName}/{machineIP}/{method}/{bucket}" // latency field template
	FieldException      = "{sdkVersion}/{machineName}/{machineIP}/{method}"          // exception field template
	TTLImpressions      = 3600                                                       // impressions default TTL
	TTLConfig           = 3600                                                       // config TTL
	TTLUniquekeys       = 3600                                                       // Uniquekeys TTL

	// TODO(mredolatti): when doing a breking change, name this `KeyConfig`, and rename `KeyConfig` to `KeyConfigLegacy`,
	// or even better, remove the old one, so that it only exists in the split-sync
	KeyInit        = "SPLITIO.telemetry.init"
	InitHashFields = "{sdkVersion}/{machineName}/{machineIP}"
)

// FieldSeparator constant
const (
	FieldSeparator = "/"
)

// Latency field section indexes
const (
	FieldLatencyIndexSdkVersion  = 0
	FieldLatencyIndexMachineName = 1
	FieldLatencyIndexMachineIP   = 2
	FieldLatencyIndexMethod      = 3
	FieldLatencyIndexBucket      = 4
)

// Exception field section indexes
const (
	FieldExceptionIndexSdkVersion  = 0
	FieldExceptionIndexMachineName = 1
	FieldExceptionIndexMachineIP   = 2
	FieldExceptionIndexMethod      = 3
)

// Latency hash-key indexes
const (
	TelemetryConfigIndexSdkVersion  = 0
	TelemetryConfigIndexMachineName = 1
	TelemetryConfigIndexMachineIP   = 2
)
