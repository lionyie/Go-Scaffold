package server

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	khttp "github.com/go-kratos/kratos/v3/transport/http"
)

// Kratos' default "json" codec uses encoding/json, which emits snake_case
// struct tags and numeric enums. The API contract is protojson (camelCase,
// enum names, 64-bit integers as strings), so both directions are forced
// through protojson while non-proto payloads keep the default behaviour.

func protojsonRequestDecoder(r *http.Request, v any) error {
	msg, ok := asProtoMessage(v)
	if !ok {
		return khttp.DefaultRequestDecoder(r, v)
	}
	if contentType := r.Header.Get("Content-Type"); !strings.Contains(strings.ToLower(contentType), "json") {
		return khttp.DefaultRequestDecoder(r, v)
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewReader(data))
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, msg)
}

func protojsonResponseEncoder(w http.ResponseWriter, r *http.Request, v any) error {
	msg, ok := asProtoMessage(v)
	if !ok {
		return khttp.DefaultResponseEncoder(w, r, v)
	}
	data, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(msg)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

// asProtoMessage accepts both T and *T where T is a proto.Message, because
// generated handlers pass fields such as **Project to ctx.Bind.
func asProtoMessage(v any) (proto.Message, bool) {
	if msg, ok := v.(proto.Message); ok {
		return msg, true
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return nil, false
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Pointer {
		return nil, false
	}
	if elem.IsNil() {
		elem.Set(reflect.New(elem.Type().Elem()))
	}
	msg, ok := elem.Interface().(proto.Message)
	return msg, ok
}
