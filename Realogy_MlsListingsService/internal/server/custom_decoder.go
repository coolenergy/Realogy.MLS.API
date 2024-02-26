package server

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"reflect"
	"strconv"
)

type customDecoder struct {
	defDecoder bsoncodec.ValueDecoder
	zeroValue  reflect.Value
}

func (d *customDecoder) DecodeValue(dctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {

	if vr.Type() != bsontype.Null {
		if vr.Type() == bsontype.String && val.Kind() == reflect.Bool {
			readString, _ := vr.ReadString()
			if val.CanSet() {
				if pb, err := strconv.ParseBool(readString); err == nil {
					val.SetBool(pb)
				} else {
					log.Errorf("error while decoding string to boolean : %v", err)
				}
			}
			return nil
		}
		err := d.defDecoder.DecodeValue(dctx, vr, val)
		if err != nil {
			log.Errorf("error while decoding : %v", err)
			vr.Skip()
		}
		return nil
	}

	if !val.CanSet() {
		return errors.New("unable to set value")
	}

	if err := vr.ReadNull(); err != nil {
		return err
	}
	val.Set(d.zeroValue)
	return nil
}