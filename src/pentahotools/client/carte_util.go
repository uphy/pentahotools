package client

import (
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"regexp"
	"strings"
)

// ParseIDAndName parses ID and Job.
func ParseIDAndName(s string) (string, string) {
	isID, _ := regexp.Match(`[a-f0-9]{8}\-[a-f0-9]{4}\-[a-f0-9]{4}\-[a-f0-9]{4}\-[a-f0-9]{12}`, []byte(s))
	if isID {
		return s, ""
	}
	return "", s
}

func decodeLoggingString(loggingString string) string {
	loggingString = loggingString[len("<![CDATA[") : len(loggingString)-len("]]>")]
	if len(loggingString) == 0 {
		return ""
	}
	reader := strings.NewReader(loggingString)
	base64Reader := base64.NewDecoder(base64.StdEncoding, reader)
	gzipReader, _ := gzip.NewReader(base64Reader)
	decoded, _ := ioutil.ReadAll(gzipReader)
	return string(decoded)
}
