package generator

import (
    "strings"
    "time"
)

// TimestampNow returns a UTC timestamp formatted as YYYYMMDDHHMMSS.
func TimestampNow() string {
    return time.Now().UTC().Format("20060102150405")
}

// TableName returns a simple pluralized table name for a resource.
// It's intentionally naive: if name ends with 's' it is returned as-is,
// otherwise we append 's'. This is sufficient for prototype scaffolding.
func TableName(name string) string {
    name = strings.ToLower(name)
    if strings.HasSuffix(name, "s") {
        return name
    }
    return name + "s"
}
