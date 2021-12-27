package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"reflect"
)

/*
	e.g.
	type wow struct {
		MatchPrefix string `ynoproto:"m"`
		One string `ynoproto:"nonempty"`
		Two int `ynoproto:"nonnegative"`
	}
*/

func Marshal(msgbuf []byte, target interface{}) (matched bool, err error) {
	e := reflect.ValueOf(target).Elem()

	t := e.Type()

	// Determine match prefix for target struct
	pref, ok := t.FieldByName("MatchPrefix")
	if !ok {
		return false, errors.New("target missing required attribute MatchPrefix")
	}

	if len(msgbuf) == 0 {
		return false, errors.New("Empty message")
	}

	msgPrefix, ok := pref.Tag.Lookup("ynoproto")
	if !ok {
		return false, errors.New("Missing required MatchPrefix annotation ynoproto")
	}

	msgVal, err := hex.DecodeString(msgPrefix)
	if err != nil {
		return false, errors.New("Failed to decode matchprefix to hex")
	}

	if len(msgVal) != 1 {
		return false, fmt.Errorf("msgVal has an incorrect length of %d", len(msgVal))
	}

	// This doesn't indicate an error; it just didn't match.
	if msgVal[0] != msgbuf[0] {
		log.Printf("%s %s", msgPrefix, string(msgbuf[0]))
		return false, nil
	}

	hasString := false
	totalFieldSize := 1
	// Start at 1 byte offset due to prefix
	for i, attrNum := 1, 1; i < len(msgbuf) && attrNum < t.NumField(); attrNum++ {
		f := e.FieldByIndex([]int{attrNum})
		ft := t.FieldByIndex([]int{attrNum})
		fieldSize := ft.Type.Size()
		totalFieldSize += int(fieldSize)
		fieldName := ft.Name

		// tagsVal, ok := ft.Tag.Lookup("ynoproto")
		// if !ok {
		// 	return false, fmt.Errorf("Failed to retrieve ynoproto tag for field %s", fieldName)
		// }

		// MUST be comma delimited
		// tags := strings.Split(tagsVal, ",")

		if !f.IsValid() {
			return false, fmt.Errorf("field %s is not valid", fieldName)
		}

		if !f.CanSet() {
			return false, fmt.Errorf("could not set field %s", fieldName)
		}

		switch f.Kind() {
		case reflect.Uint16:
			n := binary.BigEndian.Uint16(msgbuf[i:])

			// Not sure if this is really valid for uint16
			if f.OverflowUint(uint64(n)) {
				return false, fmt.Errorf("provided value, %d, would overflow if assigned to struct type", n)
			}

			f.SetUint(uint64(n))

		case reflect.String:
			hasString = true
			f.SetString(string(msgbuf[i:]))

		default:
			return false, errors.New("Unsupported type used in struct")
		}

		i += int(fieldSize)
	}

	// Can't use the reflected size of the struct because of word allignment padding
	if !hasString && uint64(totalFieldSize) != uint64(len(msgbuf)) {
		return false, fmt.Errorf("struct size %d did not match len of msgbuf %d", totalFieldSize, len(msgbuf))
	}

	return true, nil
}
