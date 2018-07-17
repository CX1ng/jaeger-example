package common

import (
	"strings"

	"google.golang.org/grpc/metadata"
)

type MdWriterReader struct {
	metadata.MD
}

func (m MdWriterReader) Set(key, val string) {
	key = strings.ToLower(key)
	m.MD[key] = append(m.MD[key], val)
}

func (m MdWriterReader) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range m.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
