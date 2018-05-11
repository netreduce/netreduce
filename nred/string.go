package nred

func unescapeString(s string) string {
	var (
		us      []byte
		escaped bool
	)

	for _, b := range []byte(s) {
		if escaped {
			switch b {
			case 'b':
				us = append(us, '\b')
			case 'f':
				us = append(us, '\f')
			case 'n':
				us = append(us, '\n')
			case 'r':
				us = append(us, '\r')
			case 't':
				us = append(us, '\t')
			case 'v':
				us = append(us, '\v')
			default:
				us = append(us, b)
			}

			escaped = false
			continue
		}

		if b == '\\' {
			escaped = true
			continue
		}

		us = append(us, b)
	}

	return string(us)
}
