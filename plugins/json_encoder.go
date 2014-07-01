/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2013-2014
# the Initial Developer. All Rights Reserved.
#
# Contributor(s):
#   Tanguy Leroux (tlrx.dev@gmail.com)
#   Rob Miller (rmiller@mozilla.com)
#
# ***** END LICENSE BLOCK *****/

package plugins

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/mozilla-services/heka/message"
	. "github.com/mozilla-services/heka/pipeline"
	"strconv"
	"strings"
	"time"
)

// Append a field (with a name and a value) to a Buffer.
func writeField(first bool, b *bytes.Buffer, name string, value string) {
	if !first {
		b.WriteString(`,`)
	}
	b.WriteString(`"`)
	b.WriteString(name)
	b.WriteString(`":`)
	b.WriteString(value)
}

const lowerhex = "0123456789abcdef"

func writeUTF16Escape(b *bytes.Buffer, c rune) {
	b.WriteString(`\u`)
	b.WriteByte(lowerhex[(c>>12)&0xF])
	b.WriteByte(lowerhex[(c>>8)&0xF])
	b.WriteByte(lowerhex[(c>>4)&0xF])
	b.WriteByte(lowerhex[c&0xF])
}

// Go json encoder will blow up on invalid utf8 so we use this custom json
// encoder. Also, go json encoder generates these funny \U escapes which I
// don't think are valid json.

// Also note that invalid utf-8 sequences get encoded as U+FFFD this is a
// feature. :)

func writeQuotedString(b *bytes.Buffer, str string) {
	b.WriteString(`"`)

	for _, c := range str {
		if c == 0x20 || c == 0x21 || (c >= 0x23 && c <= 0x5B) || (c >= 0x5D) {
			b.WriteRune(c)
		} else {
			writeUTF16Escape(b, c)
		}

	}
	b.WriteString(`"`)

}

func writeStringField(first bool, b *bytes.Buffer, name string, value string) {
	if !first {
		b.WriteString(`,`)
	}
	writeQuotedString(b, name)
	b.WriteString(`:`)
	writeQuotedString(b, value)
}

func writeRawField(first bool, b *bytes.Buffer, name string, value string) {
	if !first {
		b.WriteString(`,`)
	}
	writeQuotedString(b, name)
	b.WriteString(`:`)
	b.WriteString(value)
}

// Manually encodes the Heka message into an ElasticSearch friendly way.
type JsonEncoder struct {
	// Field names to include in ElasticSearch document for "clean" format.
	fields          []string
        timestampFormat string

}

type JsonEncoderConfig struct {
	Fields []string
	// Timestamp format. Defaults to "2006-01-02T15:04:05.000Z"
	Timestamp string
}

func (e *JsonEncoder) ConfigStruct() interface{} {
	config := &JsonEncoderConfig{
	Timestamp:            "2006-01-02T15:04:05.000Z",
	}

	config.Fields = []string{
		"Uuid",
		"Timestamp",
		"Type",
		"Logger",
		"Severity",
		"Payload",
		"EnvVersion",
		"Pid",
		"Hostname",
		"Fields",
	}

	return config
}

func (e *JsonEncoder) Init(config interface{}) (err error) {
	conf := config.(*JsonEncoderConfig)
	e.fields = conf.Fields
	e.timestampFormat = conf.Timestamp
	return
}

func (e *JsonEncoder) Encode(pack *PipelinePack) (output []byte, err error) {
	m := pack.Message
	buf := bytes.Buffer{}
	buf.WriteString(`{`)
	first := true
	for _, f := range e.fields {
		switch strings.ToLower(f) {
		case "uuid":
			writeStringField(first, &buf, f, m.GetUuidString())
		case "timestamp":
			t := time.Unix(0, m.GetTimestamp()).UTC()
			writeStringField(first, &buf, f, t.Format(e.timestampFormat))
		case "type":
			writeStringField(first, &buf, f, m.GetType())
		case "logger":
			writeStringField(first, &buf, f, m.GetLogger())
		case "severity":
			writeRawField(first, &buf, f, strconv.Itoa(int(m.GetSeverity())))
		case "payload":
			writeStringField(first, &buf, f, m.GetPayload())
		case "envversion":
			writeRawField(first, &buf, f, strconv.Quote(m.GetEnvVersion()))
		case "pid":
			writeRawField(first, &buf, f, strconv.Itoa(int(m.GetPid())))
		case "hostname":
			writeStringField(first, &buf, f, m.GetHostname())
		case "fields":
			raw := false
			for _, field := range m.Fields {
				if raw {
					data := field.GetValue().([]byte)[:]
					writeField(first, &buf, *field.Name, string(data))
					raw = false
				} else {
					switch field.GetValueType() {
					case message.Field_STRING:
						writeStringField(first, &buf, *field.Name, field.GetValue().(string))
					case message.Field_BYTES:
						data := field.GetValue().([]byte)[:]
						writeStringField(first, &buf, *field.Name,
							base64.StdEncoding.EncodeToString(data))
					case message.Field_INTEGER:
						writeRawField(first, &buf, *field.Name,
							strconv.FormatInt(field.GetValue().(int64), 10))
					case message.Field_DOUBLE:
						writeRawField(first, &buf, *field.Name,
							strconv.FormatFloat(field.GetValue().(float64), 'g', -1, 64))
					case message.Field_BOOL:
						writeRawField(first, &buf, *field.Name,
							strconv.FormatBool(field.GetValue().(bool)))
					}
				}
			}
		default:
			err = fmt.Errorf("Unable to find field: %s", f)
			return
		}
		first = false
	}
	buf.WriteString(`}`)
	buf.WriteByte(NEWLINE)
	return buf.Bytes(), err
}

// Manually encodes messages into JSON structure matching Logstash's "version
// 1" schema, for compatibility with Kibana's out-of-box Logstash dashboards.

func init() {
	RegisterPlugin("JsonEncoder", func() interface{} {
		return new(JsonEncoder)
	})
}
