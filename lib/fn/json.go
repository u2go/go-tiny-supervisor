package fn

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func JsonPretty(v any) string {
	return myJson(v, func(enc *json.Encoder) {
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
	})
}

func Json(v any) string {
	return myJson(v, func(enc *json.Encoder) {
		enc.SetEscapeHTML(false)
	})
}

func myJson(v any, optFn func(enc *json.Encoder)) string {
	w := bytes.NewBuffer(nil)
	enc := json.NewEncoder(w)
	optFn(enc)
	err := enc.Encode(v)
	if err != nil {
		return "myJson:error: " + err.Error() + "\ndata: " + fmt.Sprintf("%+v", v)
	}
	return w.String()[0 : len(w.String())-1]
}
