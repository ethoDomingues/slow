package slow

import (
	"reflect"

	"github.com/ethoDomingues/c3po"
)

func MountSchemaFromRequest(f *c3po.Fielder, req *Request) (reflect.Value, any) {

	var errs any
	var sch reflect.Value
	var v any
	in, ok := f.Tags["in"]
	if ok {
		switch in {
		default:
			v = req.Form[f.Name]
		case "files":
			v = req.Files[f.Name]
		case "headers":
			v = req.Header.Get(f.Name)
		case "query":
			v = req.Query.Get(f.Name)
		}
	} else {
		v = req.Form
	}

	switch f.Type {
	default:
		_sch := reflect.TypeOf(f.Schema)
		if _sch.Kind() == reflect.Ptr {
			_sch = _sch.Elem()
		}
		if v == nil || v == "" {
			if f.Required {
				return reflect.Value{}, c3po.RetMissing(f)
			}
			sch = reflect.New(_sch).Elem()
			break
		}

		if _v, ok := v.(map[string]any); ok {
			if _, ok := _v[f.Name]; !ok {
				if f.Required {
					return reflect.Value{}, c3po.RetMissing(f)
				}
				sch = reflect.New(_sch).Elem()
				break
			}
			v = _v[f.Name]
		}

		sch = reflect.New(_sch).Elem()
		schV := reflect.ValueOf(v)
		if !c3po.SetReflectValue(sch, &schV, f.Escape) {
			return reflect.Value{}, c3po.RetInvalidType(f)
		}
	case reflect.Array, reflect.Slice:
		if _v, ok := v.(map[string]any); ok {
			if _, ok := _v[f.Name]; !ok {
				if f.Required {
					return reflect.Value{}, c3po.RetMissing(f)
				}
				sch = reflect.MakeSlice(
					reflect.SliceOf(
						reflect.TypeOf(
							f.SliceType.Schema)), 0, 0)
				break
			}
			v = _v[f.Name]
		}
		schVal := reflect.ValueOf(v)
		if schVal.Kind() == reflect.Ptr {
			schVal = schVal.Elem()
		}
		if schVal.Kind() != reflect.Slice {
			errs = c3po.RetInvalidType(f)
			break
		}

		sliceOf := reflect.TypeOf(f.SliceType.Schema)
		lenSlice := schVal.Len()
		sch = reflect.MakeSlice(reflect.SliceOf(sliceOf), lenSlice, lenSlice)
		_errs := []any{}
		for i := 0; i < lenSlice; i++ {
			s := schVal.Index(i)
			sf := f.SliceType

			slicSch, err := sf.MountSchema(s.Interface())
			if err != nil {
				_errs = append(_errs, err)
				if f.SliceStrict {
					break
				}
			}
			sItem := sch.Index(i)
			sItem.Set(slicSch)
		}
		if len(_errs) > 0 {
			if len(_errs) == 1 {
				errs = _errs[0]
			} else {
				errs = _errs
			}
		}
	case reflect.Struct:
		_errs := []any{}
		_sch := reflect.TypeOf(f.Schema)
		if _sch.Kind() == reflect.Ptr {
			_sch = _sch.Elem()
		}
		sch = reflect.New(_sch).Elem()

		for i := 0; i < sch.NumField(); i++ {
			fieldName := f.FieldsByIndex[i]
			fielder := f.Children[fieldName]
			schF := sch.FieldByName(fieldName)

			rv, __errs := MountSchemaFromRequest(fielder, req)
			if __errs != nil {
				_errs = append(_errs, __errs)
				continue
			}

			if !c3po.SetReflectValue(schF, &rv, false) {
				_errs = append(_errs, map[string]any{fielder.Name: c3po.RetInvalidType(fielder)})
				continue
			}
		}
		if len(_errs) > 0 {
			if len(_errs) == 1 {
				errs = _errs[0]
			} else {
				errs = _errs
			}
		}
	}

	if errs != nil {
		if f.Name != "" {
			return sch, map[string]any{f.Name: errs}
		}
		if slcErr, ok := errs.([]any); ok && len(slcErr) == 1 {
			return sch, slcErr[0]
		} else {
			if mapErrs, ok := errs.(map[string]any); ok && len(mapErrs) == 1 {
				for _, err := range mapErrs {
					return sch, err
				}
			}
			return sch, errs
		}

	}
	return sch, nil
}
