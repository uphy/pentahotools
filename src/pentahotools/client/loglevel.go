package client

// LogLevel is the level of job/trans logs.
type LogLevel string

// LogLevels is the level of job/trans logs.
var LogLevels = struct {
	Nothing  LogLevel
	Error    LogLevel
	Minimal  LogLevel
	Basic    LogLevel
	Detailed LogLevel
	Debug    LogLevel
	Rowlevel LogLevel
}{"Nothing", "Error", "Minimal", "Basic", "Detailed", "Debug", "Rowlevel"}
