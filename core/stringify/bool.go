package stringify

func Bool(value bool) string {
	if value {
		return "true"
	} else {
		return "false"
	}
}
