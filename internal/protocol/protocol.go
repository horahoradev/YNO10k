package protocol

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/divan/num2words"
)

/*
	e.g.
	type wow struct {
		MatchPrefix string `ynoproto:"m"`
		One string `ynoproto:"nonempty"`
		Two int `ynoproto:"nonnegative"`
	}
*/

var delimChar string = "\uffff"

func Marshal(msgbuf []byte, target interface{}) (matched bool, err error) {
	e := reflect.ValueOf(target).Elem()

	t := e.Type()
	log.Print(t)
	// Determine match prefix for target struct
	pref, ok := t.FieldByName("MatchPrefix")
	if !ok {
		return false, errors.New("target missing required attribute MatchPrefix")
	}

	spl := strings.Split(string(msgbuf), delimChar)

	if len(spl) == 0 {
		return false, errors.New("empty message")
	}

	msgPrefix, ok := pref.Tag.Lookup("ynoproto")
	if !ok {
		return false, errors.New("Missing required MatchPrefix annotation ynoproto")
	}

	// This doesn't indicate an error; it just didn't match.
	if msgPrefix != spl[0] {
		return false, nil
	}

	// The message needs to match the target's number of fields
	// e.g. if we get m, it better have two fields
	if t.NumField() != len(spl) {
		return false, fmt.Errorf("target's number of fields, %d, does not match message's number of fields, %d", t.NumField(), len(spl))
	}

	for i := 1; i < len(spl); i++ {
		arg := spl[i]

		numWord := strings.Title(num2words.Convert(i))
		f := e.FieldByName(numWord)
		zeroVal := reflect.Value{}
		if f == zeroVal {
			return false, fmt.Errorf("Missing required struct field %s", numWord)
		}
		if !f.IsValid() {
			return false, fmt.Errorf("field %s is not valid", numWord)
		}

		if !f.CanSet() {
			return false, fmt.Errorf("could not set field %s", numWord)
		}

		switch f.Kind() {
		case reflect.Int:
			n, err := strconv.Atoi(arg)
			if err != nil {
				return false, fmt.Errorf("Failed to convert arg to int. Err: %s", err)
			}

			if f.OverflowInt(int64(n)) {
				return false, fmt.Errorf("provided value, %d, would overflow if assigned to struct type", n)
			}

			f.SetInt(int64(n))

		case reflect.String:
			f.SetString(arg)

		default:
			return false, errors.New("Unsupported type used in struct")
		}
	}

	return true, nil
}