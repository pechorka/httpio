package httpio

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
)

const defaultDelimiter = "."

type PathLookuperFunc func(r *http.Request, name string) (string, bool)

type Unmarshaler[T any] struct {
	c            *compiledType
	pathLookuper PathLookuperFunc
}

type UnmarshalerOptions struct {
	// PathLookuper to get path values
	PathLookuper PathLookuperFunc
	Delimiter    string
}

type UnmarshalerOption func(o *UnmarshalerOptions)

func WithPathLookuper(lookuper PathLookuperFunc) UnmarshalerOption {
	return func(o *UnmarshalerOptions) {
		o.PathLookuper = lookuper
	}
}

func WithDelimiter(delimiter string) UnmarshalerOption {
	return func(o *UnmarshalerOptions) {
		o.Delimiter = delimiter
	}
}

func NewUnmarshaler[T any](userOpts ...UnmarshalerOption) (*Unmarshaler[T], error) {
	opts := &UnmarshalerOptions{
		PathLookuper: defaultPathLookuper,
		Delimiter:    defaultDelimiter,
	}
	for _, opt := range userOpts {
		opt(opts)
	}
	compiledType, err := compileType[T](opts.Delimiter)
	if err != nil {
		return nil, fmt.Errorf("failed to compile type: %w", err)
	}
	return &Unmarshaler[T]{
		c:            compiledType,
		pathLookuper: opts.PathLookuper,
	}, nil
}

func defaultPathLookuper(r *http.Request, name string) (string, bool) {
	v := r.PathValue(name)
	return v, len(v) > 0
}

type tagType int

const (
	tagTypeNone tagType = iota
	tagTypeQuery
	tagTypePath
	tagTypeHeader
	tagTypeCookie
)

type valueSetterFunc func(v reflect.Value, vals []string) error

type compiledField struct {
	idx         []int
	set         valueSetterFunc
	isPtr       bool
	structField string // structName.fieldName for error messages
}

type compiledType struct {
	queryFields  map[string]compiledField
	pathFields   map[string]compiledField
	headerFields map[string]compiledField
	cookieFields map[string]compiledField
}

var compiledTypeCache = &sync.Map{}

func compileType[T any](delimiter string) (*compiledType, error) {
	t := reflect.TypeFor[T]()
	if cached, ok := compiledTypeCache.Load(t); ok {
		return cached.(*compiledType), nil
	}

	// only accept structs
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %s is not a struct", t.Name())
	}

	c := &compiledType{
		queryFields:  map[string]compiledField{},
		pathFields:   map[string]compiledField{},
		headerFields: map[string]compiledField{},
		cookieFields: map[string]compiledField{},	
	}
	walkType(t, nil, nil, delimiter, c)

	compiledTypeCache.Store(t, c)

	return c, nil
}

func walkType(
	t reflect.Type,
	pathPrefix []string,
	idxPrefix []int,
	delimiter string,
	out *compiledType,
) {
	for i := range t.NumField() {
		sf := t.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}

		name, src, ok := findTag(sf)
		if !ok {
			name = sf.Name
			src = tagTypeQuery
		}

		path := append(slices.Clone(pathPrefix), name)
		idx := append(slices.Clone(idxPrefix), sf.Index...)

		under := sf.Type
		isPtr := under.Kind() == reflect.Ptr
		if isPtr {
			under = under.Elem()
		}

		if isStructExpandable(under) {
			walkType(under, path, idx, delimiter, out)
			continue
		}

		cf := compiledField{
			idx:         idx,
			set:         makeValueSetter(sf.Type),
			isPtr:       isPtr,
			structField: fmt.Sprintf("%s.%s", t.Name(), sf.Name),
		}

		fullName := strings.Join(path, delimiter)
		switch src {
		case tagTypeQuery:
			out.queryFields[fullName] = cf
		case tagTypePath:
			out.pathFields[fullName] = cf
		case tagTypeHeader:
			headerName := http.CanonicalHeaderKey(fullName)
			out.headerFields[headerName] = cf
		case tagTypeCookie:
			out.cookieFields[fullName] = cf
		}
	}
}

func findTag(t reflect.StructField) (string, tagType, bool) {
	// Check for direct tag names: query, path, header, cookie
	if tag, ok := t.Tag.Lookup("query"); ok && tag != "" {
		return tag, tagTypeQuery, true
	}
	if tag, ok := t.Tag.Lookup("path"); ok && tag != "" {
		return tag, tagTypePath, true
	}
	if tag, ok := t.Tag.Lookup("header"); ok && tag != "" {
		return tag, tagTypeHeader, true
	}
	if tag, ok := t.Tag.Lookup("cookie"); ok && tag != "" {
		return tag, tagTypeCookie, true
	}

	return "", 0, false
}

func isStructExpandable(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	// Treat as scalar if it (or pointer to it) implements TextUnmarshaler.
	if implementsTextUnmarshaler(t) || implementsTextUnmarshaler(reflect.PointerTo(t)) {
		return false
	}
	return true
}

func implementsTextUnmarshaler(t reflect.Type) bool {
	if !t.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
		return false
	}
	return true
}

func makeValueSetter(ft reflect.Type) valueSetterFunc {
	if ft.Kind() == reflect.Pointer {
		elemSet := makeValueSetter(ft.Elem())
		return func(v reflect.Value, vals []string) error {
			if v.IsNil() {
				v.Set(reflect.New(ft.Elem()))
			}
			return elemSet(v.Elem(), vals)
		}
	}

	// Slice of scalars
	if ft.Kind() == reflect.Slice {
		elem := ft.Elem()
		// Slice of structs is not supported unless elem implements TextUnmarshaler.
		if elem.Kind() == reflect.Struct && !implementsTextUnmarshaler(elem) && !implementsTextUnmarshaler(reflect.PointerTo(elem)) {
			return func(reflect.Value, []string) error {
				return fmt.Errorf("unsupported slice element type: %v", elem)
			}
		}

		elemSet := makeScalarSetter(elem)
		return func(v reflect.Value, vals []string) error {
			if len(vals) == 0 {
				// leave zero value slice
				return nil
			}
			s := reflect.MakeSlice(ft, len(vals), len(vals))
			for i := range vals {
				if err := elemSet(s.Index(i), vals[i]); err != nil {
					return err
				}
			}
			v.Set(s)
			return nil
		}
	}

	scalar := makeScalarSetter(ft)
	return func(v reflect.Value, vals []string) error {
		if len(vals) == 0 {
			return nil
		}
		return scalar(v, vals[0])
	}
}

func makeScalarSetter(ft reflect.Type) func(reflect.Value, string) error {
	if implementsTextUnmarshaler(ft) || implementsTextUnmarshaler(reflect.PointerTo(ft)) {
		return func(v reflect.Value, s string) error {
			// Ensure addressable pointer receiver.
			var tu encoding.TextUnmarshaler
			if v.CanAddr() {
				if x, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
					tu = x
				}
			}
			if tu == nil && v.CanInterface() {
				if x, ok := v.Interface().(encoding.TextUnmarshaler); ok {
					tu = x
				}
			}
			if tu == nil {
				return fmt.Errorf("type %v claims TextUnmarshaler but value not addressable", ft)
			}
			return tu.UnmarshalText([]byte(s))
		}
	}

	switch ft.Kind() {
	case reflect.String:
		return func(v reflect.Value, s string) error {
			v.SetString(s)
			return nil
		}
	case reflect.Bool:
		return func(v reflect.Value, s string) error {
			b, err := strconv.ParseBool(s)
			if err != nil {
				return fmt.Errorf("parse bool: %w", err)
			}
			v.SetBool(b)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bits := ft.Bits()
		return func(v reflect.Value, s string) error {
			i, err := strconv.ParseInt(s, 10, bits)
			if err != nil {
				return fmt.Errorf("parse int: %w", err)
			}
			v.SetInt(i)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		bits := ft.Bits()
		return func(v reflect.Value, s string) error {
			u, err := strconv.ParseUint(s, 10, bits)
			if err != nil {
				return fmt.Errorf("parse uint: %w", err)
			}
			v.SetUint(u)
			return nil
		}
	case reflect.Float32, reflect.Float64:
		bits := ft.Bits()
		return func(v reflect.Value, s string) error {
			f, err := strconv.ParseFloat(s, bits)
			if err != nil {
				return fmt.Errorf("parse float: %w", err)
			}
			v.SetFloat(f)
			return nil
		}
	default:
		// Named types over the above kinds work fine with Set* calls.
		return func(reflect.Value, string) error {
			return fmt.Errorf("unsupported scalar type: %v", ft)
		}
	}
}

func (u *Unmarshaler[T]) Unmarshal(r *http.Request, dst *T) error {
	if u.c == nil {
		return fmt.Errorf("Unmarshaler is not initialized")
	}

	if ct := r.Header.Get("Content-Type"); ct != "" {
		if mt, _, _ := mime.ParseMediaType(ct); mt == "application/json" {
			if err := json.NewDecoder(r.Body).Decode(dst); err != nil && !errors.Is(err, io.EOF) {
				return err
			}
		}
	}

	root := reflect.ValueOf(dst).Elem()
	err := firstError(
		unmarshalQuery(r, u.c.queryFields, root),
		unmarshalPath(r, u.c.pathFields, root, u.pathLookuper),
		unmarshalHeader(r, u.c.headerFields, root),
		unmarshalCookie(r, u.c.cookieFields, root),
	)
	if err != nil {
		return err
	}

	return nil
}

func unmarshalQuery(r *http.Request, fields map[string]compiledField, dstStruct reflect.Value) error {
	if len(fields) == 0 {
		return nil
	}	

	parsedQuery := r.URL.Query()

	for key, vals := range parsedQuery {
		cf, ok := fields[key]
		if !ok {
			continue
		}

		fieldV := dstStruct.FieldByIndex(cf.idx)
		if err := cf.set(fieldV, vals); err != nil {
			return fmt.Errorf("field %s: %w", cf.structField, err)
		}
	}

	return nil
}

func unmarshalPath(
	r *http.Request,
	fields map[string]compiledField,
	dstStruct reflect.Value,
	pathLookuper PathLookuperFunc,
) error {
	if len(fields) == 0 {
		return nil
	}	

	for key, cf := range fields {
		v, okPath := pathLookuper(r, key)
		if !okPath {
			continue
		}

		fieldV := dstStruct.FieldByIndex(cf.idx)
		if err := cf.set(fieldV, []string{v}); err != nil {
			return fmt.Errorf("field %s: %w", cf.structField, err)
		}
	}
	return nil
}

func unmarshalHeader(
	r *http.Request,
	fields map[string]compiledField,
	dstStruct reflect.Value,
) error {
	if len(fields) == 0 {
		return nil
	}	

	for key, vals := range r.Header {
		cf, ok := fields[key]
		if !ok {
			continue
		}

		fieldV := dstStruct.FieldByIndex(cf.idx)
		if err := cf.set(fieldV, vals); err != nil {
			return fmt.Errorf("field %s: %w", cf.structField, err)
		}
	}
	return nil
}

func unmarshalCookie(
	r *http.Request,
	fields map[string]compiledField,
	dstStruct reflect.Value,
) error {
	if len(fields) == 0 {
		return nil
	}

	for key, cf := range fields {
		c, err := r.Cookie(key)
		if err != nil {
			return fmt.Errorf("cookie %s is invalid: %w", key, err)
		}

		fieldV := dstStruct.FieldByIndex(cf.idx)
		if err := cf.set(fieldV, []string{c.Value}); err != nil {
			return fmt.Errorf("field %s: %w", cf.structField, err)
		}
	}

	return nil
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
