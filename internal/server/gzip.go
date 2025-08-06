package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/go-kratos/kratos/v2/encoding"
)

// Name is the name registered for the json codec.
const Name = "gzip"

var (
	// MarshalOptions is a configurable JSON format marshaller.
	MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
	// UnmarshalOptions is a configurable JSON format parser.
	UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func init() {
	encoding.RegisterCodec(codec{})
}

// codec is a Codec implementation with gzip.
type codec struct{}

func (codec) Marshal(v any) ([]byte, error) {
	var zBuf bytes.Buffer
	zw := gzip.NewWriter(&zBuf)
	//defer zw.Close()
	respJson := []byte{}

	switch m := v.(type) {
	case json.Marshaler:
		respJson, _ = m.MarshalJSON()
	case proto.Message:
		respJson, _ = MarshalOptions.Marshal(m)
	default:
		respJson, _ = json.Marshal(m)
	}

	if _, err := zw.Write(respJson); err != nil {
		fmt.Println("gzip is faild,err:", err)
	}
	zw.Close()
	return zBuf.Bytes(), nil
}

func (codec) Unmarshal(data []byte, v any) error {
	body, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		fmt.Println("unzip is failed, err:", err)
	}
	defer body.Close()
	data, err = io.ReadAll(body)
	if err != nil {
		fmt.Println("read all is failed.err:", err)
	}

	switch m := v.(type) {
	case json.Unmarshaler:
		return m.UnmarshalJSON(data)
	case proto.Message:
		return UnmarshalOptions.Unmarshal(data, m)
	default:
		rv := reflect.ValueOf(v)
		for rv := rv; rv.Kind() == reflect.Ptr; {
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		if m, ok := reflect.Indirect(rv).Interface().(proto.Message); ok {
			return UnmarshalOptions.Unmarshal(data, m)
		}
		return json.Unmarshal(data, m)
	}
}

func (codec) Name() string {
	return Name
}
