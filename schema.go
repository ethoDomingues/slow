package slow

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethoDomingues/c3po"
)

func MountSchemaFromRequest(f *c3po.Fielder, rq *Request) (any, error) {
	var sch any
	var err error
	var errs = []error{}
	switch f.Type {
	default:
		v, isFile := getData(f, rq)
		schT := reflect.TypeOf(f.Schema)
		if schT.Kind() == reflect.Ptr {
			schT = schT.Elem()
		}
		if v == nil || v == "" {
			if f.Required {
				return nil, c3po.RetMissing(f)
			}
			sch = reflect.New(schT).Elem()
			break
		}
		if isFile {
			if f.IsSlice {
				return v, nil
			}
			return v.([]*File)[0], nil
		}
		_sch := reflect.New(schT).Elem()
		schV := reflect.ValueOf(v)
		if !c3po.SetReflectValue(_sch, schV, f.Escape) {
			return nil, c3po.RetInvalidType(f)
		}
		return f.CheckSchPtr(_sch), nil
	case reflect.Slice:
		v, isFile := getData(f, rq)
		if v == nil {
			sch = reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(f.SliceType.Schema)), 0, 0)
			if f.Required {
				return nil, c3po.RetMissing(f)
			}
			return sch.(reflect.Value).Interface(), nil
		}
		if isFile {
			return v, nil
		} else {
			sch, err = f.Decode(v)
		}
		if err != nil {
			errs = append(errs, err)
		}

	case reflect.Struct:
		rt := reflect.TypeOf(f.Schema).Elem()
		_sch := reflect.New(rt).Elem()
		for i := 0; i < rt.NumField(); i++ {
			fName := f.FieldsByIndex[i]
			fielder := f.Children[fName]
			rtField := _sch.Field(i)

			v, isFile := getData(fielder, rq)
			if v == nil {
				if fielder.Required {
					errs = append(errs, c3po.RetMissing(fielder))
				}
				continue
			}
			if isFile {
				if v != nil && !fielder.IsSlice {
					_v, ok := v.([]*File)
					if ok && len(_v) > 0 {
						v = _v[0]
					} else {
						if fielder.Required {
							errs = append(errs, c3po.RetMissing(fielder))
						}
						continue
					}
				}
				if !c3po.SetReflectValue(rtField, reflect.ValueOf(v), false) {
					errs = append(errs, c3po.RetInvalidType(fielder))
				}
			} else {
				schF, e := fielder.Decode(v)

				if e != nil {
					errs = append(errs, e)
				} else {
					if !c3po.SetReflectValue(rtField, reflect.ValueOf(schF), fielder.Escape) {
						errs = append(errs, c3po.RetInvalidType(fielder))
					}
				}
			}
		}
		if len(errs) == 0 {
			sch = f.CheckSchPtr(_sch)
		}
	}
	if len(errs) > 0 {
		if f.Name != "" {
			return sch, fmt.Errorf(`{"%s":%v}`, f.Name, formatErr(errs...))
		}
		return nil, formatErr(errs...)
	}
	return sch, nil
}

func formatErr(errs ...error) error {
	if len(errs) > 1 {
		errBuf := bytes.NewBufferString("[")
		for i, e := range errs {
			if i > 0 {
				errBuf.WriteString(",")
			}
			errBuf.WriteString(e.Error())
		}
		errBuf.WriteString("]")
		return errors.New(errBuf.String())
	}
	if len(errs) == 1 {
		e := errs[0]
		if e != nil {
			return e
		}
	}
	return nil
}

func getData(f *c3po.Fielder, rq *Request) (any, bool) {
	fName := f.Name
	if fName == "" {
		fName = f.RealName
	}
	if fName == "" {
		return nil, false
	}
	isFile := false
	if _, ok := f.Schema.([]*File); ok {
		isFile = true
	}
	if _, ok := f.Schema.(*File); ok {
		isFile = true
	}

	switch strings.ToLower(f.Tags["in"]) {
	default:
		if isFile {
			return rq.Files[f.Name], true
		}
		return rq.Form[f.Name], false
	case "files":
		return rq.Files[f.Name], true
	case "headers":
		return rq.Header.Get(f.Name), false
	case "query":
		return rq.Query.Get(f.Name), false
	case "auth":
		if u, p, ok := rq.BasicAuth(); ok {
			if f.Name == "user" {
				return u, false
			} else if f.Name == "password" {
				return p, false
			}
		}

	}
	return nil, false
}
