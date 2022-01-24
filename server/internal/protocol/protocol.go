package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
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

func Marshal(msgbuf []byte, target interface{}, twoBytePrefix bool) (matched bool, sequenceNumber uint8, err error) {
	// TODO: detect client socket endianness and correct somewher else lmao
	var seqNumber uint8 = 0
	if twoBytePrefix {
		// little endian LOL
		seqNumber = uint8(msgbuf[1])
		msgbuf = append([]byte{msgbuf[0]}, msgbuf[2:]...)
	}

	e := reflect.ValueOf(target).Elem()

	t := e.Type()

	// Determine match prefix for target struct
	pref, ok := t.FieldByName("MatchPrefix")
	if !ok {
		return false, 0, errors.New("target missing required attribute MatchPrefix")
	}

	if len(msgbuf) == 0 {
		return false, 0, errors.New("Empty message")
	}

	msgPrefix, ok := pref.Tag.Lookup("ynoproto")
	if !ok {
		return false, 0, errors.New("Missing required MatchPrefix annotation ynoproto")
	}

	msgVal, err := hex.DecodeString(msgPrefix)
	if err != nil {
		return false, 0, fmt.Errorf("Failed to decode matchprefix %s to hex", msgPrefix)
	}

	if len(msgVal) != 1 {
		return false, 0, fmt.Errorf("msgVal has an incorrect length of %d", len(msgVal))
	}

	// This doesn't indicate an error; it just didn't match.
	if msgVal[0] != msgbuf[0] {
		return false, 0, nil
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

		if !f.IsValid() {
			return false, 0, fmt.Errorf("field %s is not valid", fieldName)
		}

		if !f.CanSet() {
			return false, 0, fmt.Errorf("could not set field %s", fieldName)
		}

		switch f.Kind() {
		case reflect.Uint16:
			if len(msgbuf[i:]) < 2 {
				return false, 0, fmt.Errorf("invalid length for uint16")
			}
			n := binary.LittleEndian.Uint16(msgbuf[i : i+2])

			// Not sure if this is really valid for uint16
			if f.OverflowUint(uint64(n)) {
				return false, 0, fmt.Errorf("provided value, %d, would overflow if assigned to struct type", n)
			}

			f.SetUint(uint64(n))

		case reflect.Uint32:
			if len(msgbuf[i:]) < 4 {
				return false, 0, fmt.Errorf("invalid length for uint32")
			}
			n := binary.LittleEndian.Uint32(msgbuf[i : i+4])

			// Not sure if this is really valid for uint16
			if f.OverflowUint(uint64(n)) {
				return false, 0, fmt.Errorf("provided value, %d, would overflow if assigned to struct type", n)
			}

			f.SetUint(uint64(n))

		case reflect.String:
			hasString = true
			f.SetString(string(msgbuf[i:]))

		default:
			return false, 0, errors.New("Unsupported type used in struct")
		}

		i += int(fieldSize)
	}

	// Can't use the reflected size of the struct because of word alignment padding
	if !hasString && uint64(totalFieldSize) != uint64(len(msgbuf)) {
		return false, 0, fmt.Errorf("struct size %d did not match len of msgbuf %d", totalFieldSize, len(msgbuf))
	}

	return true, seqNumber, nil
}
