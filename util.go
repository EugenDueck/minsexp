package minsexp

// will only forward errors from cb; will not generate any errors of its own
func TraverseLists(exprI interface{}, cb func([]interface{}) error) error {
	switch expr := exprI.(type) {
	case []interface{}:
		e := cb(expr)
		if e != nil {
			return e
		}
		for _, v := range expr {
			e := TraverseLists(v, cb)
			if e != nil {
				return e
			}
		}
	}
	return nil
}
