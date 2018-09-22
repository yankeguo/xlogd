package main

import (
	"encoding/json"
	"strings"
	"time"
)

func decodeBeatMessage(raw string, r *Record) bool {
	var err error
	var match []string
	// trim message
	raw = strings.TrimSpace(raw)
	// search timestamp
	if match = eventLinePattern.FindStringSubmatch(raw); len(match) != 2 {
		return false
	}
	// parse timestamp
	if r.Timestamp, err = time.Parse(eventTimestampLayout, match[1]); err != nil {
		return false
	}
	// trim message
	r.Message = strings.TrimSpace(raw[len(match[0]):])
	// find crid
	if match = eventCridPattern.FindStringSubmatch(r.Message); len(match) == 2 {
		r.Crid = match[1]
	} else {
		r.Crid = "-"
	}
	return true
}

func decodeBeatSource(raw string, r *Record) bool {
	var cs []string
	// trim source
	raw = strings.TrimSpace(raw)
	if cs = strings.Split(raw, "/"); len(cs) < 3 {
		return false
	}
	// assign fields
	r.Env, r.Topic, r.Project = cs[len(cs)-3], cs[len(cs)-2], cs[len(cs)-1]
	// sanitize dot separated filename
	var ss []string
	if ss = strings.Split(r.Project, "."); len(ss) > 0 {
		r.Project = ss[0]
	}
	return true
}

func mergeRecordExtra(r *Record) bool {
	var err error
	r.Extra = map[string]interface{}{}
	if err = json.Unmarshal([]byte(r.Message), &r.Extra); err != nil {
		return false
	}
	// topic must exist
	if !decodeExtraStr(r.Extra, "topic", &r.Topic) {
		return false
	}
	// optional extra 'project', 'crid'
	decodeExtraStr(r.Extra, "project", &r.Project)
	decodeExtraStr(r.Extra, "crid", &r.Crid)
	// optional extract 'timestamp'
	if decodeExtraTime(r.Extra, "timestamp", &r.Timestamp) {
		r.NoTimeOffset = true
	}
	// clear the message
	r.Message = ""
	return true
}

func decodeExtraStr(m map[string]interface{}, key string, out *string) bool {
	if m == nil || out == nil {
		return false
	}
	if val, ok := m[key].(string); ok {
		val = strings.TrimSpace(val)
		delete(m, key) // always delete
		if len(val) > 0 {
			*out = val // update if not empty
			return true
		}
	}
	return false
}

func decodeExtraTime(m map[string]interface{}, key string, out *time.Time) bool {
	if m == nil || out == nil {
		return false
	}
	var tsStr string
	if decodeExtraStr(m, key, &tsStr) {
		if t, err := time.Parse(time.RFC3339, tsStr); err != nil {
			return false
		} else {
			*out = t // update if success
			return true
		}
	}
	return false
}