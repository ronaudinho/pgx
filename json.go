package pgtype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Json struct {
	Bytes  []byte
	Status Status
}

func (dst *Json) Set(src interface{}) error {
	if src == nil {
		*dst = Json{Status: Null}
		return nil
	}

	switch value := src.(type) {
	case string:
		*dst = Json{Bytes: []byte(value), Status: Present}
	case *string:
		if value == nil {
			*dst = Json{Status: Null}
		} else {
			*dst = Json{Bytes: []byte(*value), Status: Present}
		}
	case []byte:
		if value == nil {
			*dst = Json{Status: Null}
		} else {
			*dst = Json{Bytes: value, Status: Present}
		}
	default:
		buf, err := json.Marshal(value)
		if err != nil {
			return err
		}
		*dst = Json{Bytes: buf, Status: Present}
	}

	return nil
}

func (dst *Json) Get() interface{} {
	switch dst.Status {
	case Present:
		var i interface{}
		err := json.Unmarshal(dst.Bytes, &i)
		if err != nil {
			return dst
		}
		return i
	case Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Json) AssignTo(dst interface{}) error {
	switch v := dst.(type) {
	case *string:
		if src.Status != Present {
			v = nil
		} else {
			*v = string(src.Bytes)
		}
	case **string:
		*v = new(string)
		return src.AssignTo(*v)
	case *[]byte:
		if src.Status != Present {
			*v = nil
		} else {
			buf := make([]byte, len(src.Bytes))
			copy(buf, src.Bytes)
			*v = buf
		}
	default:
		data := src.Bytes
		if data == nil || src.Status != Present {
			data = []byte("null")
		}

		return json.Unmarshal(data, dst)
	}

	return nil
}

func (dst *Json) DecodeText(ci *ConnInfo, src []byte) error {
	if src == nil {
		*dst = Json{Status: Null}
		return nil
	}

	*dst = Json{Bytes: src, Status: Present}
	return nil
}

func (dst *Json) DecodeBinary(ci *ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *Json) EncodeText(ci *ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case Null:
		return nil, nil
	case Undefined:
		return nil, errUndefined
	}

	return append(buf, src.Bytes...), nil
}

func (src *Json) EncodeBinary(ci *ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Json) Scan(src interface{}) error {
	if src == nil {
		*dst = Json{Status: Null}
		return nil
	}

	switch src := src.(type) {
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		return dst.DecodeText(nil, srcCopy)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src *Json) Value() (driver.Value, error) {
	switch src.Status {
	case Present:
		return string(src.Bytes), nil
	case Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
