package common

import (
	"encoding/json"
	"fmt"
)

type TID struct {
	ID   int
	Next bool
}

func (f *TID) UnmarshalJSON(data []byte) error {
	var (
		err  error
		id   int
		next bool
	)
	if data[0] == QuotesByte {
		// target as string represents a value: "next:<id>", so the minimum length is 7
		if len(data) > 7 {
			err = json.Unmarshal(data[6:len(data)-1], &id)
			s := string(data[1:6])
			next = s == "next:"
		} else {
			return fmt.Errorf("target id not defined: %s", string(data))
		}
	} else {
		err = json.Unmarshal(data, &id)
	}
	*f = TID{
		ID:   id,
		Next: next,
	}
	return err
}
